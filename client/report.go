package main

import (
	"log"
	"os"
	"time"

	"github.com/korylprince/jettison/lib/rpc"
)

//GenerateReport generates an *rpc.Report from the given Config and FileService
func GenerateReport(config *Config, fs *FileService) *rpc.Report {
	return &rpc.Report{
		HardwareAddr: config.HardwareAddr,
		Location:     config.Location,
		Version:      fs.Versions(),
	}
}

//ReportService is a GRPC ReportService
func ReportService(config *Config, stream rpc.Events_StreamClient, fs *FileService) {
	log.Println("Report: Service Started")
	for {
		rpt := GenerateReport(config, fs)
		log.Printf("Report: HardwareAddr: %s, Location: %s, Version: %v",
			rpt.GetHardwareAddr(), rpt.GetLocation(), rpt.GetVersion())

		err := stream.Send(rpt)
		if err != nil {
			log.Printf("Report: EXITING, Error: %v\n", err)
			os.Exit(1)
		}

		time.Sleep(config.ReportInterval * time.Second)
	}
}
