package smsc

import (
	"sync"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/smsc3/smpp"
)

// A SMSC is server that handle SMPP protocol.
type SMSC struct {
	mu    sync.Mutex
	log   logger.Logger
	lhttp logger.Logger
	lsmpp logger.Logger

	// SMPP
	SMPPaddr string
	SystemID string
	Username string
	Password string
	sessions map[string]*smpp.Session

	// HTTP
	HTTPaddr string
}

// Initialize returns the given SMSC initialized.
func Initialize(l logger.Logger, smsc *SMSC) *SMSC {
	smsc.log = l
	smsc.lhttp = l.WithPrefix("[HTTP]")
	smsc.lsmpp = l.WithPrefix("[SMPP]")
	smsc.sessions = make(map[string]*smpp.Session, 1)

	if smsc.SystemID == "" {
		smsc.SystemID = "smsc3"
	}

	return smsc
}

// Listen starts SMSC listening addresses.
func (smsc *SMSC) Listen() error {
	err := make(chan error)

	go func() {
		err <- smsc.smpp()
	}()
	go func() {
		err <- smsc.http()
	}()

	return <-err
}

// Register registers a session.
func (smsc *SMSC) Register(name string, s *smpp.Session) {
	smsc.mu.Lock()
	defer smsc.mu.Unlock()

	smsc.sessions[name] = s
}

// Session returns a session.
func (smsc *SMSC) Session(name string) *smpp.Session {
	smsc.mu.Lock()
	defer smsc.mu.Unlock()

	return smsc.sessions[name]
}

// Unregister unregisters a session.
func (smsc *SMSC) Unregister(name string) {
	smsc.mu.Lock()
	defer smsc.mu.Unlock()

	delete(smsc.sessions, name)
}

// Stop gracefully stop the server.
func (smsc *SMSC) Stop() {
	smsc.log.Info("Gracefully stopping...")

	// smsc.mu.Lock()
	// defer smsc.mu.Unlock()

	for name, session := range smsc.sessions {
		if err := session.Close(); err != nil {
			smsc.log.WithError(err).Errorf("Could not close the session %s", name)
		}
	}
}
