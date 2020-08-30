package smsc

import (
	"io"
	"net"

	"github.com/davecgh/go-spew/spew"
	"github.com/fiorix/go-smpp/smpp/pdu"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutlv"
	"github.com/mdouchement/basex"
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
	smsc.log.Infof("Listening SMPP %s", smsc.SMPPaddr)

	for {
		c, err := l.Accept()
		if err != nil {
			smsc.log.Error(errors.Wrap(err, "smpp: accept"))
			continue
		}

		go func() {
			defer c.Close()
			c.(*net.TCPConn).SetKeepAlive(true)

			var session *smpp.Session
			var sname string
			var r pdu.Body
			for {
				p, err := pdu.Decode(c)
				if err != nil {
					if err == io.EOF {
						smsc.log.Info("Session closed")
						return
					}
					smsc.log.Error(errors.Wrap(err, "smpp: pdu decode"))
					continue // Ignoring error
				}
				smsc.log.Infof("Got a PDU: (%d) %s %s", p.Header().Seq, p.Header().ID, p.Header().Status.Error())

				// Session connection
				if sname == "" {
					sname, r, err = smsc.auth(p)
					if err != nil {
						smsc.log.Error(errors.Wrap(err, "smpp: authentication"))
						return
					}

					r.Header().Seq = p.Header().Seq // Response must have the same sequence number
					if err = r.SerializeTo(c); err != nil {
						smsc.log.Error(errors.Wrap(err, "smpp: authentication"))
						return
					}

					session = smpp.NewSession(c, smsc.SystemID)
					defer session.Close()

					smsc.Register(sname, session)
					defer smsc.Unregister(sname)

					smsc.log.Infof("Session %s opened", sname)
					continue
				}

				// Supported SMPP commands
				var r pdu.Body
				switch p.Header().ID {
				case pdu.EnquireLinkID:
					// Ping / Heartbeat
					r = pdu.NewEnquireLinkResp()
				case pdu.DeliverSMRespID:
					// Ack of a sent SMS/DLR from SMSC to ESME
					smsc.log.Infof("ACK sms/dlr")
					session.AddPDU(p)
				case pdu.SubmitSMID:
					// Receiving SMS from ESME to SMSC
					smsc.dump(p)

					id := basex.GenerateID()
					p.Fields().Set(pdufield.MessageID, id)
					session.DLRs(smsc.log, p)

					r = pdu.NewSubmitSMResp()
					r.Fields().Set(pdufield.MessageID, id)
				case pdu.UnbindID:
					// End of session
					smsc.log.Infof("Unbind session %s", sname)
					r = pdu.NewUnbindResp()
				case pdu.GenericNACKID:
					smsc.log.Warn(p.Header().Status.Error())
				default:
					spew.Dump(p)
				}

				if r != nil {
					r.Header().Seq = p.Header().Seq // Response must have the same sequence number
					if err = r.SerializeTo(c); err != nil {
						smsc.log.Errorf("smpp: %s: %s", p.Header().ID.String(), err)
						return
					}
				}
			}
		}()
	}
}

func (smsc *SMSC) auth(p pdu.Body) (string, pdu.Body, error) {
	var r pdu.Body
	switch p.Header().ID {
	case pdu.BindTransmitterID:
		r = pdu.NewBindTransmitterResp()
	case pdu.BindReceiverID:
		r = pdu.NewBindReceiverResp()
	case pdu.BindTransceiverID:
		r = pdu.NewBindTransceiverResp()
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

func (smsc *SMSC) dump(p pdu.Body) {
	l := smsc.log.
		WithField("seq", p.Header().Seq)

	for k, v := range p.Fields() {
		l = l.WithField(string(k), v.String())
	}

	l.Info(p.Header().ID.String())
}
