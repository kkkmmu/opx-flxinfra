package ipTable

import (
	_ "fmt"
	_ "net"
	"strconv"
	"strings"
	"sysd"
	"utils/logging"
)

/*
#cgo CFLAGS: -I../../../netfilter/libiptables/include -I../../../netfilter/iptables/include
#cgo LDFLAGS: -L../../../netfilter/libiptables/lib -lip4tc
#include "ipTable.h"
*/
import "C"

const (
	ALL_RULE_STR = "all"
)

type SysdIpTableHandler struct {
	logger   *logging.Writer
	ruleInfo map[string]C.ipt_config_t
}

func SysdNewSysdIpTableHandler(logger *logging.Writer) *SysdIpTableHandler {
	ipTableHdl := &SysdIpTableHandler{}
	ipTableHdl.logger = logger
	ipTableHdl.ruleInfo = make(map[string]C.ipt_config_t, 100)
	return ipTableHdl
}

func (hdl *SysdIpTableHandler) AddIpRule(config *sysd.IpTableAclConfig) {
	port, err := strconv.Atoi(config.Port)
	var iptEntry C.ipt_config_t
	rv := -1
	if err != nil {
		if strings.Compare(config.Port, ALL_RULE_STR) == 0 {
			hdl.logger.Info("Rule to be applied on all ports")
			port = 0
		}
	}
	ip := config.IpAddr
	pl := 0 // prefix length
	if strings.Contains(config.IpAddr, "/") {
		splitStr := strings.Split(config.IpAddr, "/")
		ip = splitStr[0]
		pl, _ = strconv.Atoi(splitStr[1])
	}
	entry := &C.rule_entry_t{
		Name:         C.CString(config.Name),
		PhysicalPort: C.CString(config.PhysicalPort),
		Action:       C.CString(config.Action),
		IpAddr:       C.CString(ip),
		PrefixLength: C.int(pl),
		Protocol:     C.CString(config.Protocol),
		Port:         C.uint16_t(port),
	}
	switch config.Protocol {
	case "udp":
		rv = int(C.add_iptable_udp_rule(entry, &iptEntry))

	case "tcp":
		rv = int(C.add_iptable_tcp_rule(entry, &iptEntry))

	case "icmp":
		rv = int(C.add_iptable_icmp_rule(entry, &iptEntry))
	default:
		hdl.logger.Err("Rule adding for " + config.Protocol +
			" is not supported")
		return
	}
	if rv <= 0 {
		return
	}
	hdl.ruleInfo[config.Name] = iptEntry
}

func (hdl *SysdIpTableHandler) DelIpRule(config *sysd.IpTableAclConfig) {
	entry, entryFound := hdl.ruleInfo[config.Name]
	if !entryFound {
		hdl.logger.Err("No rule found for " + config.Name)
		return
	}

	rv := int(C.del_iptable_rule(&entry))
	if rv <= 0 {
		hdl.logger.Err("Delete rule failed for " + config.Name)
		return
	}
	delete(hdl.ruleInfo, config.Name)
}
