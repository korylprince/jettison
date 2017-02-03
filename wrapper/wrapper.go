// +build windows

//jettison_wrapper.exe <service log file> <program log file> <cmd> (<args>...)
package main

import (
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/btcsuite/winsvc/svc"
)

func catch() {
	if err := recover(); err != nil {
		log.Fatalln("Panicked:", err)
	}
}

func cmd(args []string, f *os.File) *exec.Cmd {
	c := exec.Command(args[0], args[1:]...)

	stdout, err := c.StdoutPipe()
	if err != nil {
		log.Fatalln(err)
	}
	stderr, err := c.StderrPipe()
	if err != nil {
		log.Fatalln(err)
	}
	go io.Copy(f, stdout)
	go io.Copy(f, stderr)
	return c
}

type service struct{}

func waitForCmd(c *exec.Cmd, changes chan<- svc.Status, done chan<- struct{}) {
	c.Wait()
	close(done)
}

func (s *service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	defer catch()
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	f, err := os.OpenFile(os.Args[2], os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	log.Println("Logging program output to:", os.Args[2])
	p := cmd(os.Args[3:], f)
	log.Println("Executing:", os.Args[3:])
	err = p.Start()
	if err != nil {
		log.Fatalln(err)
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	log.Println("Service started")

	done := make(chan struct{})
	go waitForCmd(p, changes, done)

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				log.Println("Received Control: Interrogate")
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				log.Println("Received Control: Stop")
				changes <- svc.Status{State: svc.StopPending}
				p.Process.Kill()
				break loop
			}
		case _ = <-done:
			log.Println("Process died")
			break loop
		}
	}
	changes <- svc.Status{State: svc.Stopped}
	return
}

func main() {
	f, err := os.OpenFile(os.Args[1], os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer f.Close()
	log.Println("Logging to:", os.Args[1])
	log.SetOutput(f)

	defer catch()
	log.Println("Attempting to run service")
	err = svc.Run("jettison", &service{})
	if err != nil {
		log.Fatalln("Error running service:", err)
	}
	log.Println("Stopped")
}
