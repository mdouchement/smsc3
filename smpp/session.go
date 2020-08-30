package smpp

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"sync"
	"time"

	"github.com/fiorix/go-smpp/smpp/pdu"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutlv"
	"github.com/goburrow/cache"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/smsc3/pdutext"
)

type (
	// A Session is a SMPP session.
	Session struct {
		mu            sync.Mutex
		rnd           *rand.Rand
		c             net.Conn
		sequences     cache.Cache
		systemID      string
		international *regexp.Regexp
		national      *regexp.Regexp
	}

	// A PDU is a response to a SMPP command.
	PDU = pdu.Body
)

// NewSession returns a new Session.
func NewSession(c net.Conn, systemID string) *Session {
	return &Session{
		rnd: rand.New(rand.NewSource(time.Now().UnixNano())),
		c:   c,
		sequences: cache.New(
			cache.WithMaximumSize(4096<<20), // 4 MiB
			cache.WithExpireAfterWrite(10*time.Minute),
		),
		international: regexp.MustCompile(`^(\+|00)[0-9]+$`),
		national:      regexp.MustCompile(`^[0-9]+$`),
	}
}

// Close closes the session.
func (s *Session) Close() error {
	return s.sequences.Close()
}

// SystemID returns the system id of the SMSC.
func (s *Session) SystemID() string {
	return s.systemID
}

// AddPDU adds PDU response to the session.
func (s *Session) AddPDU(p PDU) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sequences.Put(p.Header().Seq, p)
}

// PDU returns the PDU response for the given sequence.
func (s *Session) PDU(sequence uint32) PDU {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, ok := s.sequences.GetIfPresent(sequence)
	if !ok {
		return nil
	}
	return v.(PDU)
}

// Send send the SMS to the session.
func (s *Session) Send(m *Message, p pdu.Body) error {
	send := s.single
	if m.Chunks > 1 {
		send = s.multipart
	}

	s.defaults(m, p)
	err := send(m, p)
	if err != nil {
		return err
	}

	// Wait for PDU response in order to ACK the request
	start := time.Now()
	for {
		time.Sleep(250 * time.Millisecond)
		r := s.PDU(p.Header().Seq)
		if r == nil {
			if time.Since(start) > 10*time.Second {
				return errors.New("timeout")
			}

			continue
		}

		if r.Header().Status == 0 {
			return nil
		}

		return errors.New(r.Header().Status.Error())
	}
}

func (s *Session) defaults(m *Message, p pdu.Body) {
	f := p.Fields()
	f.Set(pdufield.SourceAddr, m.Src)
	switch {
	case s.international.MatchString(m.Src):
		f.Set(pdufield.SourceAddrTON, 0x01)
	case s.national.MatchString(m.Src):
		f.Set(pdufield.SourceAddrTON, 0x02)
	default:
		f.Set(pdufield.SourceAddrTON, 0x03) // Alpha numeric
	}
	f.Set(pdufield.SourceAddrNPI, 0x00) // Unknown

	//

	f.Set(pdufield.DestinationAddr, m.Dst)
	switch {
	case s.international.MatchString(m.Dst):
		f.Set(pdufield.DestAddrTON, 0x01)
	case s.national.MatchString(m.Dst):
		f.Set(pdufield.DestAddrTON, 0x02)
	default:
		f.Set(pdufield.DestAddrTON, 0x03) // Alpha numeric
	}
	f.Set(pdufield.DestAddrNPI, 0x00) // Unknown

	//

	f.Set(pdufield.RegisteredDelivery, uint8(m.Register))
	// Check if the message has validity set.
	if m.Validity != time.Duration(0) {
		f.Set(pdufield.ValidityPeriod, convertValidity(m.Validity))
	}
	f.Set(pdufield.ServiceType, m.ServiceType)
	f.Set(pdufield.ESMClass, m.ESMClass)
	f.Set(pdufield.ProtocolID, m.ProtocolID)
	f.Set(pdufield.PriorityFlag, m.PriorityFlag)
	f.Set(pdufield.ScheduleDeliveryTime, m.ScheduleDeliveryTime)
	f.Set(pdufield.ReplaceIfPresentFlag, m.ReplaceIfPresentFlag)
	f.Set(pdufield.SMDefaultMsgID, m.SMDefaultMsgID)

	tlv := p.TLVFields()
	for f, b := range m.TLVFields {
		tlv.Set(f, b)
	}
}

func (s *Session) single(m *Message, p pdu.Body) error {
	f := p.Fields()
	f.Set(pdufield.ShortMessage, m.Text)

	return p.SerializeTo(s.c)
}

func (s *Session) multipart(m *Message, p pdu.Body) error {
	csms := s.csmsReference()
	udh := []byte{
		6,               // UDH length
		8,               // Length of CSMS identifier, CSMS 16 bit reference number
		4,               // Length of the header, excluding first two fields
		byte(csms >> 8), // CSMS reference number (MSB), must be the same for all SMS chunks
		byte(csms),      // CSMS reference number (LSB), must be the same for all SMS chunks
		byte(m.Chunks),  // Total parts
		0,               // Part number (default value)
	}
	m.ESMClass = 0x40 // The short message begins with a user data header (UDH)

	msg := m.Text.Encode()
	limit := pdutext.SizeUCS2Multipart
	if _, ok := m.Text.(pdutext.GSM7); ok {
		limit = pdutext.SizeGSM7Multipart
	}

	for i := 0; i < m.Chunks; i++ {
		udh[len(udh)-1] = byte(i + 1) // Set part number

		f := p.Fields()
		if i < m.Chunks-1 {
			f.Set(pdufield.ShortMessage, pdutext.Raw(append(udh, msg[i*limit:(i+1)*limit]...)))
		} else {
			f.Set(pdufield.ShortMessage, pdutext.Raw(append(udh, msg[i*limit:]...)))
		}
		f.Set(pdufield.DataCoding, uint8(m.Text.Type()))

		if err := p.SerializeTo(s.c); err != nil {
			return err
		}
	}

	return nil
}

// DLRs generates and sends the DLRs for the given received SMS.
// https://smpp.org/smpp-delivery-receipt.html
// https://smpp.io/dlr-receipt/
// https://github.com/pruiz/kannel/blob/master/gw/smsc/smsc_smpp.c
func (s *Session) DLRs(l logger.Logger, p pdu.Body) {
	go func() {
		rd := p.Fields()[pdufield.RegisteredDelivery]
		if rd == nil {
			return
		}

		switch rd.Bytes()[0] {
		case 0:
			// No MC Delivery Receipt requested
		case 1:
			// MC Delivery Receipt requested where final delivery outcome is delivery success or failure

			// DELIVERED (2) ; Kannel's %d the delivery report value (dlr 1)
			dlr := createDLR(p, 2)
			time.Sleep(time.Second)
			dlr.SerializeTo(s.c)
			l.Infof("DLR DELIVERED (%d)", dlr.Header().Seq)
		case 2:
			// MC Delivery Receipt requested where the final delivery outcome is delivery failure
			// This includes scenarii where the message was cancelled via the cancel_sm operation
		case 3:
			// MC Delivery Receipt requested where the final delivery outcome is success

			// DELIVERED (2) ; Kannel's %d the delivery report value (dlr 1)
			dlr := createDLR(p, 2)
			time.Sleep(time.Second)
			dlr.SerializeTo(s.c)
			l.Infof("DLR DELIVERED (%d)", dlr.Header().Seq)
		}
	}()
}

func (s *Session) csmsReference() uint16 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return uint16(s.rnd.Intn(0xFFFF))
}

func convertValidity(d time.Duration) string {
	validity := time.Now().UTC().Add(d)
	// Absolute time format YYMMDDhhmmsstnnp, see SMPP3.4 spec 7.1.1.
	return validity.Format("060102150405") + "000+"
}

// Several ways to craft a DLR:
// esm_class + message_state + receipted_message_id (SMPP3.4)
// esm_class + message_payload  (SMPP3.4)
// esm_class + short_message    (SMPP3.3)
func createDLR(p pdu.Body, state int) pdu.Body {
	src := p.Fields()
	id := src[pdufield.MessageID].String()

	dlr := pdu.NewDeliverSM()
	f := dlr.Fields()
	f.Set(pdufield.MessageID, id)

	f.Set(pdufield.SourceAddr, src[pdufield.DestinationAddr])
	f.Set(pdufield.SourceAddrTON, src[pdufield.DestAddrTON])
	f.Set(pdufield.SourceAddrNPI, src[pdufield.DestAddrNPI])

	f.Set(pdufield.DestinationAddr, src[pdufield.SourceAddr])
	f.Set(pdufield.DestAddrTON, src[pdufield.SourceAddrTON])
	f.Set(pdufield.DestAddrNPI, src[pdufield.SourceAddrNPI])

	var msg string
	date := time.Now().Format("0601021504")
	switch state {
	case 1:
		msg = "id:%s sub:001 dlvrd:000 submit date:%s done date:%s stat:ENROUTE err:000"
		msg = fmt.Sprintf(msg, id, date, date)

		// SMPP Protocol Specification v3.4
		// 5.2.12 esm_class
		f.Set(pdufield.ESMClass, 0b100000) // Temporary DLR
	case 6:
		msg = "id:%s sub:001 dlvrd:000 submit date:%s done date:%s stat:ACCEPTD err:000"
		msg = fmt.Sprintf(msg, id, date, date)

		// SMPP Protocol Specification v3.4
		// 5.2.12 esm_class
		f.Set(pdufield.ESMClass, 0b100000) // Temporary DLR
	case 2:
		msg = "id:%s sub:001 dlvrd:001 submit date:%s done date:%s stat:DELIVRD err:000"
		msg = fmt.Sprintf(msg, id, date, date)

		// SMPP Protocol Specification v3.4
		// 5.2.12 esm_class
		f.Set(pdufield.ESMClass, 0b100) // Final DLR
	default:
		panic("invalid state")
	}
	// f.Set(pdufield.ShortMessage, []byte(msg)) // Old fashion way (SMPP3.3)
	// f.Set(pdufield.DataCoding, 1) // IA5 / ASCII

	tlv := dlr.TLVFields()
	// tlv.Set(pdutlv.TagNetworkErrorCode, pdutlv.CString("")) // No error
	tlv.Set(pdutlv.TagMessageStateOption, state)
	tlv.Set(pdutlv.TagReceiptedMessageID, pdutlv.CString(id))
	tlv.Set(pdutlv.TagMessagePayload, msg)

	return dlr
}
