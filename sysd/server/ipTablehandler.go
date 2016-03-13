package server

import (
	"sysd"
)

func (server *SYSDServer) AddIpTableRule(ipaclConfig *sysd.IpTableAclConfig,
	restart bool) (bool, error) {
	return (server.sysdIpTableMgr.AddIpRule(ipaclConfig, restart))
}

func (server *SYSDServer) DelIpTableRule(ipaclConfig *sysd.IpTableAclConfig) (bool, error) {
	return (server.sysdIpTableMgr.DelIpRule(ipaclConfig))
}
