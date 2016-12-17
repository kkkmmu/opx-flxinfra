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

package rpc

import (
	"infra/statsd/api"
	"statsd"
)

func (rpcHdl *rpcServiceHandler) CreateSflowGlobal(obj *statsd.SflowGlobal) (bool, error) {
	convObj := convertFromRPCFmtSflowGlobal(obj)
	ok, err := api.CreateSflowGlobal(convObj)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) UpdateSflowGlobal(oldObj, newObj *statsd.SflowGlobal, attrset []bool, patchOpInfo []*statsd.PatchOpInfo) (bool, error) {
	convOldObj := convertFromRPCFmtSflowGlobal(oldObj)
	convNewObj := convertFromRPCFmtSflowGlobal(newObj)
	ok, err := api.UpdateSflowGlobal(convOldObj, convNewObj, attrset)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) DeleteSflowGlobal(obj *statsd.SflowGlobal) (bool, error) {
	convObj := convertFromRPCFmtSflowGlobal(obj)
	ok, err := api.DeleteSflowGlobal(convObj)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) CreateSflowCollector(obj *statsd.SflowCollector) (bool, error) {
	convObj := convertFromRPCFmtSflowCollector(obj)
	ok, err := api.CreateSflowCollector(convObj)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) UpdateSflowCollector(oldObj, newObj *statsd.SflowCollector, attrset []bool, patchOpInfo []*statsd.PatchOpInfo) (bool, error) {
	convOldObj := convertFromRPCFmtSflowCollector(oldObj)
	convNewObj := convertFromRPCFmtSflowCollector(newObj)
	ok, err := api.UpdateSflowCollector(convOldObj, convNewObj, attrset)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) DeleteSflowCollector(obj *statsd.SflowCollector) (bool, error) {
	convObj := convertFromRPCFmtSflowCollector(obj)
	ok, err := api.DeleteSflowCollector(convObj)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) GetSflowCollectorState(ipAddr string) (*statsd.SflowCollectorState, error) {
	var convObj *statsd.SflowCollectorState
	obj, err := api.GetSflowCollectorState(ipAddr)
	if err == nil {
		convObj = convertToRPCFmtSflowCollectorState(obj)
	}
	return convObj, err
}

func (rpcHdl *rpcServiceHandler) GetBulkSflowCollectorState(fromIdx, count statsd.Int) (*statsd.SflowCollectorStateGetInfo, error) {
	var bulkInfo statsd.SflowCollectorStateGetInfo
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
	return &bulkInfo, err
}

func (rpcHdl *rpcServiceHandler) CreateSflowIntf(obj *statsd.SflowIntf) (bool, error) {
	convObj := convertFromRPCFmtSflowIntf(obj)
	ok, err := api.CreateSflowIntf(convObj)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) UpdateSflowIntf(oldObj, newObj *statsd.SflowIntf, attrset []bool, patchOpInfo []*statsd.PatchOpInfo) (bool, error) {
	convOldObj := convertFromRPCFmtSflowIntf(oldObj)
	convNewObj := convertFromRPCFmtSflowIntf(newObj)
	ok, err := api.UpdateSflowIntf(convOldObj, convNewObj, attrset)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) DeleteSflowIntf(obj *statsd.SflowIntf) (bool, error) {
	convObj := convertFromRPCFmtSflowIntf(obj)
	ok, err := api.DeleteSflowIntf(convObj)
	return ok, err
}

func (rpcHdl *rpcServiceHandler) GetSflowIntfState(intfRef string) (*statsd.SflowIntfState, error) {
	var convObj *statsd.SflowIntfState
	obj, err := api.GetSflowIntfState(intfRef)
	if err == nil {
		convObj = convertToRPCFmtSflowIntfState(obj)
	}
	return convObj, err
}

func (rpcHdl *rpcServiceHandler) GetBulkSflowIntfState(fromIdx, count statsd.Int) (*statsd.SflowIntfStateGetInfo, error) {
	var bulkInfo statsd.SflowIntfStateGetInfo
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
	return &bulkInfo, err
}
