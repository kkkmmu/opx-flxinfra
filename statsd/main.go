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

package main

import (
	"infra/statsd/api"
	"infra/statsd/rpc"
	"infra/statsd/server"
	"strconv"
	"strings"
	"utils/dmnBase"
)

const (
	DMN_NAME = "statsd"
)

type Daemon struct {
	*dmnBase.FSBaseDmn
	daemonServer *server.DmnServer
	rpcServer    *rpc.RPCServer
}

var dmn Daemon

func main() {
	dmn.FSBaseDmn = dmnBase.NewBaseDmn(DMN_NAME, DMN_NAME)
	ok := dmn.Init()
	if !ok {
		panic("Daemon Base initialization failed for statsd")
	}

	serverInitParams := &server.ServerInitParams{
		DmnName:   DMN_NAME,
		ParamsDir: dmn.ParamsDir,
		DbHdl:     dmn.DbHdl,
		Logger:    dmn.FSBaseDmn.Logger,
	}
	dmn.daemonServer = server.NewSTATSDServer(serverInitParams)
	go dmn.daemonServer.Serve()

	//Wait for server started msg before initializing API layer
	_ = <-dmn.daemonServer.InitCompleteCh

	//Initialize api layer
	api.InitApiLayer(dmn.daemonServer, dmn.Logger)

	var rpcServerAddr string
	for _, value := range dmn.FSBaseDmn.ClientsList {
		if value.Name == strings.ToLower(DMN_NAME) {
			rpcServerAddr = "localhost:" + strconv.Itoa(value.Port)
			break
		}
	}
	if rpcServerAddr == "" {
		panic("Daemon statsd is not part of the system profile")
	}
	dmn.rpcServer = rpc.NewRPCServer(rpcServerAddr, dmn.FSBaseDmn.Logger, dmn.FSBaseDmn.DbHdl)

	dmn.StartKeepAlive()

	//Start RPC server
	dmn.FSBaseDmn.Logger.Info("Daemon Server started for statsd")
	dmn.rpcServer.Serve()
	panic("Daemon RPC Server terminated statsd")
}
