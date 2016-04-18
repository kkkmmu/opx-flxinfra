package server

import (
	"fmt"
	"time"
)

const (
	KA_TIMEOUT_COUNT      = 5 // After 5 KA missed from a daemon, sysd will assume the daemon as non-responsive. Restart it.
	WD_MAX_NUM_RESTARTS   = 5 // After 5 restarts, if this daemon is still not responsive then stop it.
	SYSD_TOTAL_KA_DAEMONS = 32
)

type WDInfo struct {
	Active        bool
	RecvedKACount int32
	NumRestarts   int32
}

func (server *SYSDServer) StartWDRoutine() error {
	server.KaRecvCh = make(chan string, SYSD_TOTAL_KA_DAEMONS)
	server.KaRecvMap = make(map[string]*WDInfo)
	go server.WDTimer()
	for {
		select {
		case kaDaemon := <-server.KaRecvCh:
			if server.KaRecvMap[kaDaemon] == nil {
				wdInfo := &WDInfo{}
				server.KaRecvMap[kaDaemon] = wdInfo
			}
			server.KaRecvMap[kaDaemon].RecvedKACount++
		}
	}
	return nil
}

func (server *SYSDServer) WDTimer() error {
	server.logger.Info("Starting system WD")
	wdTimer := time.NewTicker(time.Second * KA_TIMEOUT_COUNT)
	for t := range wdTimer.C {
		_ = t
		for daemon, wd := range server.KaRecvMap {
			if wd.RecvedKACount < KA_TIMEOUT_COUNT {
				server.logger.Info(fmt.Sprintln("Daemon ", daemon, " is not responsive. Received ", wd.RecvedKACount, " keepalive messages"))
			}
			server.KaRecvMap[daemon].RecvedKACount = 0
		}
	}
	return nil
}
