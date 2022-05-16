package smpp

import (
	"time"

	"github.com/mdouchement/smpp/smpp/pdu/pdufield"
	"github.com/mdouchement/smpp/smpp/pdu/pdutlv"
	"github.com/mdouchement/smsc3/pdutext"
)

// A Message configures a short message that can be submitted via the Session.
type Message struct {
	Size     int
	Segments int

	Src      string
	Dst      string
	DstList  []string // List of destination addreses for submit multi
	DLs      []string // List of destribution list for submit multi
	Text     pdutext.Codec
	Validity time.Duration
	Register pdufield.DeliverySetting

	// Other fields, normally optional.
	TLVFields            pdutlv.Fields
	ServiceType          string
	ESMClass             uint8
	ProtocolID           uint8
	PriorityFlag         uint8
	ScheduleDeliveryTime string
	ReplaceIfPresentFlag uint8
	SMDefaultMsgID       uint8
	NumberDests          uint8
}
