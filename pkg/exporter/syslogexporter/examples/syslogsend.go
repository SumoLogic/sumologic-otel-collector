package main

import (
	"log"

	"github.com/SumoLogic/sumologic-otel-collector/pkg/exporter/syslogexporter"
)

func main() {
	w, err := syslogexporter.Dial("udp", "127.0.0.1:514", syslogexporter.LOG_ERR, "testtag")
	if err != nil {
		log.Fatal("failed to connect to syslog:", err)
	}
	log.Println("this is a test")
	defer w.Close()
	w.Alert("this is an Alert")
}
