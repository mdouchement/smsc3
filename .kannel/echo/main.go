package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/mdouchement/smpp/smpp/pdu"
	"github.com/mdouchement/smpp/smpp/pdu/pdufield"
	"github.com/mdouchement/smpp/smpp/pdu/pdutext"
	"github.com/pkg/errors"
)

func main() {
	addr := os.Getenv("ECHO_ADDR")
	if addr == "" {
		addr = "localhost:20001"
	}

	c, err := net.Dial("tcp", addr)
	check(err)

	p := pdu.NewBindTransmitter()
	f := p.Fields()
	f.Set(pdufield.SystemID, "smpp-echo")
	err = p.SerializeTo(c)
	check(err)

	fmt.Println(p.Header().ID.String(), "started")

	for {
		p, err = pdu.Decode(c)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Session closed")
				return
			}
			fmt.Println(errors.Wrap(err, "smpp: pdu decode"))
			return
		}
		spew.Dump(p)
		buf := bytes.NewBuffer(nil)
		p.SerializeTo(buf)
		fmt.Println(buf.Bytes())

		sm := p.Fields()[pdufield.ShortMessage]
		if sm != nil && sm.Len() > 0 {
			switch uint8(p.Fields()[pdufield.DataCoding].Bytes()[0]) {
			case 0x00:
				msg := pdutext.GSM7(sm.Bytes()).Decode()
				fmt.Println(string(msg))
			case 0x08:
				msg := pdutext.UCS2(sm.Bytes()).Decode()
				fmt.Println(string(msg))
			}
		}

		err = pdu.NewGenericNACK().SerializeTo(c)
		check(err)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
