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

package hw

import (
	"asicdServices"
	"encoding/json"
	"errors"
	"git.apache.org/thrift.git/lib/go/thrift"
	"infra/statsd/objects"
	"io/ioutil"
	"strconv"
	"time"
	"utils/ipcutils"
	"utils/logging"
)

//Define hdl so server does not directly import asicdServices
type HwHdl struct {
	*asicdServices.ASICDServicesClient
}

type ClientJson struct {
	Name string `json:Name`
	Port int    `json:Port`
}

type AsicdClient struct {
	Address            string
	Transport          thrift.TTransport
	PtrProtocolFactory *thrift.TBinaryProtocolFactory
	ClientHdl          *asicdServices.ASICDServicesClient
}

func GetHwClntHdl(paramsFile string, logger logging.LoggerIntf) *HwHdl {
	var asicdClient AsicdClient
	logger.Debug("Inside connectToServers...paramsFile", paramsFile)
	var clientsList []ClientJson

	bytes, err := ioutil.ReadFile(paramsFile)
	if err != nil {
		logger.Err("Error in reading configuration file")
		return nil
	}

	err = json.Unmarshal(bytes, &clientsList)
	if err != nil {
		logger.Err("Error in Unmarshalling Json")
		return nil
	}

	for _, client := range clientsList {
		if client.Name == "asicd" {
			logger.Debug("found asicd at port", client.Port)
			asicdClient.Address = "localhost:" + strconv.Itoa(client.Port)
			asicdClient.Transport, asicdClient.PtrProtocolFactory, err = ipcutils.CreateIPCHandles(asicdClient.Address)
			if err != nil {
				logger.Err("Failed to connect to Asicd, retrying until connection is successful")
				count := 0
				ticker := time.NewTicker(time.Duration(1000) * time.Millisecond)
				for _ = range ticker.C {
					asicdClient.Transport, asicdClient.PtrProtocolFactory, err = ipcutils.CreateIPCHandles(asicdClient.Address)
					if err == nil {
						ticker.Stop()
						break
					}
					count++
					if (count % 10) == 0 {
						logger.Err("Still can't connect to Asicd, retrying..")
					}
				}

			}
			logger.Info("Connected to Asicd")
			asicdClient.ClientHdl = asicdServices.NewASICDServicesClientFactory(asicdClient.Transport, asicdClient.PtrProtocolFactory)
			return &HwHdl{
				ASICDServicesClient: asicdClient.ClientHdl,
			}
		}
	}
	return nil
}

func (h *HwHdl) GetSflowNetdevInfo() ([]objects.SflowNetdevInfo, error) {
	var err error
	var objList []objects.SflowNetdevInfo
	info, err := h.GetBulkSFlowIntfInfo()
	if (err != nil) || (len(info.SFlowIntfInfoList) == 0) {
		err = errors.New("GetBulkSflowIntfInfo returned nil list")
	} else {
		for _, val := range info.SFlowIntfInfoList {
			objList = append(objList, objects.SflowNetdevInfo{
				IfIndex: val.PortIfIdx,
				IntfRef: val.IntfRef,
			})
		}
	}
	return objList, err
}
