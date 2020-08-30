package smsc

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fiorix/go-smpp/smpp/pdu"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutlv"
	"github.com/mdouchement/basex"
	"github.com/mdouchement/smsc3/pdutext"
	"github.com/mdouchement/smsc3/smpp"
	"github.com/pkg/errors"
)

type (
	// An SMSParams is used to send an SMS through HTTP.
	SMSParams struct {
		Session string `json:"session"`
		From    string `json:"from"`
		To      string `json:"to"`
		Message string `json:"message"`
	}

	// An SMSRender is used to render the result of a sent SMS through HTTP.
	SMSRender struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	}
)

func (smsc *SMSC) http() error {
	http.HandleFunc("/deliver", func(w http.ResponseWriter, r *http.Request) {
		smsc.log.Info("Got a SMS to deliver")

		var params SMSParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			smsc.render(w, http.StatusInternalServerError, err.Error())
			return
		}

		{
			if params.Session == "" {
				smsc.render(w, http.StatusBadRequest, "missing session name")
				return
			}
			if params.From == "" {
				smsc.render(w, http.StatusBadRequest, "missing from")
				return
			}
			if params.To == "" {
				smsc.render(w, http.StatusBadRequest, "missing to")
				return
			}
			if params.Message == "" {
				smsc.render(w, http.StatusBadRequest, "missing message")
				return
			}
		}

		session := smsc.Session(params.Session)
		if session == nil {
			smsc.render(w, http.StatusBadRequest, "session not found")
			return
		}

		id := basex.GenerateID()

		m := &smpp.Message{
			Src:      params.From,
			Dst:      params.To,
			Register: pdufield.FinalDeliveryReceipt,
			TLVFields: pdutlv.Fields{
				pdutlv.TagReceiptedMessageID: pdutlv.CString(id),
			},
			// Optional
			ServiceType: session.SystemID(),
		}
		m.Text, m.Size, m.Chunks = pdutext.SelectCodec(params.Message)

		p := pdu.NewDeliverSM()
		smsc.log.Infof("NewDeliverSM: %d", p.Header().Seq)
		if err := session.Send(m, p); err != nil {
			smsc.render(w, http.StatusInternalServerError, err.Error())
			return
		}

		smsc.render(w, http.StatusOK, fmt.Sprintf("OK %s (%d)", id, p.Header().Seq))
	})

	smsc.log.Infof("Listening HTTP on %s", smsc.HTTPaddr)
	return http.ListenAndServe(smsc.HTTPaddr, nil)
}

func (smsc *SMSC) render(w http.ResponseWriter, code int, message string) {
	smsc.log.Infof("[%d] %s", code, message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(&SMSRender{
		Status:  code,
		Message: message,
	})
	if err != nil {
		smsc.log.Error(errors.Wrap(err, "http: render"))
	}
}
