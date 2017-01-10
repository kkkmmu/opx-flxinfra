//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//	 Unless required by applicable law or agreed to in writing, software
//	 distributed under the License is distributed on an "AS IS" BASIS,
//	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	 See the License for the specific language governing permissions and
//	 limitations under the License.
//
// _______  __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//

package rpc

import (
	"errors"
	"infra/sysd/server"
	"infra/sysd/sysdCommonDefs"
	"models/objects"
	"net"
	"strings"
	"sysd"
	"utils/logging"
)

const (
	SPECIAL_HOSTNAME_CHARS = "_"
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

func (h *SYSDHandler) SendComponentLoggingConfig(cLoggingConfig *sysd.ComponentLogging) bool {
	cConf := server.ComponentLoggingConfig{
		Component: cLoggingConfig.Module,
		Level:     logging.ConvertLevelStrToVal(cLoggingConfig.Level),
	}
	h.server.ComponentLoggingConfigCh <- cConf
	return true
}

func (h *SYSDHandler) UpdateComponentLogging(origConf *sysd.ComponentLogging, newConf *sysd.ComponentLogging, attrset []bool, op []*sysd.PatchOpInfo) (bool, error) {
	h.logger.Info("Original component config attrs:", origConf)
	if newConf == nil {
		err := errors.New("Invalid component Configuration")
		return false, err
	}
	h.logger.Info("Update component config attrs:", newConf)
	return h.SendComponentLoggingConfig(newConf), nil
}

func (h *SYSDHandler) CreateComponentLogging(cLoggingConf *sysd.ComponentLogging) (bool, error) {
	h.logger.Info("Create component config attrs:", cLoggingConf)
	return h.SendComponentLoggingConfig(cLoggingConf), nil
}

func (h *SYSDHandler) DeleteComponentLogging(cLoggingConf *sysd.ComponentLogging) (bool, error) {
	h.logger.Info("Delete component config attrs:", cLoggingConf)
	err := errors.New("CompoenentLogging delete not supported")
	return false, err
}

func (h *SYSDHandler) CreateIpTableAcl(ipaclConfig *sysd.IpTableAcl) (bool, error) {
	h.logger.Info("Create Ip Table rule " + ipaclConfig.Name)
	return h.server.AddRule(ipaclConfig)
}

func (h *SYSDHandler) UpdateIpTableAcl(origConf *sysd.IpTableAcl, newConf *sysd.IpTableAcl,
	attrset []bool, op []*sysd.PatchOpInfo) (bool, error) {
	err := errors.New("IpTableAcl update not supported")
	return false, err
}

func (h *SYSDHandler) DeleteIpTableAcl(ipaclConfig *sysd.IpTableAcl) (bool, error) {
	h.logger.Info("Delete Ip Table rule " + ipaclConfig.Name)
	return h.server.DelRule(ipaclConfig)
}

func (h *SYSDHandler) ExecuteActionDaemon(daemonConfig *sysd.Daemon) (bool, error) {
	h.logger.Info("Daemon action attrs: ", daemonConfig)
	daemonInfo, exist := h.server.DaemonMap[daemonConfig.Name]
	if exist {
		if daemonInfo.State == sysdCommonDefs.UP && daemonConfig.Op == "start" {
			err := errors.New("Daemon " + daemonConfig.Name + " is already up")
			return false, err
		}
		if daemonInfo.State == sysdCommonDefs.STOPPED && daemonConfig.Op == "stop" {
			err := errors.New("Daemon " + daemonConfig.Name + " is already stopped")
			return false, err
		}
		if daemonInfo.State == sysdCommonDefs.RESTARTING && daemonConfig.Op == "restart" {
			err := errors.New("Daemon " + daemonConfig.Name + " is in restarting state")
			return false, err
		}
		dConf := server.DaemonConfig{
			Name:     daemonConfig.Name,
			Op:       daemonConfig.Op,
			WatchDog: daemonConfig.WatchDog,
		}
		h.server.DaemonConfigCh <- dConf
		return true, nil
	} else {
		h.logger.Err("Unknown daemon name: ", daemonConfig)
		err := errors.New("Unknown daemon name " + daemonConfig.Name)
		return false, err
	}
}

func (h *SYSDHandler) ExecuteActionGlobalLogging(gLogConfig *sysd.GlobalLogging) (bool, error) {
	h.logger.Info("GlobalLogging action attrs: ", gLogConfig)
	gConf := server.GlobalLoggingConfig{
		Level: logging.ConvertLevelStrToVal(gLogConfig.Level),
	}
	h.server.GlobalLoggingConfigCh <- gConf
	return true, nil
}

func (h *SYSDHandler) GetDaemonState(name string) (*sysd.DaemonState, error) {
	h.logger.Info("Get Daemon state ", name)
	daemonStateResponse := sysd.NewDaemonState()
	dState := h.server.GetDaemonState(name)
	daemonState := h.server.ConvertDaemonStateToThrift(*dState)
	daemonStateResponse = daemonState
	return daemonStateResponse, nil
}

func (h *SYSDHandler) GetBulkDaemonState(fromIdx sysd.Int, count sysd.Int) (*sysd.DaemonStateGetInfo, error) {
	h.logger.Info("Get Daemon states ")
	nextIdx, currCount, daemonStates := h.server.GetBulkDaemonStates(int(fromIdx), int(count))
	if daemonStates == nil {
		err := errors.New("System server is busy")
		return nil, err
	}
	daemonStatesResponse := make([]*sysd.DaemonState, len(daemonStates))
	for idx, item := range daemonStates {
		daemonStatesResponse[idx] = h.server.ConvertDaemonStateToThrift(item)
	}
	daemonStateGetInfo := sysd.NewDaemonStateGetInfo()
	daemonStateGetInfo.Count = sysd.Int(currCount)
	daemonStateGetInfo.StartIdx = sysd.Int(fromIdx)
	daemonStateGetInfo.EndIdx = sysd.Int(nextIdx)
	daemonStateGetInfo.More = (nextIdx != 0)
	daemonStateGetInfo.DaemonStateList = daemonStatesResponse
	return daemonStateGetInfo, nil
}

func convertSystemParamThriftToModel(cfg *sysd.SystemParam) objects.SystemParam {
	confg := objects.SystemParam{
		Description: cfg.Description,
		SwVersion:   cfg.SwVersion,
		MgmtIp:      cfg.MgmtIp,
		Hostname:    cfg.Hostname,
		SwitchMac:   cfg.SwitchMac,
		Vrf:         cfg.Vrf,
	}
	return confg
}

func (h *SYSDHandler) CreateSystemParam(cfg *sysd.SystemParam) (bool, error) {
	if h.server.SystemInfoCreated() {
		return false, errors.New("System Params Info is already created for Default VRF, please do update to modify params")
	}
	if strings.ContainsAny(cfg.Hostname, SPECIAL_HOSTNAME_CHARS) {
		return false, errors.New("Hostname should not have Special Characters " +
			SPECIAL_HOSTNAME_CHARS)
	}
	h.logger.Info("Configuring Global Object", cfg)
	confg := convertSystemParamThriftToModel(cfg)
	h.server.SystemParamConfig <- confg
	return true, nil
}

func (h *SYSDHandler) validatUpdateSystemParam(newCfg *sysd.SystemParam, attrset []bool) ([]string, error) {
	var updatedInfo []string
	/*
		0 : string Vrf
		1 : string MgmtIp
		2 : string Hostname
		3 : string Version
		4 : string SwitchMac
		5 : string Description
	*/
	for idx, _ := range attrset {
		if attrset[idx] == false {
			continue
		}
		switch idx {
		case 0:
			return updatedInfo, errors.New("VRF update is not supported")
		case 1:
			return updatedInfo, errors.New("MgmtIp update is not supported")
		case 2:
			if strings.ContainsAny(newCfg.Hostname, SPECIAL_HOSTNAME_CHARS) {
				return updatedInfo, errors.New("Hostname should not have Special Characters " +
					SPECIAL_HOSTNAME_CHARS)
			}
			updatedInfo = append(updatedInfo, "Hostname")
		case 3:
			return updatedInfo, errors.New("Version update is not supported")
		case 4:
			return updatedInfo, errors.New("Switch Mac Address update is not supported")
		case 5:
			updatedInfo = append(updatedInfo, "Description")
		}
	}
	return updatedInfo, nil
}

func (h *SYSDHandler) UpdateSystemParam(org *sysd.SystemParam, new *sysd.SystemParam, attrset []bool,
	op []*sysd.PatchOpInfo) (bool, error) {
	h.logger.Info("Received update for system param information", org, new, attrset)
	if org == nil || new == nil {
		return false, errors.New("Invalid information provided to server")
	}
	entriesUpdated, err := h.validatUpdateSystemParam(new, attrset)
	if err != nil {
		h.logger.Err("Failed to update system params:", err)
		return false, err
	}
	cfg := convertSystemParamThriftToModel(new)
	updInfo := server.SystemParamUpdate{
		EntriesUpdated: entriesUpdated,
		NewCfg:         &cfg,
	}
	h.server.SysUpdCh <- &updInfo
	h.logger.Info("Returning true from update system param information")
	return true, nil
}

func (h *SYSDHandler) DeleteSystemParam(cfg *sysd.SystemParam) (bool, error) {
	return false, errors.New("Delete of system params for default vrf is not supported")
}

func convertSystemParamStateToThrift(info objects.SystemParamState, entry *sysd.SystemParamState) {
	entry.Vrf = string(info.Vrf)
	entry.SwitchMac = string(info.SwitchMac)
	entry.MgmtIp = string(info.MgmtIp)
	entry.SwVersion = string(info.SwVersion)
	entry.Description = string(info.Description)
	entry.Hostname = string(info.Hostname)
	entry.Distro = string(info.Distro)
	entry.Kernel = string(info.Kernel)
}

func (h *SYSDHandler) GetSystemParamState(name string) (*sysd.SystemParamState, error) {
	h.logger.Info("Get system params info for " + name)
	sysParamsResp := sysd.NewSystemParamState()
	sysInfo := h.server.GetSystemParam(name)
	if sysInfo == nil {
		return nil, errors.New("No Matching entry for " + name + "found")
	}
	convertSystemParamStateToThrift(*sysInfo, sysParamsResp)
	h.logger.Info("Returing System Info:", sysParamsResp)
	return sysParamsResp, nil
}

func (h *SYSDHandler) GetBulkSystemParamState(fromIdx sysd.Int, count sysd.Int) (*sysd.SystemParamStateGetInfo, error) {
	//@TODO: when we support vrf change get bulk... today only one system info is present
	sysParamsResp, err := h.GetSystemParamState("default")
	if err != nil {
		return nil, err
	}
	systemGetInfoResp := sysd.NewSystemParamStateGetInfo()
	systemGetInfoResp.Count = 1
	systemGetInfoResp.StartIdx = 0
	systemGetInfoResp.EndIdx = 1
	systemGetInfoResp.More = false
	respList := make([]*sysd.SystemParamState, 1)
	//respList = append(respList, sysParamsResp)
	respList[0] = sysParamsResp
	systemGetInfoResp.SystemParamStateList = respList
	return systemGetInfoResp, nil
}

func validateTacacsConfig(conf *sysd.TacacsConfig) (bool, error) {
	if conf.AuthService != "login" &&
		conf.AuthService != "enable" &&
		conf.AuthService != "ppp" &&
		conf.AuthService != "arap" &&
		conf.AuthService != "pt" &&
		conf.AuthService != "rcmd" &&
		conf.AuthService != "x25" &&
		conf.AuthService != "nasi" &&
		conf.AuthService != "fwproxy" {

		return false, errors.New("Invalid value for AuthService")
	}
	if conf.Port <= 0 {
		return false, errors.New("Invalid value for Port")
	}
	if conf.PrivilegeLevel < 1 || conf.PrivilegeLevel > 15 {
		return false, errors.New("Invalid value for PrivilegeLevel")
	}
	if conf.Debug != 0 && conf.Debug != 1 {
		return false, errors.New("Invalid value for Debug")
	}
	return true, nil
}

func (h *SYSDHandler) CreateTacacsConfig(tacacsConfig *sysd.TacacsConfig) (bool, error) {
	trial := net.ParseIP(tacacsConfig.ServerIp)
	if trial.To4() == nil && trial.To16() == nil {
		return false, errors.New("Invalid value for ServerIp")
	}
	validationOk, err := validateTacacsConfig(tacacsConfig)
	if !validationOk {
		return false, err
	}
	conf := &server.TacacsConfig{
		ServerIp:       tacacsConfig.ServerIp,
		SourceIntf:     tacacsConfig.SourceIntf,
		AuthService:    tacacsConfig.AuthService,
		Secret:         tacacsConfig.Secret,
		Port:           tacacsConfig.Port,
		PrivilegeLevel: tacacsConfig.PrivilegeLevel,
		Debug:          tacacsConfig.Debug,
	}
	return h.server.CreateTacacsConfig(conf)
}

func (h *SYSDHandler) UpdateTacacsConfig(origConf *sysd.TacacsConfig, newConf *sysd.TacacsConfig,
	attrset []bool, op []*sysd.PatchOpInfo) (bool, error) {

	validationOk, err := validateTacacsConfig(newConf)
	if !validationOk {
		return false, err
	}
	conf := &server.TacacsConfig{
		ServerIp:       newConf.ServerIp,
		SourceIntf:     newConf.SourceIntf,
		AuthService:    newConf.AuthService,
		Secret:         newConf.Secret,
		Port:           newConf.Port,
		PrivilegeLevel: newConf.PrivilegeLevel,
		Debug:          newConf.Debug,
	}
	return h.server.UpdateTacacsConfig(conf)
}

func (h *SYSDHandler) DeleteTacacsConfig(tacacsConfig *sysd.TacacsConfig) (bool, error) {
	return h.server.DeleteTacacsConfig(tacacsConfig.ServerIp)
}

func (h *SYSDHandler) GetTacacsState(name string) (*sysd.TacacsState, error) {
	serverObj, err := h.server.GetTacacsState(name)
	if err != nil {
		return nil, err
	}
	retObj := &sysd.TacacsState{
		ServerIp:       serverObj.ServerIp,
		SourceIntf:     serverObj.SourceIntf,
		AuthService:    serverObj.AuthService,
		Secret:         serverObj.Secret,
		Port:           serverObj.Port,
		PrivilegeLevel: serverObj.PrivilegeLevel,
		Debug:          serverObj.Debug,
		ConnFailReason: serverObj.ConnFailReason,
	}
	return retObj, nil
}

func (h *SYSDHandler) GetBulkTacacsState(fromIdx sysd.Int, count sysd.Int) (*sysd.TacacsStateGetInfo, error) {
	result := &sysd.TacacsStateGetInfo{}
	nextIdx, actualCount, more, tacacsStateSlice :=
		h.server.GetTacacsStateSlice(int(fromIdx), int(count))

	result.StartIdx = sysd.Int(fromIdx)
	result.EndIdx = sysd.Int(nextIdx)
	result.Count = sysd.Int(actualCount)
	result.More = more
	result.TacacsStateList = []*sysd.TacacsState{}
	for _, tacacsState := range tacacsStateSlice {
		tmp := &sysd.TacacsState{
			ServerIp:       tacacsState.ServerIp,
			SourceIntf:     tacacsState.SourceIntf,
			AuthService:    tacacsState.AuthService,
			Secret:         tacacsState.Secret,
			Port:           tacacsState.Port,
			PrivilegeLevel: tacacsState.PrivilegeLevel,
			Debug:          tacacsState.Debug,
			ConnFailReason: tacacsState.ConnFailReason,
		}
		result.TacacsStateList = append(result.TacacsStateList, tmp)
	}
	return result, nil
}

func validateTacacsGlobalConfig(conf *sysd.TacacsGlobalConfig) (bool, error) {
	if conf.Enable != "true" && conf.Enable != "false" {
		return false, errors.New("Invalid value for Enable")
	}
	if conf.Timeout < 0 {
		return false, errors.New("Invalid value for Timeout")
	}
	return true, nil
}

func (h *SYSDHandler) CreateTacacsGlobalConfig(tacacsGlobalConfig *sysd.TacacsGlobalConfig) (bool, error) {
	validationOk, err := validateTacacsGlobalConfig(tacacsGlobalConfig)
	if !validationOk {
		return false, err
	}
	conf := &server.TacacsGlobalConfig{
		ProfileName: tacacsGlobalConfig.ProfileName,
		Enable:      tacacsGlobalConfig.Enable,
		Timeout:     tacacsGlobalConfig.Timeout,
	}
	return h.server.CreateTacacsGlobalConfig(conf)
}

func (h *SYSDHandler) UpdateTacacsGlobalConfig(origConf *sysd.TacacsGlobalConfig, newConf *sysd.TacacsGlobalConfig,
	attrset []bool, op []*sysd.PatchOpInfo) (bool, error) {

	validationOk, err := validateTacacsGlobalConfig(newConf)
	if !validationOk {
		return false, err
	}
	conf := &server.TacacsGlobalConfig{
		ProfileName: newConf.ProfileName,
		Enable:      newConf.Enable,
		Timeout:     newConf.Timeout,
	}
	return h.server.UpdateTacacsGlobalConfig(conf)
}

func (h *SYSDHandler) DeleteTacacsGlobalConfig(tacacsGlobalConfig *sysd.TacacsGlobalConfig) (bool, error) {
	return false, nil
}

func (h *SYSDHandler) GetTacacsGlobalState(name string) (*sysd.TacacsGlobalState, error) {
	serverObj, err := h.server.GetTacacsGlobalState(name)
	if err != nil {
		return nil, err
	}
	retObj := &sysd.TacacsGlobalState{
		ProfileName:       serverObj.ProfileName,
		OperStatus:        serverObj.OperStatus,
		NumActiveSessions: serverObj.NumActiveSessions,
	}
	return retObj, nil
}

func (h *SYSDHandler) GetBulkTacacsGlobalState(fromIdx sysd.Int, count sysd.Int) (*sysd.TacacsGlobalStateGetInfo, error) {
	result := &sysd.TacacsGlobalStateGetInfo{}
	nextIdx, actualCount, more, tacacsGlobalStateSlice :=
		h.server.GetTacacsGlobalStateSlice(int(fromIdx), int(count))

	result.StartIdx = sysd.Int(fromIdx)
	result.EndIdx = sysd.Int(nextIdx)
	result.Count = sysd.Int(actualCount)
	result.More = more
	result.TacacsGlobalStateList = []*sysd.TacacsGlobalState{}
	for _, tacacsGlobalState := range tacacsGlobalStateSlice {
		tmp := &sysd.TacacsGlobalState{
			ProfileName:       tacacsGlobalState.ProfileName,
			OperStatus:        tacacsGlobalState.OperStatus,
			NumActiveSessions: tacacsGlobalState.NumActiveSessions,
		}
		result.TacacsGlobalStateList = append(result.TacacsGlobalStateList, tmp)
	}
	return result, nil
}

/*********************** SYSD INTERNAL API **************************/

func (h *SYSDHandler) PeriodicKeepAlive(name string) error {
	h.server.KaRecvCh <- name
	return nil
}
