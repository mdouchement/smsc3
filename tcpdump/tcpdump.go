package tcpdump

import (
	"io"
	"net"
	"os"
)

type dumper struct {
	net.Conn
	r io.Reader
	w io.Writer
}

// Dump dumps all IOs on a net.Conn.
func Dump(c net.Conn) net.Conn {
	return &dumper{
		Conn: c,
		r:    io.TeeReader(c, os.Stdout),
		w:    io.MultiWriter(c, os.Stdout),
	}
}

func (d *dumper) Read(p []byte) (n int, err error) {
	return d.r.Read(p)
}

func (d *dumper) Write(p []byte) (n int, err error) {
	return d.w.Write(p)
}

func (d *dumper) Close() error {
	return d.Conn.Close()
}
