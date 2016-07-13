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
	"infra/platformd/pluginManager"
	"infra/platformd/pluginManager/pluginCommon"
	"time"
	"utils/dbutils"
	"utils/logging"
)

type PlatformdServer struct {
	dmnName        string
	paramsDir      string
	plugin         pluginManager.PluginIntf
	dbHdl          *dbutils.DBUtil
	Logger         logging.LoggerIntf
	InitCompleteCh chan bool
}

type InitParams struct {
	DmnName     string
	ParamsDir   string
	CfgFileName string
	DbHdl       *dbutils.DBUtil
	Logger      logging.LoggerIntf
}

func NewPlatformdServer(initParams *InitParams) *PlatformdServer {
	var server PlatformdServer

	server.dmnName = initParams.DmnName
	server.paramsDir = initParams.ParamsDir
	server.dbHdl = initParams.DbHdl
	server.Logger = initParams.Logger
	server.InitCompleteCh = make(chan bool)

	CfgFileInfo, err := parseCfgFile(initParams.CfgFileName)
	if err != nil {
		server.Logger.Err("Failed to parse platformd config file, using default values for all attributes")
	}
	pluginInitParams := &pluginCommon.PluginInitParams{
		Logger: server.Logger,
	}
	server.plugin = pluginManager.NewPlugin(CfgFileInfo.PluginName, pluginInitParams)
	return &server
}

func (server *PlatformdServer) initServer() error {
	//Initialize plugin layer first
	err := server.plugin.Init()
	if err != nil {
		return err
	}

	return err
}

func (server *PlatformdServer) Serve() {
	err := server.initServer()
	if err != nil {
		panic(err)
	}
	server.InitCompleteCh <- true
	server.Logger.Debug("Server initialization complete, starting cfg/state listerner")
	for {
		time.Sleep(10)
	}
}
