//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.
//
// _______  __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----` \   \/    \/   /  |  | `---|  |----`|  ,----'|  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |     |  |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |     |  `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//

package api

import (
	"errors"
	"infra/statsd/objects"
	"infra/statsd/server"
	"utils/logging"
)

var svr *server.DmnServer
var logger logging.LoggerIntf

//Initialize server handle
func InitApiLayer(server *server.DmnServer, log logging.LoggerIntf) {
	svr = server
	logger = log
}

func CreateSflowGlobal(obj *objects.SflowGlobal) (bool, error) {
	ok, err := svr.ValidateCreateSflowGlobal(obj)
	logger.Debug("ValidateCreateSflowGlobal returned :", ok, err)
	if ok {
		svr.ReqChan <- &server.ServerRequest{
			Op: server.CREATE_SFLOW_GLOBAL,
			Data: interface{}(&server.CreateSflowGlobalInArgs{
				Obj: obj,
			}),
		}
	}
	return ok, err
}

func UpdateSflowGlobal(oldObj, newObj *objects.SflowGlobal, attrset []bool) (bool, error) {
	ok, err := svr.ValidateUpdateSflowGlobal(oldObj, newObj, attrset)
	logger.Debug("ValidateUpdateSflowGlobal returned :", ok, err)
	if ok {
		svr.ReqChan <- &server.ServerRequest{
			Op: server.UPDATE_SFLOW_GLOBAL,
			Data: interface{}(&server.UpdateSflowGlobalInArgs{
				OldObj:  oldObj,
				NewObj:  newObj,
				AttrSet: attrset,
			}),
		}
	}
	return ok, err
}

func DeleteSflowGlobal(obj *objects.SflowGlobal) (bool, error) {
	ok, err := svr.ValidateDeleteSflowGlobal(obj)
	logger.Debug("ValidateDeleteSflowGlobal returned :", ok, err)
	if ok {
		svr.ReqChan <- &server.ServerRequest{
			Op: server.DELETE_SFLOW_GLOBAL,
			Data: interface{}(&server.DeleteSflowGlobalInArgs{
				Obj: obj,
			}),
		}
	}
	return ok, err
}

func CreateSflowCollector(obj *objects.SflowCollector) (bool, error) {
	ok, err := svr.ValidateCreateSflowCollector(obj)
	logger.Debug("ValidateCreateSflowCollector returned :", ok, err)
	if ok {
		svr.ReqChan <- &server.ServerRequest{
			Op: server.CREATE_SFLOW_COLLECTOR,
			Data: interface{}(&server.CreateSflowCollectorInArgs{
				Obj: obj,
			}),
		}
	}
	return ok, err
}

func UpdateSflowCollector(oldObj, newObj *objects.SflowCollector, attrset []bool) (bool, error) {
	ok, err := svr.ValidateUpdateSflowCollector(oldObj, newObj, attrset)
	logger.Debug("ValidateUpdateSflowCollector returned :", ok, err)
	if ok {
		svr.ReqChan <- &server.ServerRequest{
			Op: server.UPDATE_SFLOW_COLLECTOR,
			Data: interface{}(&server.UpdateSflowCollectorInArgs{
				OldObj:  oldObj,
				NewObj:  newObj,
				AttrSet: attrset,
			}),
		}
	}
	return ok, err
}

func DeleteSflowCollector(obj *objects.SflowCollector) (bool, error) {
	ok, err := svr.ValidateDeleteSflowCollector(obj)
	logger.Debug("ValidateDeleteSflowCollector returned :", ok, err)
	if ok {
		svr.ReqChan <- &server.ServerRequest{
			Op: server.DELETE_SFLOW_COLLECTOR,
			Data: interface{}(&server.DeleteSflowCollectorInArgs{
				Obj: obj,
			}),
		}
	}
	return ok, err
}

func GetSflowCollectorState(ipAddr string) (*objects.SflowCollectorState, error) {
	svr.ReqChan <- &server.ServerRequest{
		Op: server.GET_SFLOW_COLLECTOR_STATE,
		Data: interface{}(&server.GetSflowCollectorStateInArgs{
			IpAddr: ipAddr,
		}),
	}
	ret := <-svr.ReplyChan
	if retObj, ok := ret.(*server.GetSflowCollectorStateOutArgs); ok {
		return retObj.Obj, retObj.Err
	} else {
		return nil, errors.New("Error: Invalid response from server during GetSflowCollectorState")
	}
}

func GetBulkSflowCollectorState(fromIdx, count int) (*objects.SflowCollectorStateGetInfo, error) {
	svr.ReqChan <- &server.ServerRequest{
		Op: server.GET_BULK_SFLOW_COLLECTOR_STATE,
		Data: interface{}(&server.GetBulkInArgs{
			FromIdx: fromIdx,
			Count:   count,
		}),
	}
	ret := <-svr.ReplyChan
	if retObj, ok := ret.(*server.GetBulkSflowCollectorStateOutArgs); ok {
		return retObj.BulkObj, retObj.Err
	} else {
		return nil, errors.New("Error: Invalid response from server during GetBulkSflowCollectorState")
	}
}

func CreateSflowIntf(obj *objects.SflowIntf) (bool, error) {
	ok, err := svr.ValidateCreateSflowIntf(obj)
	logger.Debug("ValidateCreateSflowIntf returned :", ok, err)
	if ok {
		svr.ReqChan <- &server.ServerRequest{
			Op: server.CREATE_SFLOW_INTF,
			Data: interface{}(&server.CreateSflowIntfInArgs{
				Obj: obj,
			}),
		}
	}
	return ok, err
}

func UpdateSflowIntf(oldObj, newObj *objects.SflowIntf, attrset []bool) (bool, error) {
	ok, err := svr.ValidateUpdateSflowIntf(oldObj, newObj, attrset)
	logger.Debug("ValidateUpdateSflowIntf returned :", ok, err)
	if ok {
		svr.ReqChan <- &server.ServerRequest{
			Op: server.UPDATE_SFLOW_INTF,
			Data: interface{}(&server.UpdateSflowIntfInArgs{
				OldObj:  oldObj,
				NewObj:  newObj,
				AttrSet: attrset,
			}),
		}
	}
	return ok, err
}

func DeleteSflowIntf(obj *objects.SflowIntf) (bool, error) {
	ok, err := svr.ValidateDeleteSflowIntf(obj)
	logger.Debug("ValidateDeleteSflowIntf returned :", ok, err)
	if ok {
		svr.ReqChan <- &server.ServerRequest{
			Op: server.DELETE_SFLOW_INTF,
			Data: interface{}(&server.DeleteSflowIntfInArgs{
				Obj: obj,
			}),
		}
	}
	return ok, err
}

func GetSflowIntfState(intfRef string) (*objects.SflowIntfState, error) {
	svr.ReqChan <- &server.ServerRequest{
		Op: server.GET_SFLOW_INTF_STATE,
		Data: interface{}(&server.GetSflowIntfStateInArgs{
			IntfRef: intfRef,
		}),
	}
	ret := <-svr.ReplyChan
	if retObj, ok := ret.(*server.GetSflowIntfStateOutArgs); ok {
		return retObj.Obj, retObj.Err
	} else {
		return nil, errors.New("Error: Invalid response from server during GetSflowIntfState")
	}
}

func GetBulkSflowIntfState(fromIdx, count int) (*objects.SflowIntfStateGetInfo, error) {
	svr.ReqChan <- &server.ServerRequest{
		Op: server.GET_BULK_SFLOW_INTF_STATE,
		Data: interface{}(&server.GetBulkInArgs{
			FromIdx: fromIdx,
			Count:   count,
		}),
	}
	ret := <-svr.ReplyChan
	if retObj, ok := ret.(*server.GetBulkSflowIntfStateOutArgs); ok {
		return retObj.BulkObj, retObj.Err
	} else {
		return nil, errors.New("Error: Invalid response from server during GetBulkSflowIntfState")
	}
}
