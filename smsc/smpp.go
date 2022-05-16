package smsc

import (
	"io"
	"net"

	"github.com/mdouchement/smpp/smpp/pdu"
	"github.com/mdouchement/smpp/smpp/pdu/pdufield"
	"github.com/mdouchement/smpp/smpp/pdu/pdutlv"
	"github.com/mdouchement/smsc3/smpp"
	"github.com/pkg/errors"
)

// Here we are using SMPP3.4 version.
// https://smpp.org/
// https://smpp.org/SMPP_v3_4_Issue1_2.pdf
// https://smpp.org/SMPP_v5.pdf
func (smsc *SMSC) smpp() error {
	l, err := net.Listen("tcp", smsc.SMPPaddr)
	if err != nil {
		return errors.Wrap(err, "could not listen SMPP")
	}
	smsc.lsmpp.Infof("Listening SMPP %s", smsc.SMPPaddr)

	for {
		c, err := l.Accept()
		if err != nil {
			smsc.lsmpp.Error(errors.Wrap(err, "smpp: accept"))
			continue
		}

		go func() {
			defer c.Close()
			c.(*net.TCPConn).SetKeepAlive(true)
			sc := smpp.NewConnection(smsc.lsmpp, c)

			p, err := sc.Decode()
			if err != nil {
				if err == io.EOF {
					smsc.lsmpp.Info("Session closed")
					return
				}
				smsc.lsmpp.Error(errors.Wrap(err, "smpp: pdu decode"))
				return
			}

			// Session connection
			var r pdu.Body
			sname, r, err := smsc.auth(p)
			if err != nil {
				smsc.lsmpp.Error(errors.Wrap(err, "smpp: authentication"))
				return
			}

			if err = sc.Serialize(r); err != nil {
				smsc.lsmpp.Error(errors.Wrap(err, "smpp: authentication"))
				return
			}

			session := smpp.NewSession(smsc.lsmpp, sc, sname)
			defer session.Close()

			smsc.Register(sname, session)
			defer smsc.Unregister(sname)

			smsc.lsmpp.Infof("Session %s opened", sname)

			err = session.Listen()
			if err = session.Listen(); err != nil {
				smsc.lsmpp.Error(errors.Wrap(err, "smpp: session listen"))
				return
			}
		}()
	}
}

func (smsc *SMSC) auth(p pdu.Body) (string, pdu.Body, error) {
	var r pdu.Body
	switch p.Header().ID {
	case pdu.BindTransmitterID:
		r = pdu.NewBindTransmitterRespSeq(p.Header().Seq)
	case pdu.BindReceiverID:
		r = pdu.NewBindReceiverRespSeq(p.Header().Seq)
	case pdu.BindTransceiverID:
		r = pdu.NewBindTransceiverRespSeq(p.Header().Seq)
	default:
		return "", r, errors.New("unexpected pdu, want bind")
	}

	f := p.Fields()
	user := f[pdufield.SystemID]
	password := f[pdufield.Password]

	if user == nil || password == nil {
		return "", r, errors.New("malformed pdu, missing system_id/password")
	}

	session := user.String()
	if smsc.Username != "" && user.String() != smsc.Username {
		return session, r, errors.New("invalid user")
	}

	if smsc.Password != "" && password.String() != smsc.Password {
		return session, r, errors.New("invalid passwd")
	}

	r.Fields().Set(pdufield.SystemID, smsc.SystemID)
	r.TLVFields().Set(pdutlv.TagScInterfaceVersion, 0x34) // SMPP34
	return session, r, nil
}
