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

package pluginManager

import (
	"errors"
	"fmt"
	"infra/platformd/objects"
	//"infra/platformd/pluginManager/pluginCommon"
	"utils/logging"
)

type QsfpState struct {
	Present            bool
	VendorName         string
	VendorOUI          string
	VendorPartNumber   string
	VendorRevision     string
	VendorSerialNumber string
	DataCode           string
	Temperature        float64
	Voltage            float64
	RX1Power           float64
	RX2Power           float64
	RX3Power           float64
	RX4Power           float64
	TX1Power           float64
	TX2Power           float64
	TX3Power           float64
	TX4Power           float64
	TX1Bias            float64
	TX2Bias            float64
	TX3Bias            float64
	TX4Bias            float64
}

type QsfpConfig struct {
	AdminState               string
	HigherAlarmTemperature   float64
	HigherAlarmVoltage       float64
	HigherAlarmRXPower       float64
	HigherAlarmTXPower       float64
	HigherAlarmTXBias        float64
	HigherWarningTemperature float64
	HigherWarningVoltage     float64
	HigherWarningRXPower     float64
	HigherWarningTXPower     float64
	HigherWarningTXBias      float64
	LowerAlarmTemperature    float64
	LowerAlarmVoltage        float64
	LowerAlarmRXPower        float64
	LowerAlarmTXPower        float64
	LowerAlarmTXBias         float64
	LowerWarningTemperature  float64
	LowerWarningVoltage      float64
	LowerWarningRXPower      float64
	LowerWarningTXPower      float64
	LowerWarningTXBias       float64
}

type QsfpManager struct {
	logger   logging.LoggerIntf
	plugin   PluginIntf
	configDB []QsfpConfig
}

var QsfpMgr QsfpManager

func (qMgr *QsfpManager) Init(logger logging.LoggerIntf, plugin PluginIntf) {
	qMgr.logger = logger
	qMgr.plugin = plugin
	numOfQsfps := qMgr.plugin.GetMaxNumOfQsfp()
	qMgr.configDB = make([]QsfpConfig, numOfQsfps)
	for id := 0; id < numOfQsfps; id++ {
		qsfpCfgEnt := qMgr.configDB[id]
		//TODO
		qMgr.configDB[id] = qsfpCfgEnt
	}
	qMgr.logger.Info("Qsfp Manager Init()")
}

func (qMgr *QsfpManager) Deinit() {
	qMgr.logger.Info("Fan Manager Deinit()")
}

func (qMgr *QsfpManager) GetQsfpState(id int32) (*objects.QsfpState, error) {
	var qsfpObj objects.QsfpState
	if qMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	numOfQsfp := int32(len(qMgr.configDB))
	if (id > numOfQsfp) && (id < 1) {
		return nil, errors.New("Invalid QsfpId")
	}

	qsfpState, err := qMgr.plugin.GetQsfpState(id)
	if err != nil {
		qsfpObj.QsfpId = id
		qsfpObj.Present = false
	} else {
		qsfpObj.QsfpId = id
		qsfpObj.Present = true
		qsfpObj.VendorName = qsfpState.VendorName
		qsfpObj.VendorOUI = qsfpState.VendorOUI
		qsfpObj.VendorPartNumber = qsfpState.VendorPartNumber
		qsfpObj.VendorRevision = qsfpState.VendorRevision
		qsfpObj.VendorSerialNumber = qsfpState.VendorSerialNumber
		qsfpObj.DataCode = qsfpState.DataCode
		qsfpObj.Temperature = qsfpState.Temperature
		qsfpObj.Voltage = qsfpState.Voltage
		qsfpObj.RX1Power = qsfpState.RX1Power
		qsfpObj.RX2Power = qsfpState.RX2Power
		qsfpObj.RX3Power = qsfpState.RX3Power
		qsfpObj.RX4Power = qsfpState.RX4Power
		qsfpObj.TX1Power = qsfpState.TX1Power
		qsfpObj.TX2Power = qsfpState.TX2Power
		qsfpObj.TX3Power = qsfpState.TX3Power
		qsfpObj.TX4Power = qsfpState.TX4Power
		qsfpObj.TX1Bias = qsfpState.TX1Bias
		qsfpObj.TX2Bias = qsfpState.TX2Bias
		qsfpObj.TX3Bias = qsfpState.TX3Bias
		qsfpObj.TX4Bias = qsfpState.TX4Bias
	}
	return &qsfpObj, err
}

func (qMgr *QsfpManager) GetBulkQsfpState(fromIdx int, cnt int) (*objects.QsfpStateGetInfo, error) {
	var retObj objects.QsfpStateGetInfo
	if qMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	if fromIdx >= len(qMgr.configDB) {
		return nil, errors.New("Invalid range")
	}
	if fromIdx+cnt > len(qMgr.configDB) {
		retObj.EndIdx = len(qMgr.configDB)
		retObj.More = false
		retObj.Count = 0
	} else {
		retObj.EndIdx = fromIdx + cnt
		retObj.More = true
		retObj.Count = len(qMgr.configDB) - retObj.EndIdx + 1
	}
	for idx := fromIdx; idx < retObj.EndIdx; idx++ {
		obj, err := qMgr.GetQsfpState(int32(idx + 1))
		if err != nil {
			qMgr.logger.Err(fmt.Sprintln("Error getting the qsfp state for QsfpId:", idx+1))
		}
		retObj.List = append(retObj.List, obj)
	}
	return &retObj, nil
}

func (qMgr *QsfpManager) GetQsfpConfig(id int32) (*objects.QsfpConfig, error) {
	var obj objects.QsfpConfig
	if qMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	numOfQsfp := int32(len(qMgr.configDB))
	if (id > numOfQsfp) && (id < int32(1)) {
		return nil, errors.New("Invalid QsfpId")
	}

	qsfpCfgEnt := qMgr.configDB[id-1]
	obj.QsfpId = id
	obj.AdminState = qsfpCfgEnt.AdminState

	// Need to be added
	return &obj, nil
}

func (qMgr *QsfpManager) GetBulkQsfpConfig(fromIdx int, cnt int) (*objects.QsfpConfigGetInfo, error) {
	var retObj objects.QsfpConfigGetInfo
	if qMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	numOfQsfp := len(qMgr.configDB)
	if fromIdx >= numOfQsfp {
		return nil, errors.New("Invalid range")
	}
	if fromIdx+cnt > numOfQsfp {
		retObj.EndIdx = numOfQsfp
		retObj.More = false
		retObj.Count = 0
	} else {
		retObj.EndIdx = fromIdx + cnt
		retObj.More = true
		retObj.Count = numOfQsfp - retObj.EndIdx + 1
	}
	for idx := fromIdx; idx < retObj.EndIdx; idx++ {
		obj, err := qMgr.GetQsfpConfig(int32(idx + 1))
		if err != nil {
			qMgr.logger.Err(fmt.Sprintln("Error getting the Qsfp state for QsfpId:", idx+1))
		}
		retObj.List = append(retObj.List, obj)
	}
	return &retObj, nil
}

func (qMgr *QsfpManager) UpdateQsfpConfig(oldCfg *objects.QsfpConfig, newCfg *objects.QsfpConfig, attrset []bool) (bool, error) {
	if qMgr.plugin == nil {
		return false, errors.New("Invalid platform plugin")
	}
	//TODO
	//ret, err := fMgr.plugin.UpdateFanConfig(newCfg)
	return true, nil
}
