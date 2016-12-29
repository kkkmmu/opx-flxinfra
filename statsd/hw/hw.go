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
	"sync"
	"time"
	"utils/ipcutils"
	"utils/logging"
)

//Define interface so server does not directly import asicdServices
type HwHdlIntf interface {
	GetSflowNetdevInfo() ([]objects.SflowNetdevInfo, error)
	SflowEnable(ifIndex int32) error
	SflowDisable(ifIndex int32) error
	SflowSetSamplingRate(ifIndex, rate int32) error
	SflowGetIntfCounters(intfRef string) *objects.SflowIntfCounterInfo
	SflowGetIntfCfgInfo(intfRef string) *objects.SflowIntfCfgInfo
}

type HwHdl struct {
	sync.Mutex
	logger logging.LoggerIntf
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

func GetHwClntHdl(paramsFile string, logger logging.LoggerIntf) HwHdlIntf {
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
				logger:              logger,
			}
		}
	}
	return nil
}

func (h *HwHdl) GetSflowNetdevInfo() ([]objects.SflowNetdevInfo, error) {
	var err error
	var objList []objects.SflowNetdevInfo
	h.logger.Debug("HW call : GetSflowNetdevInfo")
	h.Lock()
	info, err := h.GetBulkSFlowIntfInfo()
	if (err != nil) || (len(info.SFlowIntfInfoList) == 0) {
		err = errors.New("GetBulkSflowIntfInfo returned nil list")
	} else {
		for _, val := range info.SFlowIntfInfoList {
			objList = append(objList, objects.SflowNetdevInfo{
				IfIndex:    val.PortIfIdx,
				IntfRef:    val.IntfRef,
				NetDevName: val.NetDevName,
			})
		}
	}
	h.Unlock()
	h.logger.Debug("HW call returned : GetSflowNetdevInfo.", objList, err)
	return objList, err
}

func (h *HwHdl) SflowEnable(ifIndex int32) error {
	h.logger.Debug("HW call : SflowEnable")
	h.Lock()
	_, err := h.EnableSFlowSampling(ifIndex, true)
	h.Unlock()
	h.logger.Debug("HW call returned : SflowEnable.", err)
	return err
}

func (h *HwHdl) SflowDisable(ifIndex int32) error {
	h.logger.Debug("HW call : SflowDisable")
	h.Lock()
	_, err := h.EnableSFlowSampling(ifIndex, false)
	h.Unlock()
	h.logger.Debug("HW call returned : SflowDisable.", err)
	return err
}

func (h *HwHdl) SflowSetSamplingRate(ifIndex, rate int32) error {
	h.logger.Debug("HW call : SflowSetSamplingRate")
	h.Lock()
	_, err := h.SetSFlowSamplingRate(ifIndex, rate)
	h.Unlock()
	h.logger.Debug("HW call returned : SflowSetSamplingRate.", err)
	return err
}

func (h *HwHdl) SflowGetIntfCounters(intfRef string) *objects.SflowIntfCounterInfo {
	var convObj *objects.SflowIntfCounterInfo
	h.logger.Debug("HW call : SflowGetPortCounters - ", intfRef)
	h.Lock()
	obj, err := h.GetPortState(intfRef)
	var opState bool
	if err == nil {
		if obj.OperState == objects.ADMIN_STATE_UP {
			opState = true
		} else {
			opState = false
		}
		convObj = &objects.SflowIntfCounterInfo{
			OperState:     opState,
			IfInOctets:    uint64(obj.IfInOctets),
			IfInUcastPkts: uint64(obj.IfInUcastPkts),
			//FIXME : Need to add support for IfInMcastPkts and IfInBcastPkts
			IfInDiscards:      uint64(obj.IfInDiscards),
			IfInErrors:        uint64(obj.IfInErrors),
			IfInUnknownProtos: uint64(obj.IfInUnknownProtos),
			IfOutOctets:       uint64(obj.IfOutOctets),
			IfOutUcastPkts:    uint64(obj.IfOutUcastPkts),
			//FIXME : Need to add support for IfOutMcastPkts and IfOutBcastPkts
			IfOutDiscards: uint64(obj.IfOutDiscards),
			IfOutErrors:   uint64(obj.IfOutErrors),
		}
	} else {
		h.logger.Err("HW call : SflowGetPortCounters error when retrieving counters for interface - ", intfRef, err)
	}
	h.Unlock()
	return convObj
}

func (h *HwHdl) SflowGetIntfCfgInfo(intfRef string) *objects.SflowIntfCfgInfo {
	var convObj *objects.SflowIntfCfgInfo
	h.logger.Debug("HW call : SflowGetIntfCfgInfo - ", intfRef)
	h.Lock()
	obj, err := h.GetPort(intfRef)
	if err == nil {
		var fullDuplex bool
		//FIXME:  Need to expose define from asicd
		if obj.Duplex == "Full_Duplex" {
			fullDuplex = true
		}
		convObj = &objects.SflowIntfCfgInfo{
			Speed:      obj.Speed,
			FullDuplex: fullDuplex,
		}
	} else {
		h.logger.Err("HW call : SflowGetIntfCfgInfo error when retrieving counters for interface - ", intfRef, err)
	}
	h.Unlock()
	return convObj
}
