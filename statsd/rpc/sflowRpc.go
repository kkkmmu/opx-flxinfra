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

package rpc

import (
	"infra/statsd/api"
	"models/objects"
	"statsd"
)

func (rpcHdl *rpcServiceHandler) CreateSflowGlobal(obj *statsd.SflowGlobal) (bool, error) {
	rpcHdl.logger.Debug("CreateSflowGlobal received : ", *obj)
	convObj := convertFromRPCFmtSflowGlobal(obj)
	ok, err := api.CreateSflowGlobal(convObj)
	rpcHdl.logger.Debug("CreateSflowGlobal returning : ", ok, err)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) UpdateSflowGlobal(oldObj, newObj *statsd.SflowGlobal, attrset []bool, patchOpInfo []*statsd.PatchOpInfo) (bool, error) {
	rpcHdl.logger.Debug("UpdateSflowGlobal received : ", *oldObj, *newObj, attrset)
	convOldObj := convertFromRPCFmtSflowGlobal(oldObj)
	convNewObj := convertFromRPCFmtSflowGlobal(newObj)
	ok, err := api.UpdateSflowGlobal(convOldObj, convNewObj, attrset)
	rpcHdl.logger.Debug("UpdateSflowGlobal returning : ", ok, err)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) DeleteSflowGlobal(obj *statsd.SflowGlobal) (bool, error) {
	rpcHdl.logger.Debug("DeleteSflowGlobal received : ", *obj)
	convObj := convertFromRPCFmtSflowGlobal(obj)
	ok, err := api.DeleteSflowGlobal(convObj)
	rpcHdl.logger.Debug("DeleteSflowGlobal returning : ", ok, err)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) CreateSflowCollector(obj *statsd.SflowCollector) (bool, error) {
	rpcHdl.logger.Debug("CreateSflowCollector received : ", *obj)
	convObj := convertFromRPCFmtSflowCollector(obj)
	ok, err := api.CreateSflowCollector(convObj)
	rpcHdl.logger.Debug("CreateSflowCollector returning : ", ok, err)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) UpdateSflowCollector(oldObj, newObj *statsd.SflowCollector, attrset []bool, patchOpInfo []*statsd.PatchOpInfo) (bool, error) {
	rpcHdl.logger.Debug("UpdateSflowCollector received : ", *oldObj, *newObj, attrset)
	convOldObj := convertFromRPCFmtSflowCollector(oldObj)
	convNewObj := convertFromRPCFmtSflowCollector(newObj)
	ok, err := api.UpdateSflowCollector(convOldObj, convNewObj, attrset)
	rpcHdl.logger.Debug("UpdateSflowCollector returning : ", ok, err)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) DeleteSflowCollector(obj *statsd.SflowCollector) (bool, error) {
	rpcHdl.logger.Debug("DeleteSflowCollector received : ", *obj)
	convObj := convertFromRPCFmtSflowCollector(obj)
	ok, err := api.DeleteSflowCollector(convObj)
	rpcHdl.logger.Debug("DeleteSflowCollector returning : ", ok, err)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) GetSflowCollectorState(ipAddr string) (*statsd.SflowCollectorState, error) {
	var convObj *statsd.SflowCollectorState
	rpcHdl.logger.Debug("GetSflowCollectorState received : ", ipAddr)
	obj, err := api.GetSflowCollectorState(ipAddr)
	if err == nil {
		convObj = convertToRPCFmtSflowCollectorState(obj)
	}
	rpcHdl.logger.Debug("GetSflowCollectorState returning : ", *convObj, err)
	return convObj, err
}

func (rpcHdl *rpcServiceHandler) GetBulkSflowCollectorState(fromIdx, count statsd.Int) (*statsd.SflowCollectorStateGetInfo, error) {
	var bulkInfo statsd.SflowCollectorStateGetInfo
	rpcHdl.logger.Debug("GetBulkSflowCollectorState received : ", fromIdx, count)
	info, err := api.GetBulkSflowCollectorState(int(fromIdx), int(count))
	if err == nil {
		bulkInfo.StartIdx = fromIdx
		bulkInfo.EndIdx = statsd.Int(info.EndIdx)
		bulkInfo.More = info.More
		bulkInfo.Count = statsd.Int(len(info.List))
		for _, obj := range info.List {
			bulkInfo.SflowCollectorStateList = append(bulkInfo.SflowCollectorStateList,
				convertToRPCFmtSflowCollectorState(obj))
		}
	}
	rpcHdl.logger.Debug("GetBulkSflowCollectorState returning : ", bulkInfo, err)
	return &bulkInfo, err
}

func (rpcHdl *rpcServiceHandler) CreateSflowIntf(obj *statsd.SflowIntf) (bool, error) {
	rpcHdl.logger.Debug("CreateSflowIntf received : ", *obj)
	convObj := convertFromRPCFmtSflowIntf(obj)
	ok, err := api.CreateSflowIntf(convObj)
	rpcHdl.logger.Debug("CreateSflowIntf returning : ", ok, err)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) UpdateSflowIntf(oldObj, newObj *statsd.SflowIntf, attrset []bool, patchOpInfo []*statsd.PatchOpInfo) (bool, error) {
	rpcHdl.logger.Debug("UpdateSflowIntf received : ", *oldObj, *newObj, attrset)
	convOldObj := convertFromRPCFmtSflowIntf(oldObj)
	convNewObj := convertFromRPCFmtSflowIntf(newObj)
	ok, err := api.UpdateSflowIntf(convOldObj, convNewObj, attrset)
	rpcHdl.logger.Debug("UpdateSflowIntf returning : ", ok, err)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) DeleteSflowIntf(obj *statsd.SflowIntf) (bool, error) {
	rpcHdl.logger.Debug("DeleteSflowIntf received : ", *obj)
	convObj := convertFromRPCFmtSflowIntf(obj)
	ok, err := api.DeleteSflowIntf(convObj)
	rpcHdl.logger.Debug("DeleteSflowIntf returning : ", ok, err)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) GetSflowIntfState(intfRef string) (*statsd.SflowIntfState, error) {
	var convObj *statsd.SflowIntfState
	rpcHdl.logger.Debug("GetSflowIntfState received : ", intfRef)
	obj, err := api.GetSflowIntfState(intfRef)
	if err == nil {
		convObj = convertToRPCFmtSflowIntfState(obj)
	}
	rpcHdl.logger.Debug("GetSflowIntfState returning : ", *convObj, err)
	return convObj, err
}

func (rpcHdl *rpcServiceHandler) GetBulkSflowIntfState(fromIdx, count statsd.Int) (*statsd.SflowIntfStateGetInfo, error) {
	var bulkInfo statsd.SflowIntfStateGetInfo
	rpcHdl.logger.Debug("GetBulkSflowIntfState received : ", fromIdx, count)
	info, err := api.GetBulkSflowIntfState(int(fromIdx), int(count))
	if err == nil {
		bulkInfo.StartIdx = fromIdx
		bulkInfo.EndIdx = statsd.Int(info.EndIdx)
		bulkInfo.More = info.More
		bulkInfo.Count = statsd.Int(len(info.List))
		for _, obj := range info.List {
			bulkInfo.SflowIntfStateList = append(bulkInfo.SflowIntfStateList,
				convertToRPCFmtSflowIntfState(obj))
		}
	}
	rpcHdl.logger.Debug("GetBulkSflowIntfState returning : ", bulkInfo, err)
	return &bulkInfo, err
}

func (rpcHdl *rpcServiceHandler) replayCfgFromDB() {
	rpcHdl.logger.Debug("Replaying configuration from DB started")

	//Replay SflowGlobal info
	sflowGblList, err := rpcHdl.dbHdl.GetAllObjFromDb(objects.SflowGlobal{})
	if err != nil {
		rpcHdl.logger.Err("Error retrieving SflowGlobal configuration from DB")
		//Return here as intf/collector config will fail
		return
	}
	for _, val := range sflowGblList {
		dbObj := val.(objects.SflowGlobal)
		obj := new(statsd.SflowGlobal)
		objects.ConvertstatsdSflowGlobalObjToThrift(&dbObj, obj)
		rpcHdl.CreateSflowGlobal(obj)
	}

	//Replay SflowCollector info
	sflowCollectorList, err := rpcHdl.dbHdl.GetAllObjFromDb(objects.SflowCollector{})
	if err != nil {
		rpcHdl.logger.Err("Error retrieving SflowCollector configuration from DB")
	} else {
		for _, val := range sflowCollectorList {
			dbObj := val.(objects.SflowCollector)
			obj := new(statsd.SflowCollector)
			objects.ConvertstatsdSflowCollectorObjToThrift(&dbObj, obj)
			rpcHdl.CreateSflowCollector(obj)
		}
	}

	//Replay SflowIntf info
	sflowIntfList, err := rpcHdl.dbHdl.GetAllObjFromDb(objects.SflowIntf{})
	if err != nil {
		rpcHdl.logger.Err("Error retrieving SflowIntf configuration from DB")
	} else {
		for _, val := range sflowIntfList {
			dbObj := val.(objects.SflowIntf)
			obj := new(statsd.SflowIntf)
			objects.ConvertstatsdSflowIntfObjToThrift(&dbObj, obj)
			rpcHdl.CreateSflowIntf(obj)
		}
	}

	rpcHdl.logger.Debug("Replaying configuration from DB completed")
}
