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

package api

import (
	"infra/platformd/objects"
	"infra/platformd/server"
)

var svr *server.PlatformdServer

//var logger *logging.LogginIntf

//Initialize server handle
func InitApiLayer(server *server.PlatformdServer) {
	svr = server
	svr.Logger.Info("Initializing API layer")
}

func GetPlatformSystemState(objName string) (*objects.PlatformSystemState, error) {
	return &objects.PlatformSystemState{
		ObjName:   objName,
		SerialNum: "000011112222333",
	}, nil
}

func GetBulkPlatformSystemState(fromIdx, count int) (*objects.PlatformSystemStateGetInfo, error) {
	obj := objects.PlatformSystemState{
		ObjName:   "PlatformSystem",
		SerialNum: "000011112222333",
	}
	var retObj objects.PlatformSystemStateGetInfo

	retObj.EndIdx = 0
	retObj.Count = 1
	retObj.More = false
	retObj.List = append(retObj.List, &obj)
	return &retObj, nil
}
