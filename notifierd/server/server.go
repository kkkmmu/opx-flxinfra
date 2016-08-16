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
	//	"fmt"
	"time"
	"utils/logging"
)

type NMGRServer struct {
	Logger   logging.LoggerIntf
	InitDone chan bool
}

func NewNMGRServer(logger *logging.Writer) *NMGRServer {
	nMgrServer := &NMGRServer{}
	nMgrServer.Logger = logger
	nMgrServer.InitDone = make(chan bool)
	return nMgrServer
}

func (server *NMGRServer) StartServer() {
	//server.InitServer()
	server.InitDone <- true
	for {
		/*
		   select {
		   case req := <-server.ReqChan:
		           server.Logger.Info(fmt.Sprintln("Server request received - ", *req))
		           server.handleRPCRequest(req)
		   }
		*/
		time.Sleep(30 * time.Second)
	}
}
