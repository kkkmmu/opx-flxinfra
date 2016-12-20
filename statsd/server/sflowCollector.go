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
//   This is a auto-generated file, please do not edit!
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
	MAX_UDP_PORT = 65535
)

func newSflowCollector() *sflowCollector {
	obj := new(sflowCollector)
	obj.shutdownCh = make(chan bool)
	return obj
}

//Sflow collector related functions
func (srvr *sflowServer) ValidateCreateSflowCollector(obj *objects.SflowCollector) (bool, error) {
	var ok bool
	var err error
	collectorIP := net.ParseIP(obj.IpAddr)
	if collectorIP == nil {
		err = errors.New("Create SflowCollector failed. Invalid collector IP value provided")
	} else if _, ok := srvr.sflowCollectorDB[obj.IpAddr]; ok {
		err = errors.New("Create SflowCollector failed. Collector configuration already exists")
	} else if (obj.UdpPort < 0) || (obj.UdpPort > MAX_UDP_PORT) {
		err = errors.New("Create SflowCollector failed. Invalid UDP port value provided")
	} else if (obj.AdminState != objects.ADMIN_STATE_UP) && (obj.AdminState != objects.ADMIN_STATE_DOWN) {
		err = errors.New("Create SflowCollectore failed. Invalid AdminState value provided")
	} else {
		ok = true
	}
	return ok, err
}

func (srvr *sflowServer) createSflowCollector(obj *objects.SflowCollector) {
	sflowCollectorObj := newSflowCollector()
	sflowCollectorObj.ipAddr = obj.IpAddr
	sflowCollectorObj.udpPort = obj.UdpPort
	sflowCollectorObj.adminState = obj.AdminState

	//Add sflow collector info to DB
	srvr.dbMutex.Lock()
	srvr.sflowCollectorDB[obj.IpAddr] = sflowCollectorObj
	srvr.dbMutex.Unlock()

	//Handle post processing due to adding a new sflow collector

	//Update key cache to use for getbulk responses
	srvr.sflowCollectorDBKeyCache = append(srvr.sflowCollectorDBKeyCache, obj.IpAddr)
}

func (srvr *sflowServer) ValidateUpdateSflowCollector(oldObj, newObj *objects.SflowCollector, attrset []bool) (bool, error) {
	var ok bool
	var err error
	if (newObj.UdpPort < 0) || (newObj.UdpPort > MAX_UDP_PORT) {
		err = errors.New("Update SflowCollector failed. Invalid UdpPort value provided")
	} else if (newObj.AdminState != objects.ADMIN_STATE_UP) && (newObj.AdminState != objects.ADMIN_STATE_DOWN) {
		err = errors.New("Update SflowCollector failed. Invalid AdminState value provided")
	} else {
		ok = true
	}
	return ok, err
}

func genSflowCollectorUpdateMask(attrset []bool) uint8 {
	var mask uint8
	for idx, val := range attrset {
		if val {
			switch idx {
			case objects.SFLOW_COLLECTOR_ATTR_UDP_PORT_IDX:
				mask |= objects.SFLOW_COLLECTOR_UPDATE_ATTR_UDP_PORT
			case objects.SFLOW_COLLECTOR_ATTR_ADMIN_STATE_IDX:
				mask |= objects.SFLOW_COLLECTOR_UPDATE_ATTR_ADMIN_STATE
			}
		}
	}
	return mask
}

func (srvr *sflowServer) updateSflowCollector(oldObj, newObj *objects.SflowCollector, attrset []bool) {
	mask := genSflowCollectorUpdateMask(attrset)
	if (mask & objects.SFLOW_COLLECTOR_UPDATE_ATTR_UDP_PORT) == objects.SFLOW_COLLECTOR_UPDATE_ATTR_UDP_PORT {
	}
	if (mask & objects.SFLOW_COLLECTOR_UPDATE_ATTR_ADMIN_STATE) == objects.SFLOW_COLLECTOR_UPDATE_ATTR_ADMIN_STATE {
	}
	srvr.dbMutex.Lock()
	obj := srvr.sflowCollectorDB[newObj.IpAddr]
	obj.udpPort = newObj.UdpPort
	obj.adminState = newObj.AdminState
	srvr.dbMutex.Unlock()
}

func (srvr *sflowServer) ValidateDeleteSflowCollector(obj *objects.SflowCollector) (bool, error) {
	var ok bool
	var err error

	if _, ok = srvr.sflowCollectorDB[obj.IpAddr]; !ok {
		err = errors.New("Delete SflowCollector failed. Collector configuration does not exist")
	} else {
		ok = true
	}
	return ok, err
}

func (srvr *sflowServer) deleteSflowCollector(obj *objects.SflowCollector) {
	//Gracefully shutdown tx to collector

	//Cleanup server state for this collector
	srvr.dbMutex.Lock()
	srvr.sflowCollectorDB[obj.IpAddr] = nil
	delete(srvr.sflowCollectorDB, obj.IpAddr)
	srvr.dbMutex.Unlock()
	//Remove key from keycache used for getbulk
	for idx, key := range srvr.sflowCollectorDBKeyCache {
		if obj.IpAddr == key {
			//Delete element
			srvr.sflowCollectorDBKeyCache = append(srvr.sflowCollectorDBKeyCache[:idx], srvr.sflowCollectorDBKeyCache[idx+1:]...)
		}
	}
}

//Sflow collector state object functions
func (srvr *sflowServer) getSflowCollectorState(ipAddr string) (*objects.SflowCollectorState, error) {
	var obj *objects.SflowCollectorState
	var err error

	srvr.dbMutex.RLock()
	defer srvr.dbMutex.RUnlock()
	if val, ok := srvr.sflowCollectorDB[ipAddr]; !ok {
		err = errors.New("Get SflowCollectorState failed. Invalid IpAddr key provided")
	} else {
		obj = &objects.SflowCollectorState{
			IpAddr:                  val.ipAddr,
			OperState:               val.operstate,
			NumSflowSamplesExported: val.numSflowSamplesExported,
			NumDatagramExported:     val.numDatagramExported,
		}
	}
	return obj, err
}

func (srvr *sflowServer) getBulkSflowCollectorState(fromIdx, count int) (*objects.SflowCollectorStateGetInfo, error) {
	var objInfo objects.SflowCollectorStateGetInfo

	if (fromIdx < 0) || (count < 0) ||
		(fromIdx > (len(srvr.sflowCollectorDBKeyCache) - 1)) {
		return nil, errors.New("Invalid arguments passed to GetBulkSflowCollectorState")
	}
	srvr.dbMutex.RLock()
	defer srvr.dbMutex.RUnlock()
	for idx, key := range srvr.sflowCollectorDBKeyCache {
		obj := srvr.sflowCollectorDB[key]
		if idx >= fromIdx {
			objInfo.Count += 1
			objInfo.List = append(objInfo.List,
				&objects.SflowCollectorState{
					IpAddr:                  obj.ipAddr,
					OperState:               obj.operstate,
					NumSflowSamplesExported: obj.numSflowSamplesExported,
					NumDatagramExported:     obj.numDatagramExported,
				})
			if objInfo.Count == count {
				objInfo.More = true
				objInfo.EndIdx = idx + 1
				break
			}
		}
	}
	return &objInfo, nil
}
