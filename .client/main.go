package main

import (
	"net"
	"os"
	"os/signal"
	"regexp"
	"time"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/smpp/smpp/pdu"
	"github.com/mdouchement/smpp/smpp/pdu/pdufield"
	"github.com/mdouchement/smsc3/address"
	"github.com/mdouchement/smsc3/pdutext"
	"github.com/mdouchement/smsc3/smpp"
	"github.com/sirupsen/logrus"
)

func main() {
	l := logrus.New()
	l.SetFormatter(&logger.LogrusTextFormatter{
		DisableColors:   false,
		ForceColors:     true,
		ForceFormatting: true,
		PrefixRE:        regexp.MustCompile(`^(\[.*?\])\s`),
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log := logger.WrapLogrus(l)

	//

	conn, err := net.Dial("tcp", "localhost:20001")
	if err != nil {
		panic(err)
	}
	c := smpp.NewConnection(log, conn)

	p := pdu.NewBindTransceiver()
	f := p.Fields()
	f.Set(pdufield.SystemID, "kannel-sinch")
	f.Set(pdufield.Password, "12345678")
	f.Set(pdufield.InterfaceVersion, 0x34)
	if err = c.Serialize(p); err != nil {
		panic(err)
	}

	p, err = c.Decode() // resp
	if err != nil {
		panic(err)
	}

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, os.Kill)
		<-signals

		p = pdu.NewUnbind()
		if err = c.Serialize(p); err != nil {
			panic(err)
		}
	}()

	go func() {
		for {
			p = submitsm()
			if err = c.Serialize(p); err != nil {
				panic(err)
			}

			time.Sleep(6 * time.Second)
		}
	}()

	var exit bool
	for {
		if exit {
			time.Sleep(2 * time.Second)
			c.Close()
			return
		}

		log.Info("Awaiting PDU")
		p, err = c.Decode()
		if err != nil {
			panic(err)
		}

		var r pdu.Body
		switch p.Header().ID {
		case pdu.SubmitSMRespID:
			//
		case pdu.EnquireLinkID:
			r = pdu.NewEnquireLinkRespSeq(p.Header().Seq)
		case pdu.DeliverSMID:
			r = pdu.NewDeliverSMRespSeq(p.Header().Seq)
		case pdu.UnbindID:
			// End of session
			log.Info("Unbind session")
			r = pdu.NewUnbindRespSeq(p.Header().Seq)
			exit = true
		case pdu.GenericNACKID:
			log.Warn(p.Header().Status.Error())
		default:
			r = pdu.NewGenericNACK()
			r.Header().Status = 0x00000003 // Invalid Command ID
		}

		if r != nil {
			if err = c.Serialize(r); err != nil {
				log.Errorf("smpp: %s: %s", p.Header().ID.String(), err)
				return
			}
		}
	}
}

func submitsm() pdu.Body {
	p := pdu.NewSubmitSM(nil)

	f := p.Fields()
	src := address.Parse("Kannel")
	f.Set(pdufield.SourceAddr, src.String())
	f.Set(pdufield.SourceAddrTON, src.TON())
	f.Set(pdufield.SourceAddrNPI, src.NPI())

	dst := address.Parse("+33600000001")
	f.Set(pdufield.DestinationAddr, dst.String())
	f.Set(pdufield.DestAddrTON, dst.TON())
	f.Set(pdufield.DestAddrNPI, dst.NPI())

	sm, _, _ := pdutext.SelectCodec(time.Now().String())
	f.Set(pdufield.ShortMessage, sm)

	f.Set(pdufield.RegisteredDelivery, pdufield.FinalDeliveryReceipt)
	f.Set(pdufield.ValidityPeriod, smpp.ConvertValidity(-time.Since(time.Now().AddDate(0, 0, 2))))
	f.Set(pdufield.ESMClass, 0)

	return p
}
