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
)

func newSflowIntf() *sflowIntf {
	obj := new(sflowIntf)
	obj.shutdownCh = make(chan bool)
	return obj
}

func (srvr *sflowServer) getIfIndexFromIntfRef(intfRef string) int32 {
	var ifIndex int32
	var val netDevData
	var intfExists bool

	for ifIndex, val = range srvr.netDevInfo {
		if val.intfRef == intfRef {
			intfExists = true
			break
		}
	}
	if !intfExists {
		ifIndex = INVALID_IFINDEX
	}
	return ifIndex
}

//Sflow interface related functions
func (srvr *sflowServer) ValidateCreateSflowIntf(obj *objects.SflowIntf) (bool, error) {
	var ok bool
	var err error

	ifIndex := srvr.getIfIndexFromIntfRef(obj.IntfRef)
	if !isIfIndexValid(ifIndex) {
		err = errors.New("Create SflowIntf failed. Invalid IntfRef value provided")
	} else if (obj.AdminState != objects.ADMIN_STATE_UP) && (obj.AdminState != objects.ADMIN_STATE_DOWN) {
		err = errors.New("Create SflowIntf failed. Invalid AdminState value provided")
	} else if obj.SamplingRate < 0 {
		err = errors.New("Create SflowIntf failed. Invalid SamplingRate value provided")
	} else {
		ok = true
	}
	return ok, err
}

func (srvr *sflowServer) createSflowIntf(obj *objects.SflowIntf) {
	ifIndex := srvr.getIfIndexFromIntfRef(obj.IntfRef)

	sflowIntfObj := newSflowIntf()
	sflowIntfObj.ifIndex = ifIndex
	sflowIntfObj.adminState = obj.AdminState
	sflowIntfObj.samplingRate = obj.SamplingRate

	//Add sflow intf info into DB
	srvr.dbMutex.Lock()
	srvr.sflowIntfDB[ifIndex] = sflowIntfObj
	srvr.dbMutex.Unlock()

	if obj.AdminState == objects.ADMIN_STATE_UP {
		//Start poller for interface
		go sflowIntfObj.intfPoller()

		//Set operstate to UP
		sflowIntfObj.operstate = ADMIN_STATE_UP

		//Pass config to asicd to enable sflow cfg on interface
	}

	//Update key cache to use for getbulk responses
	srvr.sflowIntfDBKeyCache = append(srvr.sflowIntfDBKeyCache, ifIndex)
}

func (srvr *sflowServer) ValidateUpdateSflowIntf(oldObj, newObj *objects.SflowIntf, attrset []bool) (bool, error) {
	var ok bool
	var err error
	if (newObj.AdminState != objects.ADMIN_STATE_UP) && (newObj.AdminState != objects.ADMIN_STATE_DOWN) {
		err = errors.New("Update SflowIntf failed. Invalid AdminState value provided")
	} else if newObj.SamplingRate < 0 {
		err = errors.New("Update SflowIntf failed. Invalid SamplingRate value provided")
	} else {
		ok = true
	}
	return ok, err
}

func genSflowIntfUpdateMask(attrset []bool) uint8 {
	var mask uint8
	for idx, val := range attrset {
		if val {
			switch idx {
			case objects.SFLOW_INTF_ATTR_ADMIN_STATE_IDX:
				mask |= objects.SFLOW_INTF_UPDATE_ATTR_ADMIN_STATE
			case objects.SFLOW_INTF_ATTR_SAMPLING_RATE_IDX:
				mask |= objects.SFLOW_INTF_UPDATE_ATTR_SAMPLING_RATE
			}
		}
	}
	return mask
}

func (srvr *sflowServer) updateSflowIntf(oldObj, newObj *objects.SflowIntf, attrset []bool) {
	//Determine ifIndex
	ifIndex := srvr.getIfIndexFromIntfRef(newObj.IntfRef)
	srvr.dbMutex.Lock()
	obj := srvr.sflowIntfDB[ifIndex]
	obj.adminState = newObj.AdminState
	obj.samplingRate = newObj.SamplingRate
	srvr.dbMutex.Unlock()

	mask := genSflowIntfUpdateMask(attrset)
	if (mask & objects.SFLOW_INTF_UPDATE_ATTR_ADMIN_STATE) == objects.SFLOW_INTF_UPDATE_ATTR_ADMIN_STATE {
		if newObj.AdminState == objects.ADMIN_STATE_UP {
			//Pass configuration down to asicd

			//Start poller for interface
			go obj.intfPoller()
		} else {
			//Pass configuration down to asicd

			//Terminate interface polling thread
			obj.shutdownCh <- true
		}
	}
	if (mask & objects.SFLOW_INTF_UPDATE_ATTR_SAMPLING_RATE) == objects.SFLOW_INTF_UPDATE_ATTR_SAMPLING_RATE {
		//Pass configuration down to asicd
	}
}

func (srvr *sflowServer) ValidateDeleteSflowIntf(obj *objects.SflowIntf) (bool, error) {
	var ok bool
	var err error

	//Determine ifIndex
	ifIndex := srvr.getIfIndexFromIntfRef(obj.IntfRef)
	if !isIfIndexValid(ifIndex) {
		err = errors.New("Delete SflowIntf failed. Invalid IntfRef value provided")
	} else {
		ok = true
	}
	return ok, err
}

func (srvr *sflowServer) deleteSflowIntf(obj *objects.SflowIntf) {
	//Determine ifIndex
	ifIndex := srvr.getIfIndexFromIntfRef(obj.IntfRef)

	//Cleanup server state for this interface
	srvr.dbMutex.Lock()
	obj := srvr.sflowIntfDB[ifIndex]
	srvr.sflowIntfDB[ifIndex] = nil
	delete(srvr.sflowIntfDB, ifIndex)
	srvr.dbMutex.Unlock()

	//Gracefully shutdown interface Rx routine
	if obj.adminState == objects.ADMIN_STATE_UP {
		obj.shutdownCh <- true
	}

	//Remove key from keycache usede gor getbulk
	for idx, key := range srvr.sflowIntfDBKeyCache {
		if ifIndex == key {
			//Delete element
			srvr.sflowIntfDBKeyCache = append(srvr.sflowIntfDBKeyCache[:idx], srvr.sflowIntfDBKeyCache[idx+1:]...)
		}
	}
}

//Sflow interface state object functions
func (srvr *sflowServer) getSflowIntfState(intfRef string) (*objects.SflowIntfState, error) {
	var obj *objects.SflowIntfState
	var err error

	//Determine ifIndex
	ifIndex := srvr.getIfIndexFromIntfRef(intfRef)
	srvr.dbMutex.RLock()
	defer srvr.dbMutex.RUnlock()
	if val, ok := srvr.sflowIntfDB[ifIndex]; !ok {
		err = errors.New("Get SflowIntfState failed. Invalid IntfRef key provided")
	} else {
		obj = &objects.SflowIntfState{
			IntfRef:                 intfRef,
			OperState:               val.operstate,
			NumSflowSamplesExported: val.numSflowSamplesExported,
		}
	}
	return obj, err
}

func (srvr *sflowServer) getBulkSflowIntfState(fromIdx, count int) (*objects.SflowIntfStateGetInfo, error) {
	var objInfo objects.SflowIntfStateGetInfo

	if (fromIdx < 0) || (count < 0) ||
		(fromIdx > (len(srvr.sflowIntfDBKeyCache) - 1)) {
		return nil, errors.New("Invalid arguments passed to GetBulkSflowIntfState")
	}
	srvr.dbMutex.RLock()
	defer srvr.dbMutex.RUnlock()
	for idx, key := range srvr.sflowIntfDBKeyCache {
		obj := srvr.sflowIntfDB[key]
		if idx >= fromIdx {
			objInfo.Count += 1
			//Determine intfRef
			intfRef := srvr.netDevInfo[key].intfRef
			objInfo.List = append(objInfo.List,
				&objects.SflowIntfState{
					IntfRef:                 intfRef,
					OperState:               obj.operState,
					NumSflowSamplesExported: obj.numSflowSamplesExported,
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
