package rpc

import (
	"errors"
	"fmt"
	"infra/sysd/server"
	"infra/sysd/sysdCommonDefs"
	"sysd"
	"utils/logging"
)

type SYSDHandler struct {
	server *server.SYSDServer
	logger *logging.Writer
}

func NewSYSDHandler(logger *logging.Writer, server *server.SYSDServer) *SYSDHandler {
	h := new(SYSDHandler)
	h.server = server
	h.logger = logger
	return h
}

func (h *SYSDHandler) SendGlobalLoggingConfig(gLoggingConfig *sysd.SystemLogging) bool {
	var Logging bool
	if gLoggingConfig.Logging == "on" {
		Logging = true
	} else {
		Logging = false
	}
	gConf := server.GlobalLoggingConfig{
		Enable: Logging,
	}
	h.server.GlobalLoggingConfigCh <- gConf
	return true
}

func (h *SYSDHandler) SendComponentLoggingConfig(cLoggingConfig *sysd.ComponentLogging) bool {
	cConf := server.ComponentLoggingConfig{
		Component: cLoggingConfig.Module,
		Level:     logging.ConvertLevelStrToVal(cLoggingConfig.Level),
	}
	h.server.ComponentLoggingConfigCh <- cConf
	return true
}

func (h *SYSDHandler) UpdateSystemLogging(origConf *sysd.SystemLogging, newConf *sysd.SystemLogging, attrset []bool) (bool, error) {
	h.logger.Info(fmt.Sprintln("Original global config attrs:", origConf))
	if newConf == nil {
		err := errors.New("Invalid global Configuration")
		return false, err
	}
	h.logger.Info(fmt.Sprintln("Update global config attrs:", newConf))
	return h.SendGlobalLoggingConfig(newConf), nil
}

func (h *SYSDHandler) UpdateComponentLogging(origConf *sysd.ComponentLogging, newConf *sysd.ComponentLogging, attrset []bool) (bool, error) {
	h.logger.Info(fmt.Sprintln("Original component config attrs:", origConf))
	if newConf == nil {
		err := errors.New("Invalid component Configuration")
		return false, err
	}
	h.logger.Info(fmt.Sprintln("Update component config attrs:", newConf))
	return h.SendComponentLoggingConfig(newConf), nil
}

func (h *SYSDHandler) CreateSystemLogging(gLoggingConf *sysd.SystemLogging) (bool, error) {
	h.logger.Info(fmt.Sprintln("Create global config attrs:", gLoggingConf))
	return h.SendGlobalLoggingConfig(gLoggingConf), nil
}

func (h *SYSDHandler) CreateComponentLogging(cLoggingConf *sysd.ComponentLogging) (bool, error) {
	h.logger.Info(fmt.Sprintln("Create component config attrs:", cLoggingConf))
	return h.SendComponentLoggingConfig(cLoggingConf), nil
}

func (h *SYSDHandler) DeleteSystemLogging(gLoggingConf *sysd.SystemLogging) (bool, error) {
	h.logger.Info(fmt.Sprintln("Delete global config attrs:", gLoggingConf))
	err := errors.New("Not supported")
	return false, err
}

func (h *SYSDHandler) DeleteComponentLogging(cLoggingConf *sysd.ComponentLogging) (bool, error) {
	h.logger.Info(fmt.Sprintln("Delete component config attrs:", cLoggingConf))
	err := errors.New("Not supported")
	return false, err
}

func (h *SYSDHandler) CreateIpTableAcl(ipaclConfig *sysd.IpTableAcl) (bool, error) {
	h.logger.Info("Create Ip Table rule " + ipaclConfig.Name)
	h.server.IptableAddCh <- ipaclConfig
	return true, nil
	//return (h.server.AddIpTableRule(ipaclConfig, false /* non - restart*/))
}

func (h *SYSDHandler) UpdateIpTableAcl(origConf *sysd.IpTableAcl,
	newConf *sysd.IpTableAcl, attrset []bool) (bool, error) {
	err := errors.New("Not supported")
	return false, err
}

func (h *SYSDHandler) DeleteIpTableAcl(ipaclConfig *sysd.IpTableAcl) (bool, error) {
	h.logger.Info("Delete Ip Table rule " + ipaclConfig.Name)
	//return (h.server.DelIpTableRule(ipaclConfig))
	h.server.IptableDelCh <- ipaclConfig
	return true, nil
}

func (h *SYSDHandler) ExecuteActionDaemon(daemonConfig *sysd.Daemon) (bool, error) {
	h.logger.Info(fmt.Sprintln("Daemon action attrs: ", daemonConfig))
	dConf := server.DaemonConfig{
		Name:  daemonConfig.Name,
		State: daemonConfig.State,
	}
	h.server.DaemonConfigCh <- dConf
	return true, nil
}

func (h *SYSDHandler) convertDaemonStateToThrift(ent server.DaemonState) *sysd.DaemonState {
	dState := sysd.NewDaemonState()
	dState.Name = string(ent.Name)
	dState.State = string(sysdCommonDefs.ConvertDaemonStateCodeToString(ent.State))
	dState.Reason = string(ent.Reason)
	kaStr := fmt.Sprintf("Received %d keepalives", ent.RecvedKACount)
	dState.KeepAlive = string(kaStr)
	dState.RestartCount = int32(ent.NumRestarts)
	dState.RestartTime = string(ent.RestartTime)
	dState.RestartReason = string(ent.RestartReason)
	return dState
}

func (h *SYSDHandler) GetDaemonState(name string) (*sysd.DaemonState, error) {
	h.logger.Info(fmt.Sprintln("Get Daemon state ", name))
	daemonStateResponse := sysd.NewDaemonState()
	dState := h.server.GetDaemonState(name)
	daemonState := h.convertDaemonStateToThrift(*dState)
	daemonStateResponse = daemonState
	return daemonStateResponse, nil
}

func (h *SYSDHandler) GetBulkDaemonState(fromIdx sysd.Int, count sysd.Int) (*sysd.DaemonStateGetInfo, error) {
	h.logger.Info(fmt.Sprintln("Get Daemon states "))
	nextIdx, currCount, daemonStates := h.server.GetBulkDaemonStates(int(fromIdx), int(count))
	if daemonStates == nil {
		err := errors.New("System server is busy")
		return nil, err
	}
	daemonStatesResponse := make([]*sysd.DaemonState, len(daemonStates))
	for idx, item := range daemonStates {
		daemonStatesResponse[idx] = h.convertDaemonStateToThrift(item)
	}
	daemonStateGetInfo := sysd.NewDaemonStateGetInfo()
	daemonStateGetInfo.Count = sysd.Int(currCount)
	daemonStateGetInfo.StartIdx = sysd.Int(fromIdx)
	daemonStateGetInfo.EndIdx = sysd.Int(nextIdx)
	daemonStateGetInfo.More = (nextIdx != 0)
	daemonStateGetInfo.DaemonStateList = daemonStatesResponse
	return daemonStateGetInfo, nil
}

func (h *SYSDHandler) PeriodicKeepAlive(name string) error {
	h.server.KaRecvCh <- name
	return nil
}
