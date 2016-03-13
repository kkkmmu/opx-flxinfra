package rpc

import (
	"errors"
	"fmt"
	"infra/sysd/server"
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

func (h *SYSDHandler) SendGlobalLoggingConfig(gLoggingConfig *sysd.SystemLoggingConfig) bool {
	var Logging bool
	if gLoggingConfig.SystemLogging == "on" {
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

func (h *SYSDHandler) SendComponentLoggingConfig(cLoggingConfig *sysd.ComponentLoggingConfig) bool {
	cConf := server.ComponentLoggingConfig{
		Component: cLoggingConfig.Module,
		Level:     logging.ConvertLevelStrToVal(cLoggingConfig.Level),
	}
	h.server.ComponentLoggingConfigCh <- cConf
	return true
}

func (h *SYSDHandler) UpdateSystemLoggingConfig(origConf *sysd.SystemLoggingConfig, newConf *sysd.SystemLoggingConfig, attrset []bool) (bool, error) {
	h.logger.Info(fmt.Sprintln("Original global config attrs:", origConf))
	if newConf == nil {
		err := errors.New("Invalid global Configuration")
		return false, err
	}
	h.logger.Info(fmt.Sprintln("Update global config attrs:", newConf))
	return h.SendGlobalLoggingConfig(newConf), nil
}

func (h *SYSDHandler) UpdateComponentLoggingConfig(origConf *sysd.ComponentLoggingConfig, newConf *sysd.ComponentLoggingConfig, attrset []bool) (bool, error) {
	h.logger.Info(fmt.Sprintln("Original component config attrs:", origConf))
	if newConf == nil {
		err := errors.New("Invalid component Configuration")
		return false, err
	}
	h.logger.Info(fmt.Sprintln("Update component config attrs:", newConf))
	return h.SendComponentLoggingConfig(newConf), nil
}

func (h *SYSDHandler) CreateSystemLoggingConfig(gLoggingConf *sysd.SystemLoggingConfig) (bool, error) {
	h.logger.Info(fmt.Sprintln("Create global config attrs:", gLoggingConf))
	return h.SendGlobalLoggingConfig(gLoggingConf), nil
}

func (h *SYSDHandler) CreateComponentLoggingConfig(cLoggingConf *sysd.ComponentLoggingConfig) (bool, error) {
	h.logger.Info(fmt.Sprintln("Create component config attrs:", cLoggingConf))
	return h.SendComponentLoggingConfig(cLoggingConf), nil
}

func (h *SYSDHandler) DeleteSystemLoggingConfig(gLoggingConf *sysd.SystemLoggingConfig) (bool, error) {
	h.logger.Info(fmt.Sprintln("Delete global config attrs:", gLoggingConf))
	return true, nil
}

func (h *SYSDHandler) DeleteComponentLoggingConfig(cLoggingConf *sysd.ComponentLoggingConfig) (bool, error) {
	h.logger.Info(fmt.Sprintln("Delete component config attrs:", cLoggingConf))
	return true, nil
}

func (h *SYSDHandler) CreateIpTableAclConfig(ipaclConfig *sysd.IpTableAclConfig) (bool, error) {
	h.logger.Info("Create Ip Table rule " + ipaclConfig.Name)
	return (h.server.AddIpTableRule(ipaclConfig, false /* non - restart*/))
}

func (h *SYSDHandler) UpdateIpTableAclConfig(origConf *sysd.IpTableAclConfig,
	newConf *sysd.IpTableAclConfig, attrset []bool) (bool, error) {
	return true, nil
}

func (h *SYSDHandler) DeleteIpTableAclConfig(ipaclConfig *sysd.IpTableAclConfig) (bool, error) {
	h.logger.Info("Delete Ip Table rule " + ipaclConfig.Name)
	return (h.server.DelIpTableRule(ipaclConfig))
}
