//
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
	//"fmt"
	"infra/platformd/api"
	"platformd"
)

func (rpcHdl *rpcServiceHandler) CreateQsfp(config *platformd.Qsfp) (bool, error) {
	return false, errors.New("Not supported")
}

func (rpcHdl *rpcServiceHandler) DeleteQsfp(config *platformd.Qsfp) (bool, error) {
	return false, errors.New("Not supported")
}

func (rpcHdl *rpcServiceHandler) UpdateQsfp(oldConfig *platformd.Qsfp, newConfig *platformd.Qsfp, attrset []bool, op []*platformd.PatchOpInfo) (bool, error) {
	oldCfg := convertRPCToObjFmtQsfpConfig(oldConfig)
	newCfg := convertRPCToObjFmtQsfpConfig(newConfig)
	rv, err := api.UpdateQsfp(oldCfg, newCfg, attrset)
	return rv, err
}

func (rpcHdl *rpcServiceHandler) GetQsfp(QsfpId int32) (*platformd.Qsfp, error) {
	return nil, errors.New("Not supported")
}

func (rpcHdl *rpcServiceHandler) GetBulkQsfp(fromIdx, count platformd.Int) (*platformd.QsfpGetInfo, error) {
	var getBulkObj platformd.QsfpGetInfo
	var err error

	info, err := api.GetBulkQsfpConfig(int(fromIdx), int(count))
	if err != nil {
		return nil, err
	}
	getBulkObj.StartIdx = fromIdx
	getBulkObj.EndIdx = platformd.Int(info.EndIdx)
	getBulkObj.More = info.More
	getBulkObj.Count = platformd.Int(len(info.List))
	for idx := 0; idx < len(info.List); idx++ {
		getBulkObj.QsfpList = append(getBulkObj.QsfpList, convertToRPCFmtQsfpConfig(info.List[idx]))
	}
	return &getBulkObj, err
}

func (rpcHdl *rpcServiceHandler) GetQsfpState(QsfpId int32) (*platformd.QsfpState, error) {
	var rpcObj *platformd.QsfpState
	var err error

	obj, err := api.GetQsfpState(QsfpId)
	if err == nil {
		rpcObj = convertToRPCFmtQsfpState(obj)
	}
	return rpcObj, err
}

func (rpcHdl *rpcServiceHandler) GetBulkQsfpState(fromIdx, count platformd.Int) (*platformd.QsfpStateGetInfo, error) {
	var getBulkObj platformd.QsfpStateGetInfo
	var err error

	info, err := api.GetBulkQsfpState(int(fromIdx), int(count))
	if err != nil {
		return nil, err
	}
	getBulkObj.StartIdx = fromIdx
	getBulkObj.EndIdx = platformd.Int(info.EndIdx)
	getBulkObj.More = info.More
	getBulkObj.Count = platformd.Int(len(info.List))
	for idx := 0; idx < len(info.List); idx++ {
		getBulkObj.QsfpStateList = append(getBulkObj.QsfpStateList, convertToRPCFmtQsfpState(info.List[idx]))
	}
	return &getBulkObj, err
}

func (rpcHdl *rpcServiceHandler) GetQsfpPMDataState(QspfId int32, Resource string, Class string) (*platformd.QsfpPMDataState, error) {
	return nil, nil
}

func (rpcHdl *rpcServiceHandler) GetBulkQsfpPMDataState(fromIdx, count platformd.Int) (*platformd.QsfpPMDataStateGetInfo, error) {
	return nil, errors.New("Not supported")
}
