package smpp

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/goburrow/cache"
	"github.com/mdouchement/basex"
	"github.com/mdouchement/logger"
	"github.com/mdouchement/smpp/smpp/pdu"
	"github.com/mdouchement/smpp/smpp/pdu/pdufield"
	"github.com/mdouchement/smpp/smpp/pdu/pdutlv"
	"github.com/mdouchement/smsc3/address"
	"github.com/mdouchement/smsc3/pdutext"
	"github.com/pkg/errors"
)

// UDHI is the User Data Header Indicator used in esm_class.
const UDHI = 0b0100_0000

type (
	// A Session is a SMPP session.
	Session struct {
		mu        sync.Mutex
		rnd       *rand.Rand
		log       logger.Logger
		c         *Connection
		sequences cache.Cache
		systemID  string
	}

	// A PDU is a response to a SMPP command.
	PDU = pdu.Body
)

// NewSession returns a new Session.
func NewSession(l logger.Logger, c *Connection, systemID string) *Session {
	return &Session{
		rnd:      rand.New(rand.NewSource(time.Now().UnixNano())),
		log:      l,
		c:        c,
		systemID: systemID,
		sequences: cache.New(
			cache.WithMaximumSize(4096<<20), // 4 MiB
			cache.WithExpireAfterWrite(10*time.Minute),
		),
	}
}

func (s *Session) Listen() error {
	for {
		p, err := s.c.Decode()
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "close") {
				s.log.Info("Session closed")
				return s.Close()
			}
			s.log.Error(errors.Wrap(err, "smpp: pdu decode"))
			continue // Ignoring error
		}

		// Supported SMPP commands
		var r pdu.Body
		switch p.Header().ID {
		case pdu.EnquireLinkID:
			// Ping / Heartbeat
			r = pdu.NewEnquireLinkRespSeq(p.Header().Seq)
		case pdu.DeliverSMRespID:
			// Ack of a sent SMS/DLR from SMSC to ESME
			s.log.Infof("ACK sms/dlr")
			s.AddPDU(p)
		case pdu.SubmitSMID:
			// Receiving SMS from ESME to SMSC

			id := basex.GenerateID()
			p.Fields().Set(pdufield.MessageID, id)
			s.DLRs(p)

			r = pdu.NewSubmitSMRespSeq(p.Header().Seq)
			r.Fields().Set(pdufield.MessageID, id)
		case pdu.UnbindID:
			// End of session asked by ESME
			s.log.Infof("Unbinding session %s", s.systemID)
			r = pdu.NewUnbindRespSeq(p.Header().Seq)
		case pdu.UnbindRespID:
			// End of session asked by SMSC
			s.log.Infof("Unbinded session %s", s.systemID)
		case pdu.GenericNACKID:
			s.log.Warn(p.Header().Status.Error())
		default:
			r = pdu.NewGenericNACK()
			r.Header().Status = 0x00000003 // Invalid Command ID
		}

		if r != nil {
			if err = s.c.Serialize(r); err != nil {
				s.log.Errorf("smpp: %s: %s", p.Header().ID.String(), err)
				return nil
			}
		}

		// Stop the session
		if p.Header().ID == pdu.UnbindID {
			s.c.log.Infof("Closing session %s", s.systemID)

			err = s.c.Close()
			if err != nil {
				return err
			}

			return s.sequences.Close()
		}
	}
}

// Close closes the session.
func (s *Session) Close() error {
	s.c.log.Infof("Closing session %s", s.systemID)

	p := pdu.NewUnbind()
	err := s.c.Serialize(p)
	if err != nil {
		return err
	}

	_, err = s.c.Decode() // unbind_resp
	if err != nil {
		return err
	}

	err = s.c.Close()
	if err != nil {
		return err
	}

	return s.sequences.Close()
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
	if m.Segments > 1 {
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

	src := address.Parse(m.Src)
	f.Set(pdufield.SourceAddr, src.String())
	f.Set(pdufield.SourceAddrTON, src.TON())
	f.Set(pdufield.SourceAddrNPI, src.NPI())

	dst := address.Parse(m.Dst)
	f.Set(pdufield.DestinationAddr, dst.String())
	f.Set(pdufield.DestAddrTON, dst.TON())
	f.Set(pdufield.DestAddrNPI, dst.NPI())

	f.Set(pdufield.RegisteredDelivery, uint8(m.Register))
	// Check if the message has validity set.
	if m.Validity != time.Duration(0) {
		f.Set(pdufield.ValidityPeriod, ConvertValidity(m.Validity))
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

	return s.c.Serialize(p)
}

func (s *Session) multipart(m *Message, p pdu.Body) error {
	csms := s.csmsReference8()
	udh := []byte{
		5,                // UDH length
		0,                // Length of CSMS identifier, CSMS 8 bit reference number
		3,                // Length of the header, excluding first two fields
		byte(csms),       // CSMS reference number, must be the same for all SMS segments
		byte(m.Segments), // Total parts
		0,                // Part number (default value)
	}
	m.ESMClass |= UDHI // The short message begins with a user data header (UDH)

	msg := m.Text.Encode()
	limit := pdutext.SizeUCS2Multipart
	if _, ok := m.Text.(pdutext.GSM7); ok {
		limit = pdutext.SizeGSM7Multipart
	}

	for i := 0; i < m.Segments; i++ {
		udh[len(udh)-1] = byte(i + 1) // Set part number

		f := p.Fields()
		if i < m.Segments-1 {
			f.Set(pdufield.ShortMessage, pdutext.Raw(append(udh, msg[i*limit:(i+1)*limit]...)))
		} else {
			f.Set(pdufield.ShortMessage, pdutext.Raw(append(udh, msg[i*limit:]...)))
		}
		f.Set(pdufield.DataCoding, uint8(m.Text.Type()))
		f.Set(pdufield.ESMClass, m.ESMClass) // UDH Indicator

		if err := s.c.Serialize(p); err != nil {
			return err
		}
	}

	return nil
}

// DLRs generates and sends the DLRs for the given received SMS.
// https://smpp.org/smpp-delivery-receipt.html
// https://smpp.io/dlr-receipt/
// https://github.com/pruiz/kannel/blob/master/gw/smsc/smsc_smpp.c
func (s *Session) DLRs(p pdu.Body) {
	go func() {
		field := p.Fields()[pdufield.RegisteredDelivery]
		if field == nil {
			return
		}

		rd := field.Bytes()[0]
		rd &= 0b0000_0011 // Ignore 0bxxx1xxxx that may be provided for intermediate notification.

		switch rd {
		case 0:
			// No MC Delivery Receipt requested
		case 1:
			// MC Delivery Receipt requested where final delivery outcome is delivery success or failure

			// DELIVERED (2) ; Kannel's %d the delivery report value (dlr 1)
			dlr := createDLR(p, 2)
			time.Sleep(time.Second)
			err := s.c.Serialize(dlr)
			if err != nil {
				s.log.WithError(err).Error("Could not send DLR")
			}
			s.log.Infof("DLR DELIVERED (%d)", dlr.Header().Seq)
		case 2:
			// MC Delivery Receipt requested where the final delivery outcome is success

			// DELIVERED (2) ; Kannel's %d the delivery report value (dlr 1)
			dlr := createDLR(p, 2)
			time.Sleep(time.Second)
			err := s.c.Serialize(dlr)
			if err != nil {
				s.log.WithError(err).Error("Could not send DLR")
			}
			s.log.Infof("DLR DELIVERED (%d)", dlr.Header().Seq)
		}
	}()
}

func (s *Session) csmsReference8() uint8 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return uint8(s.rnd.Intn(0xFF))
}

func (s *Session) csmsReference16() uint16 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return uint16(s.rnd.Intn(0xFFFF))
}

func ConvertValidity(d time.Duration) string {
	validity := time.Now().UTC().Add(d)
	// Absolute time format YYMMDDhhmmsstnnp, see SMPP3.4 spec 7.1.1.
	return validity.Format("060102150405") + "000+"
}

// Several ways to craft a DLR:
// esm_class + short_message + receipted_message_id
func createDLR(p pdu.Body, state int) pdu.Body {
	src := p.Fields()
	id := src[pdufield.MessageID].String()

	dlr := pdu.NewDeliverSM()
	f := dlr.Fields()

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
	sm, _, _ := pdutext.SelectCodec(msg)
	f.Set(pdufield.ShortMessage, sm)

	tlv := dlr.TLVFields()
	tlv.Set(pdutlv.TagReceiptedMessageID, pdutlv.CString(id))

	return dlr
}
