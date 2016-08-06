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

package main

import (
	"infra/fMgrd/api"
	"infra/fMgrd/rpc"
	"infra/fMgrd/server"
	"strconv"
	"utils/dmnBase"
)

const (
	DMN_NAME = "FMGRD"
)

type fMgrDaemon struct {
	*dmnBase.FSBaseDmn
	server    *server.FMGRServer
	rpcServer *rpc.RPCServer
}

var dmn fMgrDaemon

func main() {
	/*
		fmt.Println("Starting fMgr daemon")
		paramsDir := flag.String("params", "./params", "Params directory")
		flag.Parse()
		fileName := *paramsDir
		if fileName[len(fileName)-1] != '/' {
			fileName = fileName + "/"
		}

		fmt.Println("Start logger")
		logger, err := logging.NewLogger("fMgrd", "FAULTMGR", true)
		if err != nil {
			fmt.Println("Failed to start the logger. Nothing will be logged...")
		}
		logger.Info("Started the logger successfully.")

		// Start keepalive routine
		go keepalive.InitKeepAlive("fMgrd", fileName)

		logger.Info("Starting Fault Manager server...")
		fMgrServer := server.NewFMGRServer(logger)
		go fMgrServer.StartServer(*paramsDir)

		<-fMgrServer.InitDone

		logger.Info(fmt.Sprintln("Starting Config listener..."))
		confIface := rpc.NewFMGRHandler(fMgrServer, logger)
		rpc.StartServer(logger, confIface, *paramsDir)
	*/
	// Get base daemon handle and initialize
	dmn.FSBaseDmn = dmnBase.NewBaseDmn(DMN_NAME, DMN_NAME)
	ok := dmn.Init()
	if ok == false {
		panic("Fault Manager Daemon: Base Daemon Initialization failed")
	}

	// Get server handle and start server
	dmn.server = server.NewFMGRServer(dmn.FSBaseDmn.Logger)
	go dmn.server.StartServer()

	//Initialize API layer
	api.InitApiLayer(dmn.server)

	//Get RPC server handle
	var rpcServerAddr string
	for _, value := range dmn.FSBaseDmn.ClientsList {
		if value.Name == "fMgrd" {
			rpcServerAddr = "localhost:" + strconv.Itoa(value.Port)
			break
		}
	}

	if rpcServerAddr == "" {
		panic("Fault Manager Daemon is not part of system profile")
	}

	dmn.rpcServer = rpc.NewRPCServer(rpcServerAddr, dmn.FSBaseDmn.Logger)

	// Start Keep Alive for watchdog
	dmn.StartKeepAlive()

	_ = <-dmn.server.InitDone

	//Start RPC server
	dmn.FSBaseDmn.Logger.Info("Fault Manager Daemon server started")
	dmn.rpcServer.Serve()
	panic("Fault Manager Daemon RPC server terminated")
}
