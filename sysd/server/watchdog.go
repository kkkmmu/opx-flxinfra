package server

import (
	"fmt"
	"time"
)

const (
	KA_TIMEOUT_COUNT = 5 // After 5 KA missed from a daemon, sysd will assume the daemon as non-responsive
)

func (server *SYSDServer) StartWDRoutine() error {
	server.KaRecvCh = make(chan string, SYSD_TOTAL_KA_DAEMONS)
	server.KaRecvMap = make(map[string]int32)
	go server.WDTimer()
	for {
		select {
		case kaDaemon := <-server.KaRecvCh:
			server.KaRecvMap[kaDaemon]++
		}
	}
	return nil
}

func (server *SYSDServer) WDTimer() error {
	server.logger.Info("Starting system WD")
	wdTimer := time.NewTicker(time.Second * KA_TIMEOUT_COUNT)
	for t := range wdTimer.C {
		_ = t
		for daemon, kaCount := range server.KaRecvMap {
			if kaCount > 0 {
				server.logger.Info(fmt.Sprintln("Daemon ", daemon, " is responsive"))
			} else {
				server.logger.Info(fmt.Sprintln("Daemon ", daemon, " is not responsive"))
			}
			server.KaRecvMap[daemon] = 0
		}
	}
	return nil
}
