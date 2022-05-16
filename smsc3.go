package main

import (
	"os"
	"os/signal"
	"regexp"

	"github.com/mdouchement/logger"
	"github.com/mdouchement/smsc3/smsc"
	"github.com/sirupsen/logrus"
)

// https://gist.github.com/adothompson/1737188

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

	s := &smsc.SMSC{
		SMPPaddr: os.Getenv("SMSC3_SMPP_ADDR"),
		Username: os.Getenv("SMSC3_USERNAME"),
		Password: os.Getenv("SMSC3_PASSWORD"),
		HTTPaddr: os.Getenv("SMSC3_HTTP_ADDR"),
	}

	if s.SMPPaddr == "" {
		s.SMPPaddr = ":20001"
	}
	if s.HTTPaddr == "" {
		s.HTTPaddr = ":6000"
	}

	smsc.Initialize(logger.WrapLogrus(l), s)

	go func() {
		if err := s.Listen(); err != nil {
			l.Fatal(err)
		}
	}()
	defer s.Stop()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	<-signals
}
