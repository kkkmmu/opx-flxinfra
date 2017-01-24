//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//       Unless required by applicable law or agreed to in writing, software
//       distributed under the License is distributed on an "AS IS" BASIS,
//       WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//       See the License for the specific language governing permissions and
//       limitations under the License.
//
// _______   __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----  \   \/    \/   /  |  |  ---|  |----    ,---- |  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |        |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |        `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//

package server

import (
	"errors"
	"infra/statsd/objects"
	"net"
)

const (
	SFLOW_PKT_MAX_SAMPLED_SIZE = 256
	MAX_UDP_DATAGRAM_SIZE      = 65507 //2^16 - 8b(UDP Hdr) - 20b(IP Hdr)
)

//Sflow global object functions
func (srvr *sflowServer) isGlobalObjCreated() bool {
	var ok bool
	if srvr.sflowGblDB != nil {
		ok = true
	} else {
		ok = false
	}
	return ok
}

func (srvr *sflowServer) ValidateCreateSflowGlobal(obj *objects.SflowGlobal) (bool, error) {
	var ok bool
	var err error
	if srvr.isGlobalObjCreated() {
		err = errors.New("Create SflowGlobal failed. Sflow global object already exists")
	} else if initFailed {
		err = errors.New("Create SflowGlobal failed. Server initalization failed, no SFLOW configuration will be accepted")
	} else {
		ok = true
	}
	return ok, err
}

func (srvr *sflowServer) createSflowGlobal(obj *objects.SflowGlobal) {
	//No error checking required as API layer has completed validation
	srvr.sflowGblDB = &sflowGbl{
		vrf:                 obj.Vrf,
		adminState:          obj.AdminState,
		agentIpAddr:         obj.AgentIpAddr,
		maxSampledSize:      obj.MaxSampledSize,
		counterPollInterval: obj.CounterPollInterval,
		maxDatagramSize:     obj.MaxDatagramSize,
	}
	//Post msg to core server to notify gbl cfg
	srvr.gblCfgCh <- &gblCfgUpdateInfo{
		op:  objects.SFLOW_GLOBAL_UPDATE_ATTR_AGENT_IPADDR,
		val: interface{}(obj.AgentIpAddr),
	}
}

func (srvr *sflowServer) ValidateUpdateSflowGlobal(oldObj, newObj *objects.SflowGlobal, attrset []bool) (bool, error) {
	var ok bool
	var err error
	if srvr.isGlobalObjCreated() {
		if (newObj.AdminState != objects.ADMIN_STATE_UP) && (newObj.AdminState != objects.ADMIN_STATE_DOWN) {
			err = errors.New("Update SflowGlobal failed. Invalid AdminState value provided")
		} else if net.ParseIP(newObj.AgentIpAddr) == nil {
			err = errors.New("Update SflowGlobal failed. Invalid agent ip address attribute value provided")
		} else if (newObj.MaxSampledSize) < 0 || (newObj.MaxSampledSize > SFLOW_PKT_MAX_SAMPLED_SIZE) {
			err = errors.New("Update SflowGlobal failed. Invalid MaxSampledSize value provided")
		} else if newObj.CounterPollInterval < 0 {
			err = errors.New("Update SflowGlobal failed. Invalid CounterPollInterval value provided")
		} else if (newObj.MaxDatagramSize < 0) || (newObj.MaxDatagramSize > MAX_UDP_DATAGRAM_SIZE) {
			err = errors.New("Create SflowCollector failed. Invalid MaxDatagramSize value provided")
		} else {
			ok = true
		}
	} else {
		err = errors.New("Update SflowGlobal failed, global object not created yet")
	}
	return ok, err
}

func genSflowGlobalUpdateMask(attrset []bool) uint8 {
	var mask uint8
	for idx, val := range attrset {
		if val {
			switch idx {
			case objects.SFLOW_GLOBAL_ATTR_ADMIN_STATE_IDX:
				mask |= objects.SFLOW_GLOBAL_UPDATE_ATTR_ADMIN_STATE
			case objects.SFLOW_GLOBAL_ATTR_AGENT_IPADDR_IDX:
				mask |= objects.SFLOW_GLOBAL_UPDATE_ATTR_AGENT_IPADDR
			case objects.SFLOW_GLOBAL_ATTR_COUNTER_POLL_INTERVAL_IDX:
				mask |= objects.SFLOW_GLOBAL_UPDATE_ATTR_COUNTER_POLL_INTERVAL
			case objects.SFLOW_GLOBAL_ATTR_MAX_SAMPLE_SIZE_IDX:
				mask |= objects.SFLOW_GLOBAL_UPDATE_ATTR_MAX_SAMPLE_SIZE
			case objects.SFLOW_GLOBAL_ATTR_MAX_DATAGRAM_SIZE_IDX:
				mask |= objects.SFLOW_GLOBAL_UPDATE_ATTR_MAX_DATAGRAM_SIZE
			}
		}
	}
	return mask
}

func (srvr *sflowServer) updateSflowGlobal(oldObj, newObj *objects.SflowGlobal, attrset []bool) {
	//Update object in DB
	srvr.dbMutex.Lock()
	srvr.sflowGblDB.vrf = newObj.Vrf
	srvr.sflowGblDB.adminState = newObj.AdminState
	srvr.sflowGblDB.agentIpAddr = newObj.AgentIpAddr
	srvr.sflowGblDB.maxSampledSize = newObj.MaxSampledSize
	srvr.sflowGblDB.counterPollInterval = newObj.CounterPollInterval
	srvr.dbMutex.Unlock()

	mask := genSflowGlobalUpdateMask(attrset)
	if (mask & objects.SFLOW_GLOBAL_UPDATE_ATTR_ADMIN_STATE) == objects.SFLOW_GLOBAL_UPDATE_ATTR_ADMIN_STATE {
		//Hold the global lock to gate all other configurations from being accepted
		srvr.dbMutex.Lock()
		if newObj.AdminState == objects.ADMIN_STATE_DOWN {
			//Effect admin down processing on all sflow collector/sflow interface go routines
			for ifIndex, intf := range srvr.sflowIntfDB {
				srvr.updateSflowIntfAdminState(intf, ifIndex, objects.ADMIN_STATE_DOWN)
			}
			for _, collector := range srvr.sflowCollectorDB {
				srvr.updateSflowCollectorAdminState(collector, objects.ADMIN_STATE_DOWN)
			}
		} else {
			//Effect admin up processing on all sflow collector/sflow interface go routines
			for ifIndex, intf := range srvr.sflowIntfDB {
				if intf.adminState == objects.ADMIN_STATE_UP {
					srvr.updateSflowIntfAdminState(intf, ifIndex, objects.ADMIN_STATE_UP)
				}
			}
			for _, collector := range srvr.sflowCollectorDB {
				if collector.adminState == objects.ADMIN_STATE_UP {
					srvr.updateSflowCollectorAdminState(collector, objects.ADMIN_STATE_UP)
				}
			}
		}
		srvr.dbMutex.Unlock()
	}
	if (mask & objects.SFLOW_GLOBAL_UPDATE_ATTR_AGENT_IPADDR) == objects.SFLOW_GLOBAL_UPDATE_ATTR_AGENT_IPADDR {
		srvr.sflowGblDB.agentIpAddr = newObj.AgentIpAddr
		srvr.gblCfgCh <- &gblCfgUpdateInfo{
			op:  objects.SFLOW_GLOBAL_UPDATE_ATTR_AGENT_IPADDR,
			val: interface{}(newObj.AgentIpAddr),
		}
	}
	if (mask & objects.SFLOW_GLOBAL_UPDATE_ATTR_COUNTER_POLL_INTERVAL) == objects.SFLOW_GLOBAL_UPDATE_ATTR_COUNTER_POLL_INTERVAL {
		//Adjust counter thread polling interval. Post update to all interface polling routines
		for _, intfObj := range srvr.sflowIntfDB {
			//Post update to intf poller if operational
			if intfObj.operstate == objects.ADMIN_STATE_UP {
				intfObj.restartCtrPollTicker <- newObj.CounterPollInterval
			}
		}
	}
	if (mask & objects.SFLOW_GLOBAL_UPDATE_ATTR_MAX_SAMPLE_SIZE) == objects.SFLOW_GLOBAL_UPDATE_ATTR_MAX_SAMPLE_SIZE {
		//No-op, core interface RX routine will pickup update automagically
	}
	if (mask & objects.SFLOW_GLOBAL_UPDATE_ATTR_MAX_DATAGRAM_SIZE) == objects.SFLOW_GLOBAL_UPDATE_ATTR_MAX_DATAGRAM_SIZE {
		//FIXME: Currently not performing any aggregation, i.e. only one sample record per datagram
	}
}

func (srvr *sflowServer) ValidateDeleteSflowGlobal(obj *objects.SflowGlobal) (bool, error) {
	var ok bool
	var err error
	if srvr.isGlobalObjCreated() {
		//Validate all collector/interface configuration has been deleted
		if len(srvr.sflowCollectorDB) != 0 {
			err = errors.New("Delete SflowGlobal failed. Please delete all sflow collector config before deleting SflowGlobal")
		} else if len(srvr.sflowIntfDB) != 0 {
			err = errors.New("Delete SflowGlobal failed. Please delete all sflow interface config before deleting SflowGlobal")
		} else {
			ok = true
		}
	} else {
		err = errors.New("Delete SflowGlobal failed. Sflow global object not created")
	}
	return ok, err
}

func (srvr *sflowServer) deleteSflowGlobal(obj *objects.SflowGlobal) {
	srvr.sflowGblDB = nil
}
