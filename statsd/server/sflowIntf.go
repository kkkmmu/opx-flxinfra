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
)

func newSflowIntf() *sflowIntf {
	obj := new(sflowIntf)
	obj.shutdownCh = make(chan bool)
	obj.initCompleteCh = make(chan bool)
	obj.restartCtrPollTicker = make(chan int32)
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
	if !srvr.isGlobalObjCreated() {
		err = errors.New("Create SflowIntf failed. SflowGlobal object not created yet")
	} else if !isIfIndexValid(ifIndex) {
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

	/* Handle post processing due to adding a new sflow interface
	   Spawn interface polling routine only if both local and global
	   admin states are UP
	*/
	if srvr.sflowGblDB.adminState == objects.ADMIN_STATE_UP {
		if obj.AdminState == objects.ADMIN_STATE_UP {
			//Start poller for interface
			go sflowIntfObj.intfPoller(&intfPollerInitParams{
				netDevName:           srvr.netDevInfo[ifIndex].netDevName,
				intfRef:              srvr.netDevInfo[ifIndex].intfRef,
				sflowIntfRecordCh:    srvr.sflowIntfRecordCh,
				sflowIntfCtrRecordCh: srvr.sflowIntfCtrRecordCh,
				pollInterval:         srvr.sflowGblDB.counterPollInterval,
			})
			logger.Debug("Spawned interface poller go routine, pending on init complete")
			//Pend on init complete channel
			ok := <-sflowIntfObj.initCompleteCh
			if ok {
				//Pass config to asicd to enable sflow on interface
				//Enable sflow on this interface
				err := hwHdl.SflowEnable(ifIndex)
				if err != nil {
					logger.Err("Create SflowIntf failed. Hw call to enable Sflow sampling returned : ", err, "Terminating interface poller")
					//Terminate interface poller
					sflowIntfObj.shutdownIntfPoller()
				} else {
					//Set sampling rate for this interface
					err := hwHdl.SflowSetSamplingRate(ifIndex, obj.SamplingRate)
					if err != nil {
						logger.Err("Create SflowIntf failed. Hw call to enable Sflow sampling returned : ", err, "Terminating interface poller")
						//Terminate interface poller
						sflowIntfObj.shutdownIntfPoller()
					}
				}
			}
		}
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

func (srvr *sflowServer) updateSflowIntfAdminState(obj *sflowIntf, ifIndex int32, adminState string) {
	if adminState == objects.ADMIN_STATE_UP {
		//Start poller for interface
		go obj.intfPoller(&intfPollerInitParams{
			netDevName:           srvr.netDevInfo[ifIndex].netDevName,
			intfRef:              srvr.netDevInfo[ifIndex].intfRef,
			sflowIntfRecordCh:    srvr.sflowIntfRecordCh,
			sflowIntfCtrRecordCh: srvr.sflowIntfCtrRecordCh,
			pollInterval:         srvr.sflowGblDB.counterPollInterval,
		})
		//Pend on init complete channel
		ok := <-obj.initCompleteCh
		if ok {
			//Pass configuration down to asicd
			//Enable sflow on this interface
			err := hwHdl.SflowEnable(ifIndex)
			if err != nil {
				logger.Err("Update SflowIntf failed. Hw call to enable Sflow sampling returned : ", err)
				//Terminate interface polling thread
				obj.shutdownIntfPoller()
			}
		}
	} else {
		if obj.operstate == objects.ADMIN_STATE_UP {
			//Pass configuration down to asicd
			//Disable sflow on this interface
			err := hwHdl.SflowDisable(ifIndex)
			if err != nil {
				logger.Err("Update SflowIntf failed. Hw call to disable Sflow sampling returned : ", err)
			} else {
				//Terminate interface polling thread
				obj.shutdownIntfPoller()
			}
		}
	}
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
		//Process admin state change only if global admin state is UP
		if srvr.sflowGblDB.adminState == objects.ADMIN_STATE_UP {
			srvr.updateSflowIntfAdminState(obj, ifIndex, newObj.AdminState)
		}
	}
	if (mask & objects.SFLOW_INTF_UPDATE_ATTR_SAMPLING_RATE) == objects.SFLOW_INTF_UPDATE_ATTR_SAMPLING_RATE {
		//Pass configuration down to asicd only if interface poller is alive
		if obj.operstate == objects.ADMIN_STATE_UP {
			//Set sampling rate for this interface
			err := hwHdl.SflowSetSamplingRate(ifIndex, newObj.SamplingRate)
			if err != nil {
				logger.Err("Update SflowIntf failed. Hw call to update Sflow sampling returned : ", err)
			}
		}
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
	sflowIntfObj := srvr.sflowIntfDB[ifIndex]
	srvr.sflowIntfDB[ifIndex] = nil
	delete(srvr.sflowIntfDB, ifIndex)
	srvr.dbMutex.Unlock()

	//Gracefully shutdown interface Rx routine
	if (sflowIntfObj.adminState == objects.ADMIN_STATE_UP) && (sflowIntfObj.operstate == objects.ADMIN_STATE_UP) {
		//Pass configuration down to asicd
		//Disable sflow on this interface
		err := hwHdl.SflowDisable(ifIndex)
		if err != nil {
			logger.Err("Failure during Delete SflowIntf failed. Hw call to disable Sflow sampling returned : ", err)
		}
		//Terminate interface polling thread
		sflowIntfObj.shutdownIntfPoller()
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
					OperState:               obj.operstate,
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
