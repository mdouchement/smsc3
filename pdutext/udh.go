package pdutext

import (
	"encoding/binary"
	"errors"
)

// UDH is the User Define Header which tells a segment details.
type UDH struct {
	Bytes    int
	ID       int
	Segments int
	Segment  int
}

// ParseUDH parses the given bytes into an UDH.
func ParseUDH(p []byte) (UDH, error) {
	udh := UDH{}

	if len(p) < 3 {
		return udh, errors.New("invalid UDH length")
	}

	l1 := int(p[0]) // UDH length
	if len(p) < l1+1 {
		return udh, errors.New("invalid UDH length")
	}
	udh.Bytes = l1 + 1

	iei := int(p[1])
	switch iei {
	case 0x00: // Concatened short message, 8-bit reference number
		udh.ID = int(p[3])
	case 0x08: // Concatened short message, 16-bit reference number
		udh.ID = int(binary.BigEndian.Uint16(p[3:]))
	}

	udh.Segments = int(p[l1-1])
	udh.Segment = int(p[l1])

	if udh.Segment > udh.Segments {
		return udh, errors.New("invlid UDH segment value or segments value")
	}

	return udh, nil
}
