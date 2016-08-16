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

package server

import (
	"utils/logging"
)

type NMGRServer struct {
	Logger      logging.LoggerIntf
	paramsDir   string
	ReqChan     chan *ServerRequest
	InitDone    chan bool
	EventEnable bool
	FaultEnable bool
	AlarmEnable bool
}

func NewNMGRServer(initParams *ServerInitParams) *NMGRServer {
	nMgrServer := &NMGRServer{}
	nMgrServer.Logger = initParams.Logger
	nMgrServer.paramsDir = initParams.ParamsDir
	nMgrServer.InitDone = make(chan bool)
	nMgrServer.ReqChan = make(chan *ServerRequest)
	return nMgrServer
}

func (server *NMGRServer) InitServer() {
	server.EventEnable = true
	server.FaultEnable = true
	server.AlarmEnable = true
}

func (server *NMGRServer) StartServer() {
	server.InitServer()
	server.InitDone <- true
	for {
		select {
		case req := <-server.ReqChan:
			server.Logger.Info("Server request received - ", *req)
			switch req.Op {
			case UPDATE_NOTIFIER_ENABLE:
				if val, ok := req.Data.(*UpdateNotifierEnableInArgs); ok {
					server.updateNotifierEnable(val.NotifierEnableOld, val.NotifierEnableNew, val.AttrSet)
				}
			default:
				server.Logger.Err("Error: Server received unrecognized request - ", req.Op)
			}
		}
	}
}
