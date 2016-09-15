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
	"infra/platformd/pluginManager/pluginCommon"
	"models/events"
	"sync"
	"time"
	"utils/eventUtils"
	"utils/logging"
	"utils/ringBuffer"
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
	PMClassAAdminState       string
	PMClassBAdminState       string
	PMClassCAdminState       string
}

const (
	qsfpClassAInterval time.Duration = time.Duration(10) * time.Second // Polling Interval 10 sec
	qsfpClassABufSize  int           = 6 * 60 * 24                     //Storage for 24 hrs
	qsfpClassBInterval time.Duration = time.Duration(15) * time.Minute // Polling Interval 15 mins
	qsfpClassBBufSize  int           = 4 * 24                          // Storage for 24 hrs
	qsfpClassCInterval time.Duration = time.Duration(24) * time.Hour   // Polling Interval 24 Hrs
	qsfpClassCBufSize  int           = 365                             // Storage for 365 days
)

const (
	TemperatureRes uint8 = 0
	VoltageRes     uint8 = 1
	RX1PowerRes    uint8 = 2
	RX2PowerRes    uint8 = 3
	RX3PowerRes    uint8 = 4
	RX4PowerRes    uint8 = 5
	TX1PowerRes    uint8 = 6
	TX2PowerRes    uint8 = 7
	TX3PowerRes    uint8 = 8
	TX4PowerRes    uint8 = 9
	TX1BiasRes     uint8 = 10
	TX2BiasRes     uint8 = 11
	TX3BiasRes     uint8 = 12
	TX4BiasRes     uint8 = 13
	MaxNumRes      uint8 = 14 // Should be always last
)

type QsfpEventStatus [MaxNumRes]EventStatus

type QsfpEventData struct {
	Value    float64
	Resource string
}

type QsfpEvent struct {
	EventId    events.EventId
	EventData  QsfpEventData
	ChannelNum uint8
}

func getQsfpResourcId(res string) (uint8, error) {
	switch res {
	case "Temperature":
		return 0, nil
	case "Voltage":
		return 1, nil
	case "RX1Power":
		return 2, nil
	case "RX2Power":
		return 3, nil
	case "RX3Power":
		return 4, nil
	case "RX4Power":
		return 5, nil
	case "TX1Power":
		return 6, nil
	case "TX2Power":
		return 7, nil
	case "TX3Power":
		return 8, nil
	case "TX4Power":
		return 9, nil
	case "TX1Bias":
		return 10, nil
	case "TX2Bias":
		return 11, nil
	case "TX3Bias":
		return 12, nil
	case "TX4Bias":
		return 13, nil
	default:
		return 0, errors.New("Invalid Resource Name")
	}
	return 0, errors.New("Invalid Resource Name")
}

type QsfpManager struct {
	logger              logging.LoggerIntf
	plugin              PluginIntf
	configMutex         sync.RWMutex
	stateMutex          sync.RWMutex
	configDB            []QsfpConfig
	classAPMTimer       *time.Timer
	classAMutex         sync.RWMutex
	classAPM            []map[uint8]*ringBuffer.RingBuffer
	classBPMTimer       *time.Timer
	classBMutex         sync.RWMutex
	classBPM            []map[uint8]*ringBuffer.RingBuffer
	classCPMTimer       *time.Timer
	classCMutex         sync.RWMutex
	classCPM            []map[uint8]*ringBuffer.RingBuffer
	eventMsgStatusMutex sync.RWMutex
	eventMsgStatus      []QsfpEventStatus
	qsfpStatus          []bool
}

var QsfpMgr QsfpManager

func (qMgr *QsfpManager) Init(logger logging.LoggerIntf, plugin PluginIntf) {
	qMgr.logger = logger
	qMgr.plugin = plugin
	numOfQsfps := qMgr.plugin.GetMaxNumOfQsfp()

	qMgr.classAPM = make([]map[uint8]*ringBuffer.RingBuffer, numOfQsfps)
	qMgr.classBPM = make([]map[uint8]*ringBuffer.RingBuffer, numOfQsfps)
	qMgr.classCPM = make([]map[uint8]*ringBuffer.RingBuffer, numOfQsfps)
	qMgr.eventMsgStatus = make([]QsfpEventStatus, numOfQsfps)
	qMgr.qsfpStatus = make([]bool, numOfQsfps)
	for idx := 0; idx < numOfQsfps; idx++ {
		qMgr.classAPM[idx] = make(map[uint8]*ringBuffer.RingBuffer)
		qMgr.classBPM[idx] = make(map[uint8]*ringBuffer.RingBuffer)
		qMgr.classCPM[idx] = make(map[uint8]*ringBuffer.RingBuffer)
	}

	qMgr.configMutex.Lock()
	qMgr.configDB = make([]QsfpConfig, numOfQsfps)
	for id := 0; id < numOfQsfps; id++ {
		qsfpCfgEnt := qMgr.configDB[id]
		qsfpCfgEnt.AdminState = "Disable"
		qsfpCfgEnt.HigherAlarmTemperature = 100.0
		qsfpCfgEnt.HigherAlarmVoltage = 10.0
		qsfpCfgEnt.HigherAlarmRXPower = 100.0
		qsfpCfgEnt.HigherAlarmTXPower = 100.0
		qsfpCfgEnt.HigherAlarmTXBias = 100.0
		qsfpCfgEnt.HigherWarningTemperature = 100.0
		qsfpCfgEnt.HigherWarningVoltage = 10.0
		qsfpCfgEnt.HigherWarningRXPower = 100.0
		qsfpCfgEnt.HigherWarningTXPower = 100.0
		qsfpCfgEnt.HigherWarningTXBias = 100.0
		qsfpCfgEnt.LowerAlarmTemperature = -100.0
		qsfpCfgEnt.LowerAlarmVoltage = -10.0
		qsfpCfgEnt.LowerAlarmRXPower = -100.0
		qsfpCfgEnt.LowerAlarmTXPower = -100.0
		qsfpCfgEnt.LowerAlarmTXBias = -100.0
		qsfpCfgEnt.LowerWarningTemperature = -100.0
		qsfpCfgEnt.LowerWarningVoltage = -10.0
		qsfpCfgEnt.LowerWarningRXPower = -100.0
		qsfpCfgEnt.LowerWarningTXPower = -100.0
		qsfpCfgEnt.LowerWarningTXBias = -100.0
		qsfpCfgEnt.PMClassAAdminState = "Disable"
		qsfpCfgEnt.PMClassBAdminState = "Disable"
		qsfpCfgEnt.PMClassCAdminState = "Disable"
		qMgr.configDB[id] = qsfpCfgEnt
	}
	qMgr.configMutex.Unlock()
	qMgr.StartQsfpPM()
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

	qMgr.stateMutex.Lock()
	qsfpState, err := qMgr.plugin.GetQsfpState(id)
	qMgr.stateMutex.Unlock()
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

	qMgr.configMutex.RLock()
	qsfpCfgEnt := qMgr.configDB[id-1]
	obj.QsfpId = id
	obj.AdminState = qsfpCfgEnt.AdminState
	obj.HigherAlarmTemperature = qsfpCfgEnt.HigherAlarmTemperature
	obj.HigherAlarmVoltage = qsfpCfgEnt.HigherAlarmVoltage
	obj.HigherAlarmRXPower = qsfpCfgEnt.HigherAlarmRXPower
	obj.HigherAlarmTXPower = qsfpCfgEnt.HigherAlarmTXPower
	obj.HigherAlarmTXBias = qsfpCfgEnt.HigherAlarmTXBias
	obj.HigherWarningTemperature = qsfpCfgEnt.HigherWarningTemperature
	obj.HigherWarningVoltage = qsfpCfgEnt.HigherWarningVoltage
	obj.HigherWarningRXPower = qsfpCfgEnt.HigherWarningRXPower
	obj.HigherWarningTXPower = qsfpCfgEnt.HigherWarningTXPower
	obj.HigherWarningTXBias = qsfpCfgEnt.HigherWarningTXBias
	obj.LowerAlarmTemperature = qsfpCfgEnt.LowerAlarmTemperature
	obj.LowerAlarmVoltage = qsfpCfgEnt.LowerAlarmVoltage
	obj.LowerAlarmRXPower = qsfpCfgEnt.LowerAlarmRXPower
	obj.LowerAlarmTXPower = qsfpCfgEnt.LowerAlarmTXPower
	obj.LowerAlarmTXBias = qsfpCfgEnt.LowerAlarmTXBias
	obj.LowerWarningTemperature = qsfpCfgEnt.LowerWarningTemperature
	obj.LowerWarningVoltage = qsfpCfgEnt.LowerWarningVoltage
	obj.LowerWarningRXPower = qsfpCfgEnt.LowerWarningRXPower
	obj.LowerWarningTXPower = qsfpCfgEnt.LowerWarningTXPower
	obj.LowerWarningTXBias = qsfpCfgEnt.LowerWarningTXBias
	obj.PMClassAAdminState = qsfpCfgEnt.PMClassAAdminState
	obj.PMClassBAdminState = qsfpCfgEnt.PMClassBAdminState
	obj.PMClassCAdminState = qsfpCfgEnt.PMClassCAdminState
	qMgr.configMutex.RUnlock()

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

func genQsfpUpdateMask(attrset []bool) uint32 {
	var mask uint32 = 0

	if attrset == nil {
		mask = objects.QSFP_UPDATE_ADMIN_STATE |
			objects.QSFP_UPDATE_HIGHER_ALARM_TEMPERATURE |
			objects.QSFP_UPDATE_HIGHER_ALARM_VOLTAGE |
			objects.QSFP_UPDATE_HIGHER_ALARM_RX_POWER |
			objects.QSFP_UPDATE_HIGHER_ALARM_TX_POWER |
			objects.QSFP_UPDATE_HIGHER_ALARM_TX_BIAS |
			objects.QSFP_UPDATE_HIGHER_WARN_TEMPERATURE |
			objects.QSFP_UPDATE_HIGHER_WARN_VOLTAGE |
			objects.QSFP_UPDATE_HIGHER_WARN_RX_POWER |
			objects.QSFP_UPDATE_HIGHER_WARN_TX_POWER |
			objects.QSFP_UPDATE_HIGHER_WARN_TX_BIAS |
			objects.QSFP_UPDATE_LOWER_ALARM_TEMPERATURE |
			objects.QSFP_UPDATE_LOWER_ALARM_VOLTAGE |
			objects.QSFP_UPDATE_LOWER_ALARM_RX_POWER |
			objects.QSFP_UPDATE_LOWER_ALARM_TX_POWER |
			objects.QSFP_UPDATE_LOWER_ALARM_TX_BIAS |
			objects.QSFP_UPDATE_LOWER_WARN_TEMPERATURE |
			objects.QSFP_UPDATE_LOWER_WARN_VOLTAGE |
			objects.QSFP_UPDATE_LOWER_WARN_RX_POWER |
			objects.QSFP_UPDATE_LOWER_WARN_TX_POWER |
			objects.QSFP_UPDATE_LOWER_WARN_TX_BIAS |
			objects.QSFP_UPDATE_PM_CLASS_A_ADMIN_STATE |
			objects.QSFP_UPDATE_PM_CLASS_B_ADMIN_STATE |
			objects.QSFP_UPDATE_PM_CLASS_C_ADMIN_STATE
	} else {
		for idx, val := range attrset {
			if true == val {
				switch idx {
				case 0:
					//QSFP Id
				case 1:
					mask |= objects.QSFP_UPDATE_ADMIN_STATE
				case 2:
					mask |= objects.QSFP_UPDATE_HIGHER_ALARM_TEMPERATURE
				case 3:
					mask |= objects.QSFP_UPDATE_HIGHER_ALARM_VOLTAGE
				case 4:
					mask |= objects.QSFP_UPDATE_HIGHER_ALARM_RX_POWER
				case 5:
					mask |= objects.QSFP_UPDATE_HIGHER_ALARM_TX_POWER
				case 6:
					mask |= objects.QSFP_UPDATE_HIGHER_ALARM_TX_BIAS
				case 7:
					mask |= objects.QSFP_UPDATE_HIGHER_WARN_TEMPERATURE
				case 8:
					mask |= objects.QSFP_UPDATE_HIGHER_WARN_VOLTAGE
				case 9:
					mask |= objects.QSFP_UPDATE_HIGHER_WARN_RX_POWER
				case 10:
					mask |= objects.QSFP_UPDATE_HIGHER_WARN_TX_POWER
				case 11:
					mask |= objects.QSFP_UPDATE_HIGHER_WARN_TX_BIAS
				case 12:
					mask |= objects.QSFP_UPDATE_LOWER_ALARM_TEMPERATURE
				case 13:
					mask |= objects.QSFP_UPDATE_LOWER_ALARM_VOLTAGE
				case 14:
					mask |= objects.QSFP_UPDATE_LOWER_ALARM_RX_POWER
				case 15:
					mask |= objects.QSFP_UPDATE_LOWER_ALARM_TX_POWER
				case 16:
					mask |= objects.QSFP_UPDATE_LOWER_ALARM_TX_BIAS
				case 17:
					mask |= objects.QSFP_UPDATE_LOWER_WARN_TEMPERATURE
				case 18:
					mask |= objects.QSFP_UPDATE_LOWER_WARN_VOLTAGE
				case 19:
					mask |= objects.QSFP_UPDATE_LOWER_WARN_RX_POWER
				case 20:
					mask |= objects.QSFP_UPDATE_LOWER_WARN_TX_POWER
				case 21:
					mask |= objects.QSFP_UPDATE_LOWER_WARN_TX_BIAS
				case 22:
					mask |= objects.QSFP_UPDATE_PM_CLASS_A_ADMIN_STATE
				case 23:
					mask |= objects.QSFP_UPDATE_PM_CLASS_B_ADMIN_STATE
				case 24:
					mask |= objects.QSFP_UPDATE_PM_CLASS_C_ADMIN_STATE
				}
			}
		}
	}
	return mask
}

func (qMgr *QsfpManager) UpdateQsfpConfig(oldCfg *objects.QsfpConfig, newCfg *objects.QsfpConfig, attrset []bool) (bool, error) {
	if qMgr.plugin == nil {
		return false, errors.New("Invalid platform plugin")
	}
	numOfQsfp := len(qMgr.configDB)
	if newCfg.QsfpId > int32(numOfQsfp) || newCfg.QsfpId < int32(1) {
		return false, errors.New("Invalid Qsfp Id")
	}
	qMgr.configMutex.Lock()
	var cfgEnt QsfpConfig
	qsfpCfgEnt := qMgr.configDB[newCfg.QsfpId-1]
	mask := genQsfpUpdateMask(attrset)
	if mask&objects.QSFP_UPDATE_ADMIN_STATE == objects.QSFP_UPDATE_ADMIN_STATE {
		if newCfg.AdminState != "Enable" && newCfg.AdminState != "Disable" {
			qMgr.configMutex.Unlock()
			return false, errors.New("Invalid AdminState Value")
		}

		cfgEnt.AdminState = newCfg.AdminState
	} else {
		cfgEnt.AdminState = qsfpCfgEnt.AdminState
	}
	if mask&objects.QSFP_UPDATE_HIGHER_ALARM_TEMPERATURE == objects.QSFP_UPDATE_HIGHER_ALARM_TEMPERATURE {
		cfgEnt.HigherAlarmTemperature = newCfg.HigherAlarmTemperature
	} else {
		cfgEnt.HigherAlarmTemperature = qsfpCfgEnt.HigherAlarmTemperature
	}
	if mask&objects.QSFP_UPDATE_HIGHER_ALARM_VOLTAGE == objects.QSFP_UPDATE_HIGHER_ALARM_VOLTAGE {
		cfgEnt.HigherAlarmVoltage = newCfg.HigherAlarmVoltage
	} else {
		cfgEnt.HigherAlarmVoltage = qsfpCfgEnt.HigherAlarmVoltage
	}
	if mask&objects.QSFP_UPDATE_HIGHER_ALARM_RX_POWER == objects.QSFP_UPDATE_HIGHER_ALARM_RX_POWER {
		cfgEnt.HigherAlarmRXPower = newCfg.HigherAlarmRXPower
	} else {
		cfgEnt.HigherAlarmRXPower = qsfpCfgEnt.HigherAlarmRXPower
	}
	if mask&objects.QSFP_UPDATE_HIGHER_ALARM_TX_POWER == objects.QSFP_UPDATE_HIGHER_ALARM_TX_POWER {
		cfgEnt.HigherAlarmTXPower = newCfg.HigherAlarmTXPower
	} else {
		cfgEnt.HigherAlarmTXPower = qsfpCfgEnt.HigherAlarmTXPower
	}
	if mask&objects.QSFP_UPDATE_HIGHER_ALARM_TX_BIAS == objects.QSFP_UPDATE_HIGHER_ALARM_TX_BIAS {
		cfgEnt.HigherAlarmTXBias = newCfg.HigherAlarmTXBias
	} else {
		cfgEnt.HigherAlarmTXBias = qsfpCfgEnt.HigherAlarmTXBias
	}
	if mask&objects.QSFP_UPDATE_HIGHER_WARN_TEMPERATURE == objects.QSFP_UPDATE_HIGHER_WARN_TEMPERATURE {
		cfgEnt.HigherWarningTemperature = newCfg.HigherWarningTemperature
	} else {
		cfgEnt.HigherWarningTemperature = qsfpCfgEnt.HigherWarningTemperature
	}
	if mask&objects.QSFP_UPDATE_HIGHER_WARN_VOLTAGE == objects.QSFP_UPDATE_HIGHER_WARN_VOLTAGE {
		cfgEnt.HigherWarningVoltage = newCfg.HigherWarningVoltage
	} else {
		cfgEnt.HigherWarningVoltage = qsfpCfgEnt.HigherWarningVoltage
	}
	if mask&objects.QSFP_UPDATE_HIGHER_WARN_RX_POWER == objects.QSFP_UPDATE_HIGHER_WARN_RX_POWER {
		cfgEnt.HigherWarningRXPower = newCfg.HigherWarningRXPower
	} else {
		cfgEnt.HigherWarningRXPower = qsfpCfgEnt.HigherWarningRXPower
	}
	if mask&objects.QSFP_UPDATE_HIGHER_WARN_TX_POWER == objects.QSFP_UPDATE_HIGHER_WARN_TX_POWER {
		cfgEnt.HigherWarningTXPower = newCfg.HigherWarningTXPower
	} else {
		cfgEnt.HigherWarningTXPower = qsfpCfgEnt.HigherWarningTXPower
	}
	if mask&objects.QSFP_UPDATE_HIGHER_WARN_TX_BIAS == objects.QSFP_UPDATE_HIGHER_WARN_TX_BIAS {
		cfgEnt.HigherWarningTXBias = newCfg.HigherWarningTXBias
	} else {
		cfgEnt.HigherWarningTXBias = qsfpCfgEnt.HigherWarningTXBias
	}
	if mask&objects.QSFP_UPDATE_LOWER_ALARM_TEMPERATURE == objects.QSFP_UPDATE_LOWER_ALARM_TEMPERATURE {
		cfgEnt.LowerAlarmTemperature = newCfg.LowerAlarmTemperature
	} else {
		cfgEnt.LowerAlarmTemperature = qsfpCfgEnt.LowerAlarmTemperature
	}
	if mask&objects.QSFP_UPDATE_LOWER_ALARM_VOLTAGE == objects.QSFP_UPDATE_LOWER_ALARM_VOLTAGE {
		cfgEnt.LowerAlarmVoltage = newCfg.LowerAlarmVoltage
	} else {
		cfgEnt.LowerAlarmVoltage = qsfpCfgEnt.LowerAlarmVoltage
	}
	if mask&objects.QSFP_UPDATE_LOWER_ALARM_RX_POWER == objects.QSFP_UPDATE_LOWER_ALARM_RX_POWER {
		cfgEnt.LowerAlarmRXPower = newCfg.LowerAlarmRXPower
	} else {
		cfgEnt.LowerAlarmRXPower = qsfpCfgEnt.LowerAlarmRXPower
	}
	if mask&objects.QSFP_UPDATE_LOWER_ALARM_TX_POWER == objects.QSFP_UPDATE_LOWER_ALARM_TX_POWER {
		cfgEnt.LowerAlarmTXPower = newCfg.LowerAlarmTXPower
	} else {
		cfgEnt.LowerAlarmTXPower = qsfpCfgEnt.LowerAlarmTXPower
	}
	if mask&objects.QSFP_UPDATE_LOWER_ALARM_TX_BIAS == objects.QSFP_UPDATE_LOWER_ALARM_TX_BIAS {
		cfgEnt.LowerAlarmTXBias = newCfg.LowerAlarmTXBias
	} else {
		cfgEnt.LowerAlarmTXBias = qsfpCfgEnt.LowerAlarmTXBias
	}
	if mask&objects.QSFP_UPDATE_LOWER_WARN_TEMPERATURE == objects.QSFP_UPDATE_LOWER_WARN_TEMPERATURE {
		cfgEnt.LowerWarningTemperature = newCfg.LowerWarningTemperature
	} else {
		cfgEnt.LowerWarningTemperature = qsfpCfgEnt.LowerWarningTemperature
	}
	if mask&objects.QSFP_UPDATE_LOWER_WARN_VOLTAGE == objects.QSFP_UPDATE_LOWER_WARN_VOLTAGE {
		cfgEnt.LowerWarningVoltage = newCfg.LowerWarningVoltage
	} else {
		cfgEnt.LowerWarningVoltage = qsfpCfgEnt.LowerWarningVoltage
	}
	if mask&objects.QSFP_UPDATE_LOWER_WARN_RX_POWER == objects.QSFP_UPDATE_LOWER_WARN_RX_POWER {
		cfgEnt.LowerWarningRXPower = newCfg.LowerWarningRXPower
	} else {
		cfgEnt.LowerWarningRXPower = qsfpCfgEnt.LowerWarningRXPower
	}
	if mask&objects.QSFP_UPDATE_LOWER_WARN_TX_POWER == objects.QSFP_UPDATE_LOWER_WARN_TX_POWER {
		cfgEnt.LowerWarningTXPower = newCfg.LowerWarningTXPower
	} else {
		cfgEnt.LowerWarningTXPower = qsfpCfgEnt.LowerWarningTXPower
	}
	if mask&objects.QSFP_UPDATE_LOWER_WARN_TX_BIAS == objects.QSFP_UPDATE_LOWER_WARN_TX_BIAS {
		cfgEnt.LowerWarningTXBias = newCfg.LowerWarningTXBias
	} else {
		cfgEnt.LowerWarningTXBias = qsfpCfgEnt.LowerWarningTXBias
	}
	if mask&objects.QSFP_UPDATE_PM_CLASS_A_ADMIN_STATE == objects.QSFP_UPDATE_PM_CLASS_A_ADMIN_STATE {
		if newCfg.PMClassAAdminState != "Enable" && newCfg.PMClassAAdminState != "Disable" {
			qMgr.configMutex.Unlock()
			return false, errors.New("Invalid PMClassAAdminState Value")
		}
		cfgEnt.PMClassAAdminState = newCfg.PMClassAAdminState
	} else {
		cfgEnt.PMClassAAdminState = qsfpCfgEnt.PMClassAAdminState
	}
	if mask&objects.QSFP_UPDATE_PM_CLASS_B_ADMIN_STATE == objects.QSFP_UPDATE_PM_CLASS_B_ADMIN_STATE {
		if newCfg.PMClassBAdminState != "Enable" && newCfg.PMClassBAdminState != "Disable" {
			qMgr.configMutex.Unlock()
			return false, errors.New("Invalid PMClassBAdminState Value")
		}
		cfgEnt.PMClassBAdminState = newCfg.PMClassBAdminState
	} else {
		cfgEnt.PMClassBAdminState = qsfpCfgEnt.PMClassBAdminState
	}
	if mask&objects.QSFP_UPDATE_PM_CLASS_C_ADMIN_STATE == objects.QSFP_UPDATE_PM_CLASS_C_ADMIN_STATE {
		if newCfg.PMClassCAdminState != "Enable" && newCfg.PMClassCAdminState != "Disable" {
			qMgr.configMutex.Unlock()
			return false, errors.New("Invalid PMClassCAdminState Value")
		}
		cfgEnt.PMClassCAdminState = newCfg.PMClassCAdminState
	} else {
		cfgEnt.PMClassCAdminState = qsfpCfgEnt.PMClassCAdminState
	}

	if !(cfgEnt.HigherAlarmTemperature >= cfgEnt.HigherWarningTemperature &&
		cfgEnt.HigherWarningTemperature > cfgEnt.LowerWarningTemperature &&
		cfgEnt.LowerWarningTemperature >= cfgEnt.LowerAlarmTemperature) {
		qMgr.configMutex.Unlock()
		return false, errors.New("Invalid configuration, please verify the thresholds")
	}

	if !(cfgEnt.HigherAlarmVoltage >= cfgEnt.HigherWarningVoltage &&
		cfgEnt.HigherWarningVoltage > cfgEnt.LowerWarningVoltage &&
		cfgEnt.LowerWarningVoltage >= cfgEnt.LowerAlarmVoltage) {
		qMgr.configMutex.Unlock()
		return false, errors.New("Invalid configuration, please verify the thresholds")
	}

	if !(cfgEnt.HigherAlarmRXPower >= cfgEnt.HigherWarningRXPower &&
		cfgEnt.HigherWarningRXPower > cfgEnt.LowerWarningRXPower &&
		cfgEnt.LowerWarningRXPower >= cfgEnt.LowerAlarmRXPower) {
		qMgr.configMutex.Unlock()
		return false, errors.New("Invalid configuration, please verify the thresholds")
	}

	if !(cfgEnt.HigherAlarmTXPower >= cfgEnt.HigherWarningTXPower &&
		cfgEnt.HigherWarningTXPower > cfgEnt.LowerWarningTXPower &&
		cfgEnt.LowerWarningTXPower >= cfgEnt.LowerAlarmTXPower) {
		qMgr.configMutex.Unlock()
		return false, errors.New("Invalid configuration, please verify the thresholds")
	}

	if !(cfgEnt.HigherAlarmTXBias >= cfgEnt.HigherWarningTXBias &&
		cfgEnt.HigherWarningTXBias > cfgEnt.LowerWarningTXBias &&
		cfgEnt.LowerWarningTXBias >= cfgEnt.LowerAlarmTXBias) {
		qMgr.configMutex.Unlock()
		return false, errors.New("Invalid configuration, please verify the thresholds")
	}

	if qsfpCfgEnt.AdminState != cfgEnt.AdminState {
		qMgr.stateMutex.Lock()
		switch cfgEnt.AdminState {
		case "Enable":
			qMgr.qsfpStatus[newCfg.QsfpId-1] = true
		case "Disable":
			//Clear all the existing Faults
			qMgr.logger.Info("Clear all the existing Faults")
		}
		qMgr.stateMutex.Unlock()
	}
	if qMgr.configDB[newCfg.QsfpId-1].PMClassAAdminState != cfgEnt.PMClassAAdminState {
		if cfgEnt.PMClassAAdminState == "Disable" {
			// Flush PM RingBuffer
			qMgr.flushHistoricPM(newCfg.QsfpId, "Class-A")
			qMgr.logger.Info("Flush Class A PM Ring buffer")
		}
	}
	if qMgr.configDB[newCfg.QsfpId-1].PMClassBAdminState != cfgEnt.PMClassBAdminState {
		if cfgEnt.PMClassBAdminState == "Disable" {
			// Flush PM RingBuffer
			qMgr.flushHistoricPM(newCfg.QsfpId, "Class-B")
			qMgr.logger.Info("Flush Class B PM Ring buffer")
		}
	}
	if qMgr.configDB[newCfg.QsfpId-1].PMClassCAdminState != cfgEnt.PMClassCAdminState {
		if cfgEnt.PMClassCAdminState == "Disable" {
			// Flush PM RingBuffer
			qMgr.flushHistoricPM(newCfg.QsfpId, "Class-C")
			qMgr.logger.Info("Flush Class C PM Ring buffer")
		}
	}
	qMgr.configDB[newCfg.QsfpId-1] = cfgEnt

	qMgr.configMutex.Unlock()

	return true, nil
}

func (qMgr *QsfpManager) flushHistoricPM(QsfpId int32, Class string) {
	switch Class {
	case "Class-A":
		qMgr.classAMutex.Lock()
		for idx := 0; idx < int(MaxNumRes); idx++ {
			resId := uint8(idx)
			qMgr.classAPM[QsfpId-1][resId].FlushRingBuffer()
		}
		qMgr.classBMutex.Lock()
	case "Class-B":
		qMgr.classBMutex.Lock()
		for idx := 0; idx < int(MaxNumRes); idx++ {
			resId := uint8(idx)
			qMgr.classBPM[QsfpId-1][resId].FlushRingBuffer()
		}
		qMgr.classBMutex.Lock()
	case "Class-C":
		qMgr.classCMutex.Lock()
		for idx := 0; idx < int(MaxNumRes); idx++ {
			resId := uint8(idx)
			qMgr.classCPM[QsfpId-1][resId].FlushRingBuffer()
		}
		qMgr.classCMutex.Lock()
	}
}

func (qMgr *QsfpManager) GetQsfpPMState(QsfpId int32, Resource string, Class string) (*objects.QsfpPMState, error) {
	var qsfpPMObj objects.QsfpPMState
	if qMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	numOfQsfp := len(qMgr.configDB)
	if QsfpId > int32(numOfQsfp) || QsfpId < int32(1) {
		return nil, errors.New("Invalid Qsfp Id")
	}
	qMgr.configMutex.RLock()
	qsfpCfgEnt := qMgr.configDB[QsfpId-1]
	/*
		if qsfpCfgEnt.AdminState == "Disable" {
			qMgr.configMutex.RUnlock()
			return nil, errors.New("Qsfp is admin disabled")
		}
		qMgr.configMutex.RUnlock()
	*/
	resId, err := getQsfpResourcId(Resource)
	if err != nil {
		qMgr.configMutex.RUnlock()
		return nil, errors.New("Invalid Resource Name")
	}
	switch Class {
	case "Class-A":
		if qsfpCfgEnt.PMClassAAdminState == "Enable" {
			qMgr.classAMutex.RLock()
			qsfpPMObj.Data = qMgr.classAPM[QsfpId-1][resId].GetListOfEntriesFromRingBuffer()
			qMgr.classAMutex.RUnlock()
		}
	case "Class-B":
		if qsfpCfgEnt.PMClassBAdminState == "Enable" {
			qMgr.classBMutex.RLock()
			qsfpPMObj.Data = qMgr.classBPM[QsfpId-1][resId].GetListOfEntriesFromRingBuffer()
			qMgr.classBMutex.RUnlock()
		}
	case "Class-C":
		if qsfpCfgEnt.PMClassCAdminState == "Enable" {
			qMgr.classCMutex.RLock()
			qsfpPMObj.Data = qMgr.classCPM[QsfpId-1][resId].GetListOfEntriesFromRingBuffer()
			qMgr.classCMutex.RUnlock()
		}
	default:
		qMgr.configMutex.RUnlock()
		return nil, errors.New("Invalid Class")
	}
	qMgr.configMutex.RUnlock()
	qsfpPMObj.QsfpId = QsfpId
	qsfpPMObj.Resource = Resource
	qsfpPMObj.Class = Class
	return &qsfpPMObj, nil
}

func getPMData(data *pluginCommon.QsfpPMData, resId uint8) (objects.QsfpPMData, error) {
	qsfpPMData := objects.QsfpPMData{
		TimeStamp: time.Now().String(),
	}
	switch resId {
	case TemperatureRes:
		qsfpPMData.Value = data.Temperature
	case VoltageRes:
		qsfpPMData.Value = data.Voltage
	case RX1PowerRes:
		qsfpPMData.Value = data.RX1Power
	case RX2PowerRes:
		qsfpPMData.Value = data.RX2Power
	case RX3PowerRes:
		qsfpPMData.Value = data.RX3Power
	case RX4PowerRes:
		qsfpPMData.Value = data.RX4Power
	case TX1PowerRes:
		qsfpPMData.Value = data.TX1Power
	case TX2PowerRes:
		qsfpPMData.Value = data.TX2Power
	case TX3PowerRes:
		qsfpPMData.Value = data.TX3Power
	case TX4PowerRes:
		qsfpPMData.Value = data.TX4Power
	case TX1BiasRes:
		qsfpPMData.Value = data.TX1Bias
	case TX2BiasRes:
		qsfpPMData.Value = data.TX2Bias
	case TX3BiasRes:
		qsfpPMData.Value = data.TX3Bias
	case TX4BiasRes:
		qsfpPMData.Value = data.TX4Bias
	default:
		return qsfpPMData, errors.New("Invalid resource Id")
	}
	return qsfpPMData, nil
}

func getCurrentEventStatus(pmData *pluginCommon.QsfpPMData, qsfpCfgEnt QsfpConfig) QsfpEventStatus {
	var curEventStatus QsfpEventStatus
	var data float64
	var highAlarmData float64
	var highWarnData float64
	var lowWarnData float64
	var lowAlarmData float64
	for idx := 0; idx < int(MaxNumRes); idx++ {
		resId := uint8(idx)
		switch resId {
		case TemperatureRes:
			data = pmData.Temperature
			highAlarmData = qsfpCfgEnt.HigherAlarmTemperature
			highWarnData = qsfpCfgEnt.HigherWarningTemperature
			lowWarnData = qsfpCfgEnt.LowerWarningTemperature
			lowAlarmData = qsfpCfgEnt.LowerAlarmTemperature
		case VoltageRes:
			data = pmData.Voltage
			highAlarmData = qsfpCfgEnt.HigherAlarmVoltage
			highWarnData = qsfpCfgEnt.HigherWarningVoltage
			lowWarnData = qsfpCfgEnt.LowerWarningVoltage
			lowAlarmData = qsfpCfgEnt.LowerAlarmVoltage
		case RX1PowerRes, RX2PowerRes, RX3PowerRes, RX4PowerRes:
			switch resId {
			case RX1PowerRes:
				data = pmData.RX1Power
			case RX2PowerRes:
				data = pmData.RX2Power
			case RX3PowerRes:
				data = pmData.RX3Power
			case RX4PowerRes:
				data = pmData.RX4Power
			}
			highAlarmData = qsfpCfgEnt.HigherAlarmRXPower
			highWarnData = qsfpCfgEnt.HigherWarningRXPower
			lowWarnData = qsfpCfgEnt.LowerWarningRXPower
			lowAlarmData = qsfpCfgEnt.LowerAlarmRXPower
		case TX1PowerRes, TX2PowerRes, TX3PowerRes, TX4PowerRes:
			switch resId {
			case TX1PowerRes:
				data = pmData.TX1Power
			case TX2PowerRes:
				data = pmData.TX2Power
			case TX3PowerRes:
				data = pmData.TX3Power
			case TX4PowerRes:
				data = pmData.TX4Power
			}
			highAlarmData = qsfpCfgEnt.HigherAlarmTXPower
			highWarnData = qsfpCfgEnt.HigherWarningTXPower
			lowWarnData = qsfpCfgEnt.LowerWarningTXPower
			lowAlarmData = qsfpCfgEnt.LowerAlarmTXPower
		case TX1BiasRes, TX2BiasRes, TX3BiasRes, TX4BiasRes:
			switch resId {
			case TX1BiasRes:
				data = pmData.TX1Bias
			case TX2BiasRes:
				data = pmData.TX2Bias
			case TX3BiasRes:
				data = pmData.TX3Bias
			case TX4BiasRes:
				data = pmData.TX4Bias
			}
			highAlarmData = qsfpCfgEnt.HigherAlarmTXBias
			highWarnData = qsfpCfgEnt.HigherWarningTXBias
			lowWarnData = qsfpCfgEnt.LowerWarningTXBias
			lowAlarmData = qsfpCfgEnt.LowerAlarmTXBias
		}
		if data >= highAlarmData {
			curEventStatus[idx].SentHigherAlarm = true
		}
		if data >= highWarnData {
			curEventStatus[idx].SentHigherWarn = true
		}
		if data <= lowAlarmData {
			curEventStatus[idx].SentLowerAlarm = true
		}
		if data <= lowWarnData {
			curEventStatus[idx].SentLowerWarn = true
		}
	}
	return curEventStatus
}

func getQsfpHigherAlarmEvent(idx int, prevEventStatus QsfpEventStatus, curEventStatus QsfpEventStatus) (QsfpEvent, error) {
	var evt QsfpEvent
	resId := uint8(idx)
	if prevEventStatus[idx].SentHigherAlarm != curEventStatus[idx].SentHigherAlarm {
		if curEventStatus[idx].SentHigherAlarm == true {
			switch resId {
			case TemperatureRes:
				evt.EventId = events.QsfpTemperatureHigherTCAAlarm
			case VoltageRes:
				evt.EventId = events.QsfpVoltageHigherTCAAlarm
			case RX1PowerRes, RX2PowerRes, RX3PowerRes, RX4PowerRes:
				evt.EventId = events.QsfpRXPowerHigherTCAAlarm
			case TX1PowerRes, TX2PowerRes, TX3PowerRes, TX4PowerRes:
				evt.EventId = events.QsfpTXPowerHigherTCAAlarm
			case TX1BiasRes, TX2BiasRes, TX3BiasRes, TX4BiasRes:
				evt.EventId = events.QsfpTXBiasHigherTCAAlarm
			}
		} else {
			switch resId {
			case TemperatureRes:
				evt.EventId = events.QsfpTemperatureHigherTCAAlarmClear
			case VoltageRes:
				evt.EventId = events.QsfpVoltageHigherTCAAlarmClear
			case RX1PowerRes, RX2PowerRes, RX3PowerRes, RX4PowerRes:
				evt.EventId = events.QsfpRXPowerHigherTCAAlarmClear
			case TX1PowerRes, TX2PowerRes, TX3PowerRes, TX4PowerRes:
				evt.EventId = events.QsfpTXPowerHigherTCAAlarmClear
			case TX1BiasRes, TX2BiasRes, TX3BiasRes, TX4BiasRes:
				evt.EventId = events.QsfpTXBiasHigherTCAAlarmClear
			}
		}
		switch resId {
		case RX1PowerRes, TX1PowerRes, TX1BiasRes:
			evt.ChannelNum = 1
		case RX2PowerRes, TX2PowerRes, TX2BiasRes:
			evt.ChannelNum = 2
		case RX3PowerRes, TX3PowerRes, TX3BiasRes:
			evt.ChannelNum = 3
		case RX4PowerRes, TX4PowerRes, TX4BiasRes:
			evt.ChannelNum = 4
		}
		return evt, nil
	}
	return evt, errors.New("No Event")
}

func getQsfpHigherWarnEvent(idx int, prevEventStatus QsfpEventStatus, curEventStatus QsfpEventStatus) (QsfpEvent, error) {
	var evt QsfpEvent
	resId := uint8(idx)
	if prevEventStatus[idx].SentHigherWarn != curEventStatus[idx].SentHigherWarn {
		if curEventStatus[idx].SentHigherWarn == true {
			switch resId {
			case TemperatureRes:
				evt.EventId = events.QsfpTemperatureHigherTCAWarn
			case VoltageRes:
				evt.EventId = events.QsfpVoltageHigherTCAWarn
			case RX1PowerRes, RX2PowerRes, RX3PowerRes, RX4PowerRes:
				evt.EventId = events.QsfpRXPowerHigherTCAWarn
			case TX1PowerRes, TX2PowerRes, TX3PowerRes, TX4PowerRes:
				evt.EventId = events.QsfpTXPowerHigherTCAWarn
			case TX1BiasRes, TX2BiasRes, TX3BiasRes, TX4BiasRes:
				evt.EventId = events.QsfpTXBiasHigherTCAWarn
			}
		} else {
			switch resId {
			case TemperatureRes:
				evt.EventId = events.QsfpTemperatureHigherTCAWarnClear
			case VoltageRes:
				evt.EventId = events.QsfpVoltageHigherTCAWarnClear
			case RX1PowerRes, RX2PowerRes, RX3PowerRes, RX4PowerRes:
				evt.EventId = events.QsfpRXPowerHigherTCAWarnClear
			case TX1PowerRes, TX2PowerRes, TX3PowerRes, TX4PowerRes:
				evt.EventId = events.QsfpTXPowerHigherTCAWarnClear
			case TX1BiasRes, TX2BiasRes, TX3BiasRes, TX4BiasRes:
				evt.EventId = events.QsfpTXBiasHigherTCAWarnClear
			}
		}
		switch resId {
		case RX1PowerRes, TX1PowerRes, TX1BiasRes:
			evt.ChannelNum = 1
		case RX2PowerRes, TX2PowerRes, TX2BiasRes:
			evt.ChannelNum = 2
		case RX3PowerRes, TX3PowerRes, TX3BiasRes:
			evt.ChannelNum = 3
		case RX4PowerRes, TX4PowerRes, TX4BiasRes:
			evt.ChannelNum = 4
		}
		return evt, nil
	}
	return evt, errors.New("No Event")
}

func getQsfpLowerAlarmEvent(idx int, prevEventStatus QsfpEventStatus, curEventStatus QsfpEventStatus) (QsfpEvent, error) {
	var evt QsfpEvent
	resId := uint8(idx)
	if prevEventStatus[idx].SentLowerAlarm != curEventStatus[idx].SentLowerAlarm {
		if curEventStatus[idx].SentLowerAlarm == true {
			switch resId {
			case TemperatureRes:
				evt.EventId = events.QsfpTemperatureLowerTCAAlarm
			case VoltageRes:
				evt.EventId = events.QsfpVoltageLowerTCAAlarm
			case RX1PowerRes, RX2PowerRes, RX3PowerRes, RX4PowerRes:
				evt.EventId = events.QsfpRXPowerLowerTCAAlarm
			case TX1PowerRes, TX2PowerRes, TX3PowerRes, TX4PowerRes:
				evt.EventId = events.QsfpTXPowerLowerTCAAlarm
			case TX1BiasRes, TX2BiasRes, TX3BiasRes, TX4BiasRes:
				evt.EventId = events.QsfpTXBiasLowerTCAAlarm
			}
		} else {
			switch resId {
			case TemperatureRes:
				evt.EventId = events.QsfpTemperatureLowerTCAAlarmClear
			case VoltageRes:
				evt.EventId = events.QsfpVoltageLowerTCAAlarmClear
			case RX1PowerRes, RX2PowerRes, RX3PowerRes, RX4PowerRes:
				evt.EventId = events.QsfpRXPowerLowerTCAAlarmClear
			case TX1PowerRes, TX2PowerRes, TX3PowerRes, TX4PowerRes:
				evt.EventId = events.QsfpTXPowerLowerTCAAlarmClear
			case TX1BiasRes, TX2BiasRes, TX3BiasRes, TX4BiasRes:
				evt.EventId = events.QsfpTXBiasLowerTCAAlarmClear
			}
		}
		switch resId {
		case RX1PowerRes, TX1PowerRes, TX1BiasRes:
			evt.ChannelNum = 1
		case RX2PowerRes, TX2PowerRes, TX2BiasRes:
			evt.ChannelNum = 2
		case RX3PowerRes, TX3PowerRes, TX3BiasRes:
			evt.ChannelNum = 3
		case RX4PowerRes, TX4PowerRes, TX4BiasRes:
			evt.ChannelNum = 4
		}
		return evt, nil
	}
	return evt, errors.New("No Event")
}

func getQsfpLowerWarnEvent(idx int, prevEventStatus QsfpEventStatus, curEventStatus QsfpEventStatus) (QsfpEvent, error) {
	var evt QsfpEvent
	resId := uint8(idx)
	if prevEventStatus[idx].SentLowerWarn != curEventStatus[idx].SentLowerWarn {
		if curEventStatus[idx].SentLowerWarn == true {
			switch resId {
			case TemperatureRes:
				evt.EventId = events.QsfpTemperatureLowerTCAWarn
			case VoltageRes:
				evt.EventId = events.QsfpVoltageLowerTCAWarn
			case RX1PowerRes, RX2PowerRes, RX3PowerRes, RX4PowerRes:
				evt.EventId = events.QsfpRXPowerLowerTCAWarn
			case TX1PowerRes, TX2PowerRes, TX3PowerRes, TX4PowerRes:
				evt.EventId = events.QsfpTXPowerLowerTCAWarn
			case TX1BiasRes, TX2BiasRes, TX3BiasRes, TX4BiasRes:
				evt.EventId = events.QsfpTXBiasLowerTCAWarn
			}
		} else {
			switch resId {
			case TemperatureRes:
				evt.EventId = events.QsfpTemperatureLowerTCAWarnClear
			case VoltageRes:
				evt.EventId = events.QsfpVoltageLowerTCAWarnClear
			case RX1PowerRes, RX2PowerRes, RX3PowerRes, RX4PowerRes:
				evt.EventId = events.QsfpRXPowerLowerTCAWarnClear
			case TX1PowerRes, TX2PowerRes, TX3PowerRes, TX4PowerRes:
				evt.EventId = events.QsfpTXPowerLowerTCAWarnClear
			case TX1BiasRes, TX2BiasRes, TX3BiasRes, TX4BiasRes:
				evt.EventId = events.QsfpTXBiasLowerTCAWarnClear
			}
		}
		switch resId {
		case RX1PowerRes, TX1PowerRes, TX1BiasRes:
			evt.ChannelNum = 1
		case RX2PowerRes, TX2PowerRes, TX2BiasRes:
			evt.ChannelNum = 2
		case RX3PowerRes, TX3PowerRes, TX3BiasRes:
			evt.ChannelNum = 3
		case RX4PowerRes, TX4PowerRes, TX4BiasRes:
			evt.ChannelNum = 4
		}
		return evt, nil
	}
	return evt, errors.New("No Event")
}

func getListofQsfpEvent(prevEventStatus QsfpEventStatus, curEventStatus QsfpEventStatus) []QsfpEvent {
	var evts []QsfpEvent
	for idx := 0; idx < int(MaxNumRes); idx++ {
		evt, err := getQsfpHigherAlarmEvent(idx, prevEventStatus, curEventStatus)
		if err == nil {
			evts = append(evts, evt)
		}
		evt, err = getQsfpHigherWarnEvent(idx, prevEventStatus, curEventStatus)
		if err == nil {
			evts = append(evts, evt)
		}
		evt, err = getQsfpLowerAlarmEvent(idx, prevEventStatus, curEventStatus)
		if err == nil {
			evts = append(evts, evt)
		}
		evt, err = getQsfpLowerWarnEvent(idx, prevEventStatus, curEventStatus)
		if err == nil {
			evts = append(evts, evt)
		}
	}
	return evts
}

func (qMgr *QsfpManager) publishQsfpEvents(qsfpId int32, data *pluginCommon.QsfpPMData, evts []QsfpEvent) {
	eventKey := events.QsfpKey{
		QsfpId: qsfpId,
	}
	txEvent := eventUtils.TxEvent{
		Key:            eventKey,
		AdditionalInfo: "",
	}

	for _, evt := range evts {
		txEvent.EventId = evt.EventId
		switch evt.EventId {
		case events.QsfpTemperatureHigherTCAAlarm,
			events.QsfpTemperatureHigherTCAWarn,
			events.QsfpTemperatureLowerTCAAlarm,
			events.QsfpTemperatureLowerTCAWarn:
			evt.EventData.Resource = "Temperature"
			evt.EventData.Value = data.Temperature
		case events.QsfpVoltageHigherTCAAlarm,
			events.QsfpVoltageHigherTCAWarn,
			events.QsfpVoltageLowerTCAAlarm,
			events.QsfpVoltageLowerTCAWarn:
			evt.EventData.Resource = "Voltage"
			evt.EventData.Value = data.Voltage
		case events.QsfpRXPowerHigherTCAAlarm,
			events.QsfpRXPowerHigherTCAWarn,
			events.QsfpRXPowerLowerTCAAlarm,
			events.QsfpRXPowerLowerTCAWarn:
			switch evt.ChannelNum {
			case 1:
				evt.EventData.Resource = "RX1Power"
				evt.EventData.Value = data.RX1Power
			case 2:
				evt.EventData.Resource = "RX2Power"
				evt.EventData.Value = data.RX2Power
			case 3:
				evt.EventData.Resource = "RX3Power"
				evt.EventData.Value = data.RX3Power
			case 4:
				evt.EventData.Resource = "RX4Power"
				evt.EventData.Value = data.RX4Power
			}
		case events.QsfpTXPowerHigherTCAAlarm,
			events.QsfpTXPowerHigherTCAWarn,
			events.QsfpTXPowerLowerTCAAlarm,
			events.QsfpTXPowerLowerTCAWarn:
			switch evt.ChannelNum {
			case 1:
				evt.EventData.Resource = "TX1Power"
				evt.EventData.Value = data.TX1Power
			case 2:
				evt.EventData.Resource = "TX2Power"
				evt.EventData.Value = data.TX2Power
			case 3:
				evt.EventData.Resource = "TX3Power"
				evt.EventData.Value = data.TX3Power
			case 4:
				evt.EventData.Resource = "TX4Power"
				evt.EventData.Value = data.TX4Power
			}
		case events.QsfpTXBiasHigherTCAAlarm,
			events.QsfpTXBiasHigherTCAWarn,
			events.QsfpTXBiasLowerTCAAlarm,
			events.QsfpTXBiasLowerTCAWarn:
			switch evt.ChannelNum {
			case 1:
				evt.EventData.Resource = "TX1Bias"
				evt.EventData.Value = data.TX1Bias
			case 2:
				evt.EventData.Resource = "TX2Bias"
				evt.EventData.Value = data.TX2Bias
			case 3:
				evt.EventData.Resource = "TX3Bias"
				evt.EventData.Value = data.TX3Bias
			case 4:
				evt.EventData.Resource = "TX4Bias"
				evt.EventData.Value = data.TX4Bias
			}
		}
		txEvent.AdditionalData = evt.EventData
		txEvt := txEvent
		err := eventUtils.PublishEvents(&txEvt)
		if err != nil {
			qMgr.logger.Err("Error publishing event:", txEvt)
		}
	}
}

func (qMgr *QsfpManager) processQsfpEvents(data *pluginCommon.QsfpPMData, qsfpId int32) {
	qsfpCfgEnt := qMgr.configDB[qsfpId-1]

	var evts []QsfpEvent
	qMgr.eventMsgStatusMutex.RLock()
	prevEventStatus := qMgr.eventMsgStatus[qsfpId-1]
	qMgr.eventMsgStatusMutex.RUnlock()
	curEventStatus := getCurrentEventStatus(data, qsfpCfgEnt)
	evts = append(evts, getListofQsfpEvent(prevEventStatus, curEventStatus)...)
	if prevEventStatus != curEventStatus {
		qMgr.eventMsgStatusMutex.Lock()
		qMgr.eventMsgStatus[qsfpId-1] = curEventStatus
		qMgr.eventMsgStatusMutex.Unlock()
	}
	qMgr.publishQsfpEvents(qsfpId, data, evts)
}

func (qMgr *QsfpManager) ProcessQsfpPMData(data *pluginCommon.QsfpPMData, qsfpId int32, class string) {
	qsfpCfgEnt := qMgr.configDB[qsfpId-1]
	if qsfpCfgEnt.AdminState == "Enable" {
		qMgr.processQsfpEvents(data, qsfpId)
	}
	switch class {
	case "Class-A":
		if qsfpCfgEnt.PMClassAAdminState == "Enable" {
			qMgr.classAMutex.Lock()
			for idx := 0; idx < int(MaxNumRes); idx++ {
				resId := uint8(idx)
				pmData, err := getPMData(data, resId)
				if err != nil {
					qMgr.logger.Err("Wrong resource Id:", resId)
					continue
				}
				qMgr.classAPM[qsfpId-1][resId].InsertIntoRingBuffer(pmData)
			}
			qMgr.classAMutex.Unlock()
		}
	case "Class-B":
		if qsfpCfgEnt.PMClassBAdminState == "Enable" {
			qMgr.classBMutex.Lock()
			for idx := 0; idx < int(MaxNumRes); idx++ {
				resId := uint8(idx)
				pmData, err := getPMData(data, resId)
				if err != nil {
					qMgr.logger.Err("Wrong resource Id:", resId)
					continue
				}
				qMgr.classBPM[qsfpId-1][resId].InsertIntoRingBuffer(pmData)
			}
			qMgr.classBMutex.Unlock()
		}
	case "Class-C":
		qMgr.classCMutex.Lock()
		for idx := 0; idx < int(MaxNumRes); idx++ {
			resId := uint8(idx)
			pmData, err := getPMData(data, resId)
			if err != nil {
				qMgr.logger.Err("Wrong resource Id:", resId)
				continue
			}
			qMgr.classCPM[qsfpId-1][resId].InsertIntoRingBuffer(pmData)
		}
		qMgr.classCMutex.Unlock()
	}
}

func (qMgr *QsfpManager) StartQsfpPMClass(class string) {
	qMgr.stateMutex.Lock()
	qMgr.configMutex.RLock()
	for idx := 0; idx < len(qMgr.configDB); idx++ {
		qsfpId := int32(idx + 1)
		qsfpCfgEnt := qMgr.configDB[idx]
		if qsfpCfgEnt.AdminState == "Enable" {
			qsfpPMData, err := qMgr.plugin.GetQsfpPMData(qsfpId)
			if err == nil {
				if qMgr.qsfpStatus[idx] == false {
					qMgr.qsfpStatus[idx] = true
					// Raise Qsfp Present Event
					qMgr.logger.Info("Raise Qsfp Present Event")
				}
				qMgr.ProcessQsfpPMData(&qsfpPMData, qsfpId, class)
			} else {
				if qMgr.qsfpStatus[idx] == true {
					qMgr.qsfpStatus[idx] = false
					// Raise Qsfp Missing Event
					qMgr.logger.Info("Raise Qsfp Missing Event")
				}
			}
		}
	}
	qMgr.configMutex.RUnlock()
	qMgr.stateMutex.Unlock()
	switch class {
	case "Class-A":
		classAPMFunc := func() {
			qMgr.stateMutex.Lock()
			qMgr.configMutex.RLock()
			for idx := 0; idx < len(qMgr.configDB); idx++ {
				qsfpId := int32(idx + 1)
				qsfpCfgEnt := qMgr.configDB[idx]
				if qsfpCfgEnt.AdminState == "Enable" {
					qsfpPMData, err := qMgr.plugin.GetQsfpPMData(qsfpId)
					if err == nil {
						if qMgr.qsfpStatus[idx] == false {
							qMgr.qsfpStatus[idx] = true
							// Raise Qsfp Present Event
							qMgr.logger.Info("Raise Qsfp Present Event")
						}
						qMgr.ProcessQsfpPMData(&qsfpPMData, qsfpId, class)
					} else {
						if qMgr.qsfpStatus[idx] == true {
							qMgr.qsfpStatus[idx] = false
							// Raise Qsfp Missing Event
							qMgr.logger.Info("Raise Qsfp Missing Event")
						}
					}
				}
			}
			qMgr.configMutex.RUnlock()
			qMgr.stateMutex.Unlock()
			qMgr.classAPMTimer.Reset(qsfpClassAInterval)
		}
		qMgr.classAPMTimer = time.AfterFunc(qsfpClassAInterval, classAPMFunc)
	case "Class-B":
		classBPMFunc := func() {
			qMgr.stateMutex.Lock()
			qMgr.configMutex.RLock()
			for idx := 0; idx < len(qMgr.configDB); idx++ {
				qsfpId := int32(idx + 1)
				qsfpCfgEnt := qMgr.configDB[idx]
				if qsfpCfgEnt.AdminState == "Enable" {
					qsfpPMData, err := qMgr.plugin.GetQsfpPMData(qsfpId)
					if err == nil {
						if qMgr.qsfpStatus[idx] == false {
							qMgr.qsfpStatus[idx] = true
							// Raise Qsfp Present Event
							qMgr.logger.Info("Raise Qsfp Present Event")
						}
						qMgr.ProcessQsfpPMData(&qsfpPMData, qsfpId, class)
					} else {
						if qMgr.qsfpStatus[idx] == true {
							qMgr.qsfpStatus[idx] = false
							// Raise Qsfp Missing Event
							qMgr.logger.Info("Raise Qsfp Missing Event")
						}
					}
				}
			}
			qMgr.configMutex.RUnlock()
			qMgr.stateMutex.Unlock()
			qMgr.classBPMTimer.Reset(qsfpClassBInterval)
		}
		qMgr.classBPMTimer = time.AfterFunc(qsfpClassBInterval, classBPMFunc)
	case "Class-C":
		classCPMFunc := func() {
			qMgr.stateMutex.Lock()
			qMgr.configMutex.RLock()
			for idx := 0; idx < len(qMgr.configDB); idx++ {
				qsfpId := int32(idx + 1)
				qsfpCfgEnt := qMgr.configDB[idx]
				if qsfpCfgEnt.AdminState == "Enable" {
					qsfpPMData, err := qMgr.plugin.GetQsfpPMData(qsfpId)
					if err == nil {
						if qMgr.qsfpStatus[idx] == false {
							qMgr.qsfpStatus[idx] = true
							// Raise Qsfp Present Event
							qMgr.logger.Info("Raise Qsfp Present Event")
						}
						qMgr.ProcessQsfpPMData(&qsfpPMData, qsfpId, class)
					} else {
						if qMgr.qsfpStatus[idx] == true {
							qMgr.qsfpStatus[idx] = false
							// Raise Qsfp Missing Event
							qMgr.logger.Info("Raise Qsfp Missing Event")
						}
					}
				}
			}
			qMgr.configMutex.RUnlock()
			qMgr.stateMutex.Unlock()
			qMgr.classCPMTimer.Reset(qsfpClassCInterval)
		}
		qMgr.classCPMTimer = time.AfterFunc(qsfpClassCInterval, classCPMFunc)
	}
}

func (qMgr *QsfpManager) InitQsfpPM() {
	for idx := 0; idx < len(qMgr.configDB); idx++ {
		for res := 0; res < int(MaxNumRes); res++ {
			qMgr.classAPM[idx][uint8(res)] = new(ringBuffer.RingBuffer)
			qMgr.classAPM[idx][uint8(res)].SetRingBufferCapacity(qsfpClassABufSize)
			qMgr.classBPM[idx][uint8(res)] = new(ringBuffer.RingBuffer)
			qMgr.classBPM[idx][uint8(res)].SetRingBufferCapacity(qsfpClassBBufSize)
			qMgr.classCPM[idx][uint8(res)] = new(ringBuffer.RingBuffer)
			qMgr.classCPM[idx][uint8(res)].SetRingBufferCapacity(qsfpClassCBufSize)
		}
	}
}

func (qMgr *QsfpManager) StartQsfpPM() {
	qMgr.InitQsfpPM()
	qMgr.StartQsfpPMClass("Class-A")
	qMgr.StartQsfpPMClass("Class-B")
	qMgr.StartQsfpPMClass("Class-C")
}
