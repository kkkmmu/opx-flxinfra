package server

import (
	"encoding/json"
	"fmt"
	"infra/sysd/sysdCommonDefs"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	KA_TIMEOUT_COUNT_MIN  = 0
	KA_TIMEOUT_COUNT      = 5 // After 5 KA missed from a daemon, sysd will assume the daemon as non-responsive. Restart it.
	WD_MAX_NUM_RESTARTS   = 5 // After 5 restarts, if this daemon is still not responsive then stop it.
	SYSD_TOTAL_KA_DAEMONS = 32
)

const (
	REASON_NONE           = "None"
	REASON_KA_FAIL        = "Failed to receive keepalive messages"
	REASON_USER_RESTART   = "Restarted by user"
	REASON_DAEMON_STOPPED = "Stopped by user"
)

type DaemonInfo struct {
	State         sysdCommonDefs.SRDaemonStatus
	Reason        string
	RecvedKACount int32
	NumRestarts   int32
	RestartTime   string
	RestartReason string
}

func (server *SYSDServer) StartWDRoutine() error {
	server.KaRecvCh = make(chan string, sysdCommonDefs.SYSD_TOTAL_KA_DAEMONS)
	server.DaemonMap = make(map[string]*DaemonInfo)
	go server.WDTimer()
	for {
		select {
		case kaDaemon := <-server.KaRecvCh:
			daemonInfo, exist := server.DaemonMap[kaDaemon]
			if !exist {
				daemonInfo = &DaemonInfo{}
				//server.KaRecvMap[kaDaemon] = wdInfo
			}
			daemonInfo.RecvedKACount++
			if daemonInfo.State != sysdCommonDefs.KA_UP {
				daemonInfo.State = sysdCommonDefs.KA_UP
				server.PublishDaemonKANotification(kaDaemon, sysdCommonDefs.KA_UP)
			}
		case daemonConfig := <-server.DaemonConfigCh:
			server.logger.Info(fmt.Sprintln("Received daemon config for: ", daemonConfig.Name, " to ", daemonConfig.State))
			daemon := daemonConfig.Name
			state := daemonConfig.State
			daemonInfo, exist := server.DaemonMap[daemon]
			if state == "start" {
				if !exist {
					daemonInfo = &DaemonInfo{}
					//server.KaRecvMap[kaDaemon] = wdInfo
				}
				server.ToggleFlexswitchDaemon(daemon, true)
				daemonInfo.State = sysdCommonDefs.KA_UP
				daemonInfo.Reason = REASON_NONE
			} else if state == "stop" {
				if exist {
					server.ToggleFlexswitchDaemon(daemon, false)
					daemonInfo.State = sysdCommonDefs.KA_STOPPED
					daemonInfo.Reason = REASON_DAEMON_STOPPED
					server.PublishDaemonKANotification(daemon, sysdCommonDefs.KA_STOPPED)
				} else {
					server.logger.Info(fmt.Sprintln("Received call to stop unknown daemon ", daemon))
				}
			}

		}
	}
	return nil
}

func (server *SYSDServer) PublishDaemonKANotification(name string, status sysdCommonDefs.SRDaemonStatus) error {
	msg := sysdCommonDefs.DaemonStatus{
		Name:   name,
		Status: status,
	}
	msgBuf, err := json.Marshal(msg)
	if err != nil {
		server.logger.Err("Failed to marshal daemon status")
		return err
	}
	notification := sysdCommonDefs.Notification{
		Type:    uint8(sysdCommonDefs.KA_DAEMON),
		Payload: msgBuf,
	}
	notificationBuf, err := json.Marshal(notification)
	if err != nil {
		server.logger.Err("Failed to marshal daemon status message")
		return err
	}
	server.notificationCh <- notificationBuf
	return nil
}

func (server *SYSDServer) ToggleFlexswitchDaemon(daemon string, up bool) error {
	var (
		cmdOut []byte
		err    error
		op     string
	)
	cmdDir := strings.TrimSuffix(server.paramsDir, "params/")
	cmdName := cmdDir + "flexswitch"
	if up {
		op = "start"
	} else {
		op = "stop"
	}
	cmdArgs := []string{"-n", daemon, "-o", op, "-d", cmdDir}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		server.logger.Info(fmt.Sprintln(os.Stderr, "There was an error to ", op, " flexswitch daemon ", daemon, " : ", err))
		return err
	}
	out := string(cmdOut)
	server.logger.Info(fmt.Sprintln("Flexswitch daemon ", daemon, op, " returned ", out))

	return nil
}

func (server *SYSDServer) RestartFlexswitchDaemon(daemon string) error {
	server.ToggleFlexswitchDaemon(daemon, false)
	server.PublishDaemonKANotification(daemon, sysdCommonDefs.KA_DOWN)
	server.ToggleFlexswitchDaemon(daemon, true)
	return nil
}

func (server *SYSDServer) WDTimer() error {
	server.logger.Info("Starting system WD")
	wdTimer := time.NewTicker(time.Second * KA_TIMEOUT_COUNT)
	for t := range wdTimer.C {
		_ = t
		for daemon, daemonInfo := range server.DaemonMap {
			if daemonInfo.RecvedKACount < KA_TIMEOUT_COUNT && daemonInfo.RecvedKACount > KA_TIMEOUT_COUNT_MIN {
				server.logger.Info(fmt.Sprintln("Daemon ", daemon, " is slowing down. Monitoring it."))
			}
			if daemonInfo.RecvedKACount == KA_TIMEOUT_COUNT_MIN {
				if daemonInfo.State == sysdCommonDefs.KA_UP {
					server.logger.Info(fmt.Sprintln("Daemon ", daemon, " is not responsive. Restarting it."))
					daemonInfo.State = sysdCommonDefs.KA_DOWN
					go server.RestartFlexswitchDaemon(daemon)
					daemonInfo.NumRestarts++
					daemonInfo.RestartTime = time.Now().String()
					daemonInfo.RestartReason = REASON_KA_FAIL
				}
			}
			daemonInfo.RecvedKACount = 0
		}
	}
	return nil
}

func (server *SYSDServer) GetDaemonState(name string) *DaemonState {
	daemonState := new(DaemonState)
	daemonInfo, found := server.DaemonMap[name]
	if found {
		daemonState.Name = name
		daemonState.State = daemonInfo.State
		daemonState.Reason = daemonInfo.Reason
		daemonState.RecvedKACount = daemonInfo.RecvedKACount
		daemonState.NumRestarts = daemonInfo.NumRestarts
		daemonState.RestartTime = daemonInfo.RestartTime
		daemonState.RestartReason = daemonInfo.RestartReason
	}
	return daemonState
}
func (server *SYSDServer) GetBulkDaemonStates(idx int, cnt int) (int, int, []DaemonState) {
	var nextIdx int
	var count int
	result := make([]DaemonState, cnt)
	i := 0
	for daemon, daemonInfo := range server.DaemonMap {
		result[i].Name = daemon
		result[i].State = daemonInfo.State
		result[i].Reason = daemonInfo.Reason
		result[i].RecvedKACount = daemonInfo.RecvedKACount
		result[i].NumRestarts = daemonInfo.NumRestarts
		result[i].RestartTime = daemonInfo.RestartTime
		result[i].RestartReason = daemonInfo.RestartReason
		i++
	}
	count = i
	nextIdx = 0
	return nextIdx, count, result
}
