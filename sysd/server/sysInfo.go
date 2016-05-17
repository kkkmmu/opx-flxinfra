//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//	 Unless required by applicable law or agreed to in writing, software
//	 distributed under the License is distributed on an "AS IS" BASIS,
//	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	 See the License for the specific language governing permissions and
//	 limitations under the License.
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
	"encoding/json"
	"fmt"
	"infra/sysd/sysdCommonDefs"
	"io/ioutil"
	"models"
	_ "utils/logging"
)

/*
type SystemInfo struct {
	Description string
	Hostname    string
	RouterId    string
	Version     string
	SwitchMac   string
	MgmtIp      string
}
*/

type SwitchCfgJson struct {
	SwitchMac   string `json:"SwitchMac"`
	RouterId    string `json:"RouterId"`
	Hostname    string `json:"HostName"`
	Version     string `json:"Version"`
	MgmtIp      string `json:"MgmtIp"`
	Description string `json:"Description"`
	//PluginList       []string          `json:"PluginList"`
	//IfNameMap        map[string]string `json:"IfNameMap"`
	//IfNamePrefix     map[string]string `json:"IfNamePrefix"`
	//SysRsvdVlanRange string            `json:"SysRsvdVlanRange"`
}

// Func to insert the entry in database
//func (sys *SystemInfo) Insert() {
func (svr *SYSDServer) InsertSystemInfoInDB() error {
	return svr.dbHdl.StoreObjectInDb(svr.SysInfo)
}

func (svr *SYSDServer) ReadSystemInfoFromDB() error {
	svr.logger.Info("Reading System Information From Db")
	dbHdl := svr.dbHdl
	if dbHdl != nil {
		var dbObj models.SystemParams
		objList, err := dbHdl.GetAllObjFromDb(dbObj)
		if err != nil {
			svr.logger.Err("DB query failed for System Info")
			return err
		}
		svr.logger.Info(fmt.Sprintln("Total System Entries are", len(objList)))
		for idx := 0; idx < len(objList); idx++ {
			dbObject := objList[idx].(models.SystemParams)
			svr.SysInfo.SwitchMac = dbObject.SwitchMac
			svr.SysInfo.RouterId = dbObject.RouterId
			svr.SysInfo.MgmtIp = dbObject.MgmtIp
			svr.SysInfo.Version = dbObject.Version
			svr.SysInfo.Description = dbObject.Description
			svr.SysInfo.Hostname = dbObject.Hostname
			break
		}
	}
	svr.logger.Info("reading system info from db done")
	return nil
}

// Func to send update nanomsg update notification to all the dameons on the system
func (svr *SYSDServer) SendSystemUpdate() ([]byte, error) {
	msgBuf, err := json.Marshal(svr.SysInfo)
	if err != nil {
		return nil, err
	}
	notification := sysdCommonDefs.Notification{
		Type:    uint8(sysdCommonDefs.SYSTEM_Info),
		Payload: msgBuf,
	}
	notificationBuf, err := json.Marshal(notification)
	if err != nil {
		return nil, err
	}
	return notificationBuf, nil
}

// Initialize system information using json file...or whatever other means are
func (svr *SYSDServer) InitSystemInfo() { //paramsDir string, logger *logging.Writer) *SystemInfo {
	//sysInfo := &SystemInfo{}
	svr.SysInfo = &models.SystemParams{}
	err := svr.ReadSystemInfoFromDB()
	if err != nil { // Reading from DB is not sucess, in that case lets parse json file and get the system info
		sysInfo := svr.SysInfo
		cfgFileData, err := ioutil.ReadFile(svr.paramsDir + "../sysprofile/systemProfile.json")
		if err != nil {
			//logger.Err(fmt.Sprintln("Error reading file, err:", err))
			svr.logger.Err(fmt.Sprintln("Error reading file, err:", err))
			//return sysInfo
		}
		// Get this info from systemProfile
		var cfg SwitchCfgJson
		err = json.Unmarshal(cfgFileData, &cfg)
		if err != nil {
			//logger.Err(fmt.Sprintln("Error Unmarshalling cfg json data, err:", err))
			svr.logger.Err(fmt.Sprintln("Error Unmarshalling cfg json data, err:", err))
			//return sysInfo
		}
		// @TODO: make this readable from system profile
		sysInfo.SwitchMac = cfg.SwitchMac
		sysInfo.RouterId = cfg.RouterId
		sysInfo.MgmtIp = cfg.MgmtIp
		sysInfo.Version = cfg.Version
		sysInfo.Description = cfg.Description
		sysInfo.Hostname = cfg.Hostname
		sysInfo.Vrf = "default"
		/*
			sysInfo.RouterId = "10.1.1.254"
			sysInfo.MgmtIp = "10.1.10.244"
			sysInfo.Version = "v1.0.0"
			sysInfo.Description = "FlexSwitch System"
			sysInfo.Hostname = "dummy HostName"
			sysInfo.Vrf = "default"
		*/
		// If the entry doesn't exist into DB then it means that we just read the entry from system profile..
		// Update the information into the DB
		err = svr.InsertSystemInfoInDB()
		if err != nil {
			svr.logger.Info(fmt.Sprintln("Failed to insert system params into db", err))
		}
	}
	svr.SendSystemUpdate()
	//return sysInfo
}
