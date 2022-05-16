package smpp

import (
	"net"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/smpp/smpp/pdu"
)

// A Connection embedds a net.Conn decorated with a dumper.
type Connection struct {
	net.Conn
	log logger.Logger
}

// NewConnection return a new Connection
func NewConnection(l logger.Logger, c net.Conn) *Connection {
	return &Connection{
		Conn: c,
		log:  l,
	}
}

// Decode reads the connection to decode a PDU.
func (c *Connection) Decode() (pdu.Body, error) {
	p, err := pdu.Decode(c.Conn)
	if err == nil {
		Dump(c.log, p)
	}
	return p, err
}

// Serialize writes the given PDU on the connection.
func (c *Connection) Serialize(p pdu.Body) error {
	Dump(c.log, p)
	return p.SerializeTo(c)
}
