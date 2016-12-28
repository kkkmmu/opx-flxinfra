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
	"infra/statsd/hw"
	"utils/keepalive"
)

func NewSTATSDServer(initParams *ServerInitParams) *DmnServer {
	logger = initParams.Logger
	srvr := DmnServer{}
	srvr.dbHdl = initParams.DbHdl
	srvr.paramsDir = initParams.ParamsDir
	srvr.InitCompleteCh = make(chan bool)
	srvr.ReqChan = make(chan *ServerRequest)
	srvr.ReplyChan = make(chan interface{})
	srvr.sflowServer = new(sflowServer)
	return &srvr
}

func (srvr *DmnServer) initServer() error {
	//Init base server
	srvr.initSflowServer()
	//Get hw handle
	clientsFile := srvr.paramsDir + "/clients.json"
	hwHdl = hw.GetHwClntHdl(clientsFile, logger)
	if hwHdl == nil {
		return errors.New("Failed to initalize hardware handle")
	}
	//Construct netdev infra
	err := srvr.constructSflowInfra()
	if err != nil {
		return errors.New("Failed to construct Sflow infrastructure : " + err.Error())
	}
	//Spawn sflow core routines
	go srvr.sflowServer.sflowCoreRx()
	go srvr.sflowServer.sflowCoreTx()
	return nil
}

func (srvr *DmnServer) Serve() {
	logger.Info("Server initialization started")
	err := srvr.initServer()
	if err != nil {
		logger.Err(err)
		initFailed = true
	}
	daemonStatusListener := keepalive.InitDaemonStatusListener()
	if daemonStatusListener != nil {
		go daemonStatusListener.StartDaemonStatusListner()
	}
	srvr.InitCompleteCh <- true
	logger.Info("Server initialization complete, starting cfg/state listerner")
	for {
		select {
		case req := <-srvr.ReqChan:
			srvr.processRequest(req)

		case daemonStatus := <-daemonStatusListener.DaemonStatusCh:
			logger.Info("Received daemon status: ", daemonStatus.Name, daemonStatus.Status)
		}
	}
}

func (srvr *DmnServer) processRequest(req *ServerRequest) {
	logger.Debug("Server request received : ", *req)
	switch req.Op {
	case CREATE_SFLOW_GLOBAL:
		if val, ok := req.Data.(*CreateSflowGlobalInArgs); ok {
			srvr.createSflowGlobal(val.Obj)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - CREATE_SFLOW_GLOBAL")
		}
	case UPDATE_SFLOW_GLOBAL:
		if val, ok := req.Data.(*UpdateSflowGlobalInArgs); ok {
			srvr.updateSflowGlobal(val.OldObj, val.NewObj, val.AttrSet)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - UPDATE_SFLOW_GLOBAL")
		}

	case DELETE_SFLOW_GLOBAL:
		if val, ok := req.Data.(*DeleteSflowGlobalInArgs); ok {
			srvr.deleteSflowGlobal(val.Obj)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - DELETE_SFLOW_GLOBAL")
		}

	case CREATE_SFLOW_COLLECTOR:
		if val, ok := req.Data.(*CreateSflowCollectorInArgs); ok {
			srvr.createSflowCollector(val.Obj)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - CREATE_SFLOW_COLLECTOR")
		}

	case UPDATE_SFLOW_COLLECTOR:
		if val, ok := req.Data.(*UpdateSflowCollectorInArgs); ok {
			srvr.updateSflowCollector(val.OldObj, val.NewObj, val.AttrSet)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - UPDATE_SFLOW_COLLECTOR")
		}

	case DELETE_SFLOW_COLLECTOR:
		if val, ok := req.Data.(*DeleteSflowCollectorInArgs); ok {
			srvr.deleteSflowCollector(val.Obj)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - DELETE_SFLOW_COLLECTOR")
		}

	case CREATE_SFLOW_INTF:
		if val, ok := req.Data.(*CreateSflowIntfInArgs); ok {
			srvr.createSflowIntf(val.Obj)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - CREATE_SFLOW_INTF")
		}

	case UPDATE_SFLOW_INTF:
		if val, ok := req.Data.(*UpdateSflowIntfInArgs); ok {
			srvr.updateSflowIntf(val.OldObj, val.NewObj, val.AttrSet)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - UPDATE_SFLOW_INTF")
		}

	case DELETE_SFLOW_INTF:
		if val, ok := req.Data.(*DeleteSflowIntfInArgs); ok {
			srvr.deleteSflowIntf(val.Obj)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - DELETE_SFLOW_INTF")
		}

	case GET_SFLOW_COLLECTOR_STATE:
		var retObj GetSflowCollectorStateOutArgs
		if val, ok := req.Data.(*GetSflowCollectorStateInArgs); ok {
			retObj.Obj, retObj.Err = srvr.getSflowCollectorState(val.IpAddr)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - GET_SFLOW_COLLECTOR_STATE")
		}
		srvr.ReplyChan <- interface{}(&retObj)

	case GET_BULK_SFLOW_COLLECTOR_STATE:
		var retObj GetBulkSflowCollectorStateOutArgs
		if val, ok := req.Data.(*GetBulkInArgs); ok {
			retObj.BulkObj, retObj.Err = srvr.getBulkSflowCollectorState(val.FromIdx, val.Count)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - GET_BULK_SFLOW_COLLECTOR_STATE")
		}
		srvr.ReplyChan <- interface{}(&retObj)

	case GET_SFLOW_INTF_STATE:
		var retObj GetSflowIntfStateOutArgs
		if val, ok := req.Data.(*GetSflowIntfStateInArgs); ok {
			retObj.Obj, retObj.Err = srvr.getSflowIntfState(val.IntfRef)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - GET_SFLOW_INTF_STATE")
		}
		srvr.ReplyChan <- interface{}(&retObj)

	case GET_BULK_SFLOW_INTF_STATE:
		var retObj GetBulkSflowIntfStateOutArgs
		if val, ok := req.Data.(*GetBulkInArgs); ok {
			retObj.BulkObj, retObj.Err = srvr.getBulkSflowIntfState(val.FromIdx, val.Count)
		} else {
			logger.Err("Invalid data format received by server.Request opcode - GET_BULK_SFLOW_INTF_STATE")
		}
		srvr.ReplyChan <- interface{}(&retObj)

	default:
		logger.Err("Invalid server request received. Ignoring request : ", req.Op)
	}
	logger.Debug("Server request served")
}
