package main

import (
	"log"
	"os"
	"time"

	"github.com/korylprince/jettison/lib/rpc"
)

func GenerateReport(config *Config, fs *FileService) *rpc.Report {
	return &rpc.Report{
		Serial:       config.Serial,
		HardwareAddr: config.HardwareAddr,
		Location:     config.Location,
		Version:      fs.Versions(),
	}
}

func ReportService(config *Config, stream rpc.Events_StreamClient, fs *FileService) {
	log.Println("Report: Service Started")
	for {
		rpt := GenerateReport(config, fs)
		log.Printf("Report: Serial: %s, HardwareAddr: %s, Location: %s, Version: %v",
			rpt.Serial, rpt.HardwareAddr, rpt.Location, rpt.Version)

		err := stream.Send(rpt)
		if err != nil {
			log.Printf("Report: EXITING, Error: %v\n", err)
			os.Exit(1)
		}

		time.Sleep(config.ReportInterval * time.Second)
	}
}
