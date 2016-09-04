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
	"utils/logging"
)

type FanSensorConfig struct {
	AdminState             string
	HigherAlarmThreshold   int32
	HigherWarningThreshold int32
	LowerWarningThreshold  int32
	LowerAlarmThreshold    int32
}

type TemperatureSensorConfig struct {
	AdminState             string
	HigherAlarmThreshold   float64
	HigherWarningThreshold float64
	LowerWarningThreshold  float64
	LowerAlarmThreshold    float64
}

type VoltageSensorConfig struct {
	AdminState             string
	HigherAlarmThreshold   float64
	HigherWarningThreshold float64
	LowerWarningThreshold  float64
	LowerAlarmThreshold    float64
}

type PowerConverterSensorConfig struct {
	AdminState             string
	HigherAlarmThreshold   float64
	HigherWarningThreshold float64
	LowerWarningThreshold  float64
	LowerAlarmThreshold    float64
}

type SensorManager struct {
	logger                   logging.LoggerIntf
	plugin                   PluginIntf
	fanSensorList            []string
	fanConfigDB              map[string]FanSensorConfig
	tempSensorList           []string
	tempConfigDB             map[string]TemperatureSensorConfig
	voltageSensorList        []string
	voltageConfigDB          map[string]VoltageSensorConfig
	powerConverterSensorList []string
	powerConverterConfigDB   map[string]PowerConverterSensorConfig
	SensorState              *pluginCommon.SensorState
}

var SensorMgr SensorManager

func (sMgr *SensorManager) Init(logger logging.LoggerIntf, plugin PluginIntf) {
	sMgr.logger = logger
	sMgr.logger.Info("sensor Manager Init( Start)")
	sMgr.plugin = plugin
	sMgr.SensorState = new(pluginCommon.SensorState)
	sMgr.SensorState.FanSensor = make(map[string]pluginCommon.FanSensorData)
	sMgr.SensorState.TemperatureSensor = make(map[string]pluginCommon.TemperatureSensorData)
	sMgr.SensorState.VoltageSensor = make(map[string]pluginCommon.VoltageSensorData)
	sMgr.SensorState.PowerConverterSensor = make(map[string]pluginCommon.PowerConverterSensorData)

	err := sMgr.plugin.GetAllSensorState(sMgr.SensorState)
	if err != nil {
		sMgr.logger.Info("Sensor Manager Init() Failed")
		return
	}
	sMgr.logger.Info("Sensor State:", sMgr.SensorState)
	sMgr.fanConfigDB = make(map[string]FanSensorConfig)
	for name, _ := range sMgr.SensorState.FanSensor {
		sMgr.logger.Info("Fan Sensor:", name)
		fanCfgEnt, _ := sMgr.fanConfigDB[name]
		// TODO: Read Json
		fanCfgEnt.AdminState = "Enabled"
		fanCfgEnt.HigherAlarmThreshold = 11000
		fanCfgEnt.HigherWarningThreshold = 11000
		fanCfgEnt.LowerAlarmThreshold = 1000
		fanCfgEnt.LowerWarningThreshold = 1000
		sMgr.fanConfigDB[name] = fanCfgEnt
		sMgr.fanSensorList = append(sMgr.fanSensorList, name)
	}
	sMgr.tempConfigDB = make(map[string]TemperatureSensorConfig)
	for name, _ := range sMgr.SensorState.TemperatureSensor {
		sMgr.logger.Info("Temperature Sensor:", name)
		tempCfgEnt, _ := sMgr.tempConfigDB[name]
		// TODO: Read Json
		tempCfgEnt.AdminState = "Enabled"
		tempCfgEnt.HigherAlarmThreshold = 11000.0
		tempCfgEnt.HigherWarningThreshold = 11000.0
		tempCfgEnt.LowerAlarmThreshold = 1000.0
		tempCfgEnt.LowerWarningThreshold = 1000.0
		sMgr.tempConfigDB[name] = tempCfgEnt
		sMgr.tempSensorList = append(sMgr.tempSensorList, name)
	}
	sMgr.voltageConfigDB = make(map[string]VoltageSensorConfig)
	for name, _ := range sMgr.SensorState.VoltageSensor {
		sMgr.logger.Info("Voltage Sensor:", name)
		voltageCfgEnt, _ := sMgr.voltageConfigDB[name]
		// TODO: Read Json
		voltageCfgEnt.AdminState = "Enabled"
		voltageCfgEnt.HigherAlarmThreshold = 11000
		voltageCfgEnt.HigherWarningThreshold = 11000
		voltageCfgEnt.LowerAlarmThreshold = 1000
		voltageCfgEnt.LowerWarningThreshold = 1000
		sMgr.voltageConfigDB[name] = voltageCfgEnt
		sMgr.voltageSensorList = append(sMgr.voltageSensorList, name)
	}
	sMgr.powerConverterConfigDB = make(map[string]PowerConverterSensorConfig)
	for name, _ := range sMgr.SensorState.PowerConverterSensor {
		sMgr.logger.Info("Power Sensor:", name)
		powerConverterCfgEnt, _ := sMgr.powerConverterConfigDB[name]
		// TODO: Read Json
		powerConverterCfgEnt.AdminState = "Enabled"
		powerConverterCfgEnt.HigherAlarmThreshold = 11000
		powerConverterCfgEnt.HigherWarningThreshold = 11000
		powerConverterCfgEnt.LowerAlarmThreshold = 1000
		powerConverterCfgEnt.LowerWarningThreshold = 1000
		sMgr.powerConverterConfigDB[name] = powerConverterCfgEnt
		sMgr.powerConverterSensorList = append(sMgr.powerConverterSensorList, name)
	}
	sMgr.logger.Info("sensor Manager Init( Done)")
}

func (sMgr *SensorManager) Deinit() {
	sMgr.logger.Info("Sensor Manager Deinit()")
}

func (sMgr *SensorManager) GetFanSensorState(Name string) (*objects.FanSensorState, error) {
	var fanSensorObj objects.FanSensorState
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	_, exist := sMgr.fanConfigDB[Name]
	if !exist {
		return nil, errors.New("Invalid Fan Sensor Name")
	}

	err := sMgr.plugin.GetAllSensorState(sMgr.SensorState)
	if err != nil {
		return nil, err
	}
	fanSensorState, _ := sMgr.SensorState.FanSensor[Name]
	fanSensorObj.Name = Name
	fanSensorObj.CurrentSpeed = fanSensorState.Value
	return &fanSensorObj, err
}

func (sMgr *SensorManager) GetBulkFanSensorState(fromIdx int, cnt int) (*objects.FanSensorStateGetInfo, error) {
	var retObj objects.FanSensorStateGetInfo
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	if fromIdx >= len(sMgr.fanSensorList) {
		return nil, errors.New("Invalid range")
	}
	if fromIdx+cnt > len(sMgr.fanSensorList) {
		retObj.EndIdx = len(sMgr.fanSensorList)
		retObj.More = false
		retObj.Count = 0
	} else {
		retObj.EndIdx = fromIdx + cnt
		retObj.More = true
		retObj.Count = len(sMgr.fanSensorList) - retObj.EndIdx + 1
	}
	for idx := fromIdx; idx < retObj.EndIdx; idx++ {
		fanSensorName := sMgr.fanSensorList[idx]
		obj, err := sMgr.GetFanSensorState(fanSensorName)
		if err != nil {
			sMgr.logger.Err(fmt.Sprintln("Error getting the fan state for fan Sensor:", fanSensorName))
		}
		retObj.List = append(retObj.List, obj)
	}
	return &retObj, nil
}

func (sMgr *SensorManager) GetFanSensorConfig(Name string) (*objects.FanSensorConfig, error) {
	var fanSensorObj objects.FanSensorConfig
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	fanSensorCfgEnt, exist := sMgr.fanConfigDB[Name]
	if !exist {
		return nil, errors.New("Invalid Fan Sensor Name")
	}
	fanSensorObj.Name = Name
	fanSensorObj.AdminState = fanSensorCfgEnt.AdminState
	fanSensorObj.HigherAlarmThreshold = fanSensorCfgEnt.HigherAlarmThreshold
	fanSensorObj.HigherWarningThreshold = fanSensorCfgEnt.HigherWarningThreshold
	fanSensorObj.LowerAlarmThreshold = fanSensorCfgEnt.LowerAlarmThreshold
	fanSensorObj.LowerWarningThreshold = fanSensorCfgEnt.LowerWarningThreshold
	return &fanSensorObj, nil
}

func (sMgr *SensorManager) GetBulkFanSensorConfig(fromIdx int, cnt int) (*objects.FanSensorConfigGetInfo, error) {
	var retObj objects.FanSensorConfigGetInfo
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	if fromIdx >= len(sMgr.fanSensorList) {
		return nil, errors.New("Invalid range")
	}
	if fromIdx+cnt > len(sMgr.fanSensorList) {
		retObj.EndIdx = len(sMgr.fanSensorList)
		retObj.More = false
		retObj.Count = 0
	} else {
		retObj.EndIdx = fromIdx + cnt
		retObj.More = true
		retObj.Count = len(sMgr.fanSensorList) - retObj.EndIdx + 1
	}
	for idx := fromIdx; idx < retObj.EndIdx; idx++ {
		fanSensorName := sMgr.fanSensorList[idx]
		obj, err := sMgr.GetFanSensorConfig(fanSensorName)
		if err != nil {
			sMgr.logger.Err(fmt.Sprintln("Error getting the fan state for fan sensor:", fanSensorName))
		}
		retObj.List = append(retObj.List, obj)
	}
	return &retObj, nil
}

func genFanSensorUpdateMask(attrset []bool) uint32 {
	var mask uint32 = 0
	for idx, val := range attrset {
		if true == val {
			switch idx {
			case 0:
				//ObjKey Fan Name
			case 1:
				mask |= objects.FAN_SENSOR_UPDATE_ADMIN_STATE
			case 2:
				mask |= objects.FAN_SENSOR_UPDATE_HIGHER_ALARM_THRESHOLD
			case 3:
				mask |= objects.FAN_SENSOR_UPDATE_HIGHER_WARN_THRESHOLD
			case 4:
				mask |= objects.FAN_SENSOR_UPDATE_LOWER_WARN_THRESHOLD
			case 5:
				mask |= objects.FAN_SENSOR_UPDATE_LOWER_ALARM_THRESHOLD
			}
		}
	}
	return mask
}

func (sMgr *SensorManager) UpdateFanSensorConfig(oldCfg *objects.FanSensorConfig, newCfg *objects.FanSensorConfig, attrset []bool) (bool, error) {
	if sMgr.plugin == nil {
		return false, errors.New("Invalid platform plugin")
	}
	fanSensorCfgEnt, exist := sMgr.fanConfigDB[newCfg.Name]
	if !exist {
		return false, errors.New("Invalid FanSensor Name")
	}
	mask := genFanSensorUpdateMask(attrset)
	if mask&objects.FAN_SENSOR_UPDATE_ADMIN_STATE == objects.FAN_SENSOR_UPDATE_ADMIN_STATE {
		fanSensorCfgEnt.AdminState = newCfg.AdminState
	}
	if mask&objects.FAN_SENSOR_UPDATE_HIGHER_ALARM_THRESHOLD == objects.FAN_SENSOR_UPDATE_HIGHER_ALARM_THRESHOLD {
		fanSensorCfgEnt.HigherAlarmThreshold = newCfg.HigherAlarmThreshold
	}
	if mask&objects.FAN_SENSOR_UPDATE_HIGHER_WARN_THRESHOLD == objects.FAN_SENSOR_UPDATE_HIGHER_WARN_THRESHOLD {
		fanSensorCfgEnt.HigherWarningThreshold = newCfg.HigherWarningThreshold
	}
	if mask&objects.FAN_SENSOR_UPDATE_LOWER_ALARM_THRESHOLD == objects.FAN_SENSOR_UPDATE_LOWER_ALARM_THRESHOLD {
		fanSensorCfgEnt.LowerAlarmThreshold = newCfg.LowerAlarmThreshold
	}
	if mask&objects.FAN_SENSOR_UPDATE_LOWER_WARN_THRESHOLD == objects.FAN_SENSOR_UPDATE_LOWER_WARN_THRESHOLD {
		fanSensorCfgEnt.LowerWarningThreshold = newCfg.LowerWarningThreshold
	}
	sMgr.fanConfigDB[newCfg.Name] = fanSensorCfgEnt
	return true, nil
}

func (sMgr *SensorManager) GetTemperatureSensorState(Name string) (*objects.TemperatureSensorState, error) {
	var tempSensorObj objects.TemperatureSensorState
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	_, exist := sMgr.tempConfigDB[Name]
	if !exist {
		return nil, errors.New("Invalid Temperature Sensor Name")
	}

	err := sMgr.plugin.GetAllSensorState(sMgr.SensorState)
	if err != nil {
		return nil, err
	}
	tempSensorState, _ := sMgr.SensorState.TemperatureSensor[Name]
	tempSensorObj.Name = Name
	tempSensorObj.CurrentTemperature = tempSensorState.Value
	return &tempSensorObj, err
}

func (sMgr *SensorManager) GetBulkTemperatureSensorState(fromIdx int, cnt int) (*objects.TemperatureSensorStateGetInfo, error) {
	var retObj objects.TemperatureSensorStateGetInfo
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	if fromIdx >= len(sMgr.tempSensorList) {
		return nil, errors.New("Invalid range")
	}
	if fromIdx+cnt > len(sMgr.tempSensorList) {
		retObj.EndIdx = len(sMgr.tempSensorList)
		retObj.More = false
		retObj.Count = 0
	} else {
		retObj.EndIdx = fromIdx + cnt
		retObj.More = true
		retObj.Count = len(sMgr.tempSensorList) - retObj.EndIdx + 1
	}
	for idx := fromIdx; idx < retObj.EndIdx; idx++ {
		tempSensorName := sMgr.tempSensorList[idx]
		obj, err := sMgr.GetTemperatureSensorState(tempSensorName)
		if err != nil {
			sMgr.logger.Err(fmt.Sprintln("Error getting the temp state for temp Sensor:", tempSensorName))
		}
		retObj.List = append(retObj.List, obj)
	}
	return &retObj, nil
}

func (sMgr *SensorManager) GetTemperatureSensorConfig(Name string) (*objects.TemperatureSensorConfig, error) {
	var tempSensorObj objects.TemperatureSensorConfig
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	tempSensorCfgEnt, exist := sMgr.tempConfigDB[Name]
	if !exist {
		return nil, errors.New("Invalid Temperature Sensor Name")
	}
	tempSensorObj.Name = Name
	tempSensorObj.AdminState = tempSensorCfgEnt.AdminState
	tempSensorObj.HigherAlarmThreshold = tempSensorCfgEnt.HigherAlarmThreshold
	tempSensorObj.HigherWarningThreshold = tempSensorCfgEnt.HigherWarningThreshold
	tempSensorObj.LowerAlarmThreshold = tempSensorCfgEnt.LowerAlarmThreshold
	tempSensorObj.LowerWarningThreshold = tempSensorCfgEnt.LowerWarningThreshold
	return &tempSensorObj, nil
}

func (sMgr *SensorManager) GetBulkTemperatureSensorConfig(fromIdx int, cnt int) (*objects.TemperatureSensorConfigGetInfo, error) {
	var retObj objects.TemperatureSensorConfigGetInfo
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	if fromIdx >= len(sMgr.tempSensorList) {
		return nil, errors.New("Invalid range")
	}
	if fromIdx+cnt > len(sMgr.tempSensorList) {
		retObj.EndIdx = len(sMgr.tempSensorList)
		retObj.More = false
		retObj.Count = 0
	} else {
		retObj.EndIdx = fromIdx + cnt
		retObj.More = true
		retObj.Count = len(sMgr.tempSensorList) - retObj.EndIdx + 1
	}
	for idx := fromIdx; idx < retObj.EndIdx; idx++ {
		tempSensorName := sMgr.tempSensorList[idx]
		obj, err := sMgr.GetTemperatureSensorConfig(tempSensorName)
		if err != nil {
			sMgr.logger.Err(fmt.Sprintln("Error getting the temp state for temp sensor:", tempSensorName))
		}
		retObj.List = append(retObj.List, obj)
	}
	return &retObj, nil
}

func genTempSensorUpdateMask(attrset []bool) uint32 {
	var mask uint32 = 0
	for idx, val := range attrset {
		if true == val {
			switch idx {
			case 0:
				//ObjKey Temp Sensor Name
			case 1:
				mask |= objects.TEMP_SENSOR_UPDATE_ADMIN_STATE
			case 2:
				mask |= objects.TEMP_SENSOR_UPDATE_HIGHER_ALARM_THRESHOLD
			case 3:
				mask |= objects.TEMP_SENSOR_UPDATE_HIGHER_WARN_THRESHOLD
			case 4:
				mask |= objects.TEMP_SENSOR_UPDATE_LOWER_WARN_THRESHOLD
			case 5:
				mask |= objects.TEMP_SENSOR_UPDATE_LOWER_ALARM_THRESHOLD
			}
		}
	}
	return mask
}

func (sMgr *SensorManager) UpdateTemperatureSensorConfig(oldCfg *objects.TemperatureSensorConfig, newCfg *objects.TemperatureSensorConfig, attrset []bool) (bool, error) {
	if sMgr.plugin == nil {
		return false, errors.New("Invalid platform plugin")
	}
	tempSensorCfgEnt, exist := sMgr.tempConfigDB[newCfg.Name]
	if !exist {
		return false, errors.New("Invalid TemperatureSensor Name")
	}
	mask := genTempSensorUpdateMask(attrset)
	if mask&objects.TEMP_SENSOR_UPDATE_ADMIN_STATE == objects.TEMP_SENSOR_UPDATE_ADMIN_STATE {
		tempSensorCfgEnt.AdminState = newCfg.AdminState
	}
	if mask&objects.TEMP_SENSOR_UPDATE_HIGHER_ALARM_THRESHOLD == objects.TEMP_SENSOR_UPDATE_HIGHER_ALARM_THRESHOLD {
		tempSensorCfgEnt.HigherAlarmThreshold = newCfg.HigherAlarmThreshold
	}
	if mask&objects.TEMP_SENSOR_UPDATE_HIGHER_WARN_THRESHOLD == objects.TEMP_SENSOR_UPDATE_HIGHER_WARN_THRESHOLD {
		tempSensorCfgEnt.HigherWarningThreshold = newCfg.HigherWarningThreshold
	}
	if mask&objects.TEMP_SENSOR_UPDATE_LOWER_ALARM_THRESHOLD == objects.TEMP_SENSOR_UPDATE_LOWER_ALARM_THRESHOLD {
		tempSensorCfgEnt.LowerAlarmThreshold = newCfg.LowerAlarmThreshold
	}
	if mask&objects.TEMP_SENSOR_UPDATE_LOWER_WARN_THRESHOLD == objects.TEMP_SENSOR_UPDATE_LOWER_WARN_THRESHOLD {
		tempSensorCfgEnt.LowerWarningThreshold = newCfg.LowerWarningThreshold
	}
	sMgr.tempConfigDB[newCfg.Name] = tempSensorCfgEnt
	return true, nil
}

func (sMgr *SensorManager) GetVoltageSensorState(Name string) (*objects.VoltageSensorState, error) {
	var voltageSensorObj objects.VoltageSensorState
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	_, exist := sMgr.voltageConfigDB[Name]
	if !exist {
		return nil, errors.New("Invalid Voltage Sensor Name")
	}

	err := sMgr.plugin.GetAllSensorState(sMgr.SensorState)
	if err != nil {
		return nil, err
	}
	voltageSensorState, _ := sMgr.SensorState.VoltageSensor[Name]
	voltageSensorObj.Name = Name
	voltageSensorObj.CurrentVoltage = voltageSensorState.Value
	return &voltageSensorObj, err
}

func (sMgr *SensorManager) GetBulkVoltageSensorState(fromIdx int, cnt int) (*objects.VoltageSensorStateGetInfo, error) {
	var retObj objects.VoltageSensorStateGetInfo
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	if fromIdx >= len(sMgr.voltageSensorList) {
		return nil, errors.New("Invalid range")
	}
	if fromIdx+cnt > len(sMgr.voltageSensorList) {
		retObj.EndIdx = len(sMgr.voltageSensorList)
		retObj.More = false
		retObj.Count = 0
	} else {
		retObj.EndIdx = fromIdx + cnt
		retObj.More = true
		retObj.Count = len(sMgr.voltageSensorList) - retObj.EndIdx + 1
	}
	for idx := fromIdx; idx < retObj.EndIdx; idx++ {
		voltageSensorName := sMgr.voltageSensorList[idx]
		obj, err := sMgr.GetVoltageSensorState(voltageSensorName)
		if err != nil {
			sMgr.logger.Err(fmt.Sprintln("Error getting the voltage state for voltage Sensor:", voltageSensorName))
		}
		retObj.List = append(retObj.List, obj)
	}
	return &retObj, nil
}

func (sMgr *SensorManager) GetVoltageSensorConfig(Name string) (*objects.VoltageSensorConfig, error) {
	var voltageSensorObj objects.VoltageSensorConfig
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	voltageSensorCfgEnt, exist := sMgr.voltageConfigDB[Name]
	if !exist {
		return nil, errors.New("Invalid Voltage Sensor Name")
	}
	voltageSensorObj.Name = Name
	voltageSensorObj.AdminState = voltageSensorCfgEnt.AdminState
	voltageSensorObj.HigherAlarmThreshold = voltageSensorCfgEnt.HigherAlarmThreshold
	voltageSensorObj.HigherWarningThreshold = voltageSensorCfgEnt.HigherWarningThreshold
	voltageSensorObj.LowerAlarmThreshold = voltageSensorCfgEnt.LowerAlarmThreshold
	voltageSensorObj.LowerWarningThreshold = voltageSensorCfgEnt.LowerWarningThreshold
	return &voltageSensorObj, nil
}

func (sMgr *SensorManager) GetBulkVoltageSensorConfig(fromIdx int, cnt int) (*objects.VoltageSensorConfigGetInfo, error) {
	var retObj objects.VoltageSensorConfigGetInfo
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	if fromIdx >= len(sMgr.voltageSensorList) {
		return nil, errors.New("Invalid range")
	}
	if fromIdx+cnt > len(sMgr.voltageSensorList) {
		retObj.EndIdx = len(sMgr.voltageSensorList)
		retObj.More = false
		retObj.Count = 0
	} else {
		retObj.EndIdx = fromIdx + cnt
		retObj.More = true
		retObj.Count = len(sMgr.voltageSensorList) - retObj.EndIdx + 1
	}
	for idx := fromIdx; idx < retObj.EndIdx; idx++ {
		voltageSensorName := sMgr.voltageSensorList[idx]
		obj, err := sMgr.GetVoltageSensorConfig(voltageSensorName)
		if err != nil {
			sMgr.logger.Err(fmt.Sprintln("Error getting the voltage state for voltage sensor:", voltageSensorName))
		}
		retObj.List = append(retObj.List, obj)
	}
	return &retObj, nil
}

func genVoltageSensorUpdateMask(attrset []bool) uint32 {
	var mask uint32 = 0
	for idx, val := range attrset {
		if true == val {
			switch idx {
			case 0:
				//ObjKey Voltage Sensor Name
			case 1:
				mask |= objects.VOLTAGE_SENSOR_UPDATE_ADMIN_STATE
			case 2:
				mask |= objects.VOLTAGE_SENSOR_UPDATE_HIGHER_ALARM_THRESHOLD
			case 3:
				mask |= objects.VOLTAGE_SENSOR_UPDATE_HIGHER_WARN_THRESHOLD
			case 4:
				mask |= objects.VOLTAGE_SENSOR_UPDATE_LOWER_WARN_THRESHOLD
			case 5:
				mask |= objects.VOLTAGE_SENSOR_UPDATE_LOWER_ALARM_THRESHOLD
			}
		}
	}
	return mask
}

func (sMgr *SensorManager) UpdateVoltageSensorConfig(oldCfg *objects.VoltageSensorConfig, newCfg *objects.VoltageSensorConfig, attrset []bool) (bool, error) {
	if sMgr.plugin == nil {
		return false, errors.New("Invalid platform plugin")
	}
	voltageSensorCfgEnt, exist := sMgr.voltageConfigDB[newCfg.Name]
	if !exist {
		return false, errors.New("Invalid VoltageSensor Name")
	}
	mask := genVoltageSensorUpdateMask(attrset)
	if mask&objects.VOLTAGE_SENSOR_UPDATE_ADMIN_STATE == objects.VOLTAGE_SENSOR_UPDATE_ADMIN_STATE {
		voltageSensorCfgEnt.AdminState = newCfg.AdminState
	}
	if mask&objects.VOLTAGE_SENSOR_UPDATE_HIGHER_ALARM_THRESHOLD == objects.VOLTAGE_SENSOR_UPDATE_HIGHER_ALARM_THRESHOLD {
		voltageSensorCfgEnt.HigherAlarmThreshold = newCfg.HigherAlarmThreshold
	}
	if mask&objects.VOLTAGE_SENSOR_UPDATE_HIGHER_WARN_THRESHOLD == objects.VOLTAGE_SENSOR_UPDATE_HIGHER_WARN_THRESHOLD {
		voltageSensorCfgEnt.HigherWarningThreshold = newCfg.HigherWarningThreshold
	}
	if mask&objects.VOLTAGE_SENSOR_UPDATE_LOWER_ALARM_THRESHOLD == objects.VOLTAGE_SENSOR_UPDATE_LOWER_ALARM_THRESHOLD {
		voltageSensorCfgEnt.LowerAlarmThreshold = newCfg.LowerAlarmThreshold
	}
	if mask&objects.VOLTAGE_SENSOR_UPDATE_LOWER_WARN_THRESHOLD == objects.VOLTAGE_SENSOR_UPDATE_LOWER_WARN_THRESHOLD {
		voltageSensorCfgEnt.LowerWarningThreshold = newCfg.LowerWarningThreshold
	}
	sMgr.voltageConfigDB[newCfg.Name] = voltageSensorCfgEnt

	return true, nil
}

func (sMgr *SensorManager) GetPowerConverterSensorState(Name string) (*objects.PowerConverterSensorState, error) {
	var powerConverterSensorObj objects.PowerConverterSensorState
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	_, exist := sMgr.powerConverterConfigDB[Name]
	if !exist {
		return nil, errors.New("Invalid PowerConverter Sensor Name")
	}

	err := sMgr.plugin.GetAllSensorState(sMgr.SensorState)
	if err != nil {
		return nil, err
	}
	powerConverterSensorState, _ := sMgr.SensorState.PowerConverterSensor[Name]
	powerConverterSensorObj.Name = Name
	powerConverterSensorObj.CurrentPower = powerConverterSensorState.Value
	return &powerConverterSensorObj, err
}

func (sMgr *SensorManager) GetBulkPowerConverterSensorState(fromIdx int, cnt int) (*objects.PowerConverterSensorStateGetInfo, error) {
	var retObj objects.PowerConverterSensorStateGetInfo
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	if fromIdx >= len(sMgr.powerConverterSensorList) {
		return nil, errors.New("Invalid range")
	}
	if fromIdx+cnt > len(sMgr.powerConverterSensorList) {
		retObj.EndIdx = len(sMgr.powerConverterSensorList)
		retObj.More = false
		retObj.Count = 0
	} else {
		retObj.EndIdx = fromIdx + cnt
		retObj.More = true
		retObj.Count = len(sMgr.powerConverterSensorList) - retObj.EndIdx + 1
	}
	for idx := fromIdx; idx < retObj.EndIdx; idx++ {
		powerConverterSensorName := sMgr.powerConverterSensorList[idx]
		obj, err := sMgr.GetPowerConverterSensorState(powerConverterSensorName)
		if err != nil {
			sMgr.logger.Err(fmt.Sprintln("Error getting the powerConverter state for powerConverter Sensor:", powerConverterSensorName))
		}
		retObj.List = append(retObj.List, obj)
	}
	return &retObj, nil
}

func (sMgr *SensorManager) GetPowerConverterSensorConfig(Name string) (*objects.PowerConverterSensorConfig, error) {
	var powerConverterSensorObj objects.PowerConverterSensorConfig
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	powerConverterSensorCfgEnt, exist := sMgr.powerConverterConfigDB[Name]
	if !exist {
		return nil, errors.New("Invalid PowerConverter Sensor Name")
	}
	powerConverterSensorObj.Name = Name
	powerConverterSensorObj.AdminState = powerConverterSensorCfgEnt.AdminState
	powerConverterSensorObj.HigherAlarmThreshold = powerConverterSensorCfgEnt.HigherAlarmThreshold
	powerConverterSensorObj.HigherWarningThreshold = powerConverterSensorCfgEnt.HigherWarningThreshold
	powerConverterSensorObj.LowerAlarmThreshold = powerConverterSensorCfgEnt.LowerAlarmThreshold
	powerConverterSensorObj.LowerWarningThreshold = powerConverterSensorCfgEnt.LowerWarningThreshold
	return &powerConverterSensorObj, nil
}

func (sMgr *SensorManager) GetBulkPowerConverterSensorConfig(fromIdx int, cnt int) (*objects.PowerConverterSensorConfigGetInfo, error) {
	var retObj objects.PowerConverterSensorConfigGetInfo
	if sMgr.plugin == nil {
		return nil, errors.New("Invalid platform plugin")
	}
	if fromIdx >= len(sMgr.powerConverterSensorList) {
		return nil, errors.New("Invalid range")
	}
	if fromIdx+cnt > len(sMgr.powerConverterSensorList) {
		retObj.EndIdx = len(sMgr.powerConverterSensorList)
		retObj.More = false
		retObj.Count = 0
	} else {
		retObj.EndIdx = fromIdx + cnt
		retObj.More = true
		retObj.Count = len(sMgr.powerConverterSensorList) - retObj.EndIdx + 1
	}
	for idx := fromIdx; idx < retObj.EndIdx; idx++ {
		powerConverterSensorName := sMgr.powerConverterSensorList[idx]
		obj, err := sMgr.GetPowerConverterSensorConfig(powerConverterSensorName)
		if err != nil {
			sMgr.logger.Err(fmt.Sprintln("Error getting the powerConverter state for powerConverter sensor:", powerConverterSensorName))
		}
		retObj.List = append(retObj.List, obj)
	}
	return &retObj, nil
}

func genPowerConverterSensorUpdateMask(attrset []bool) uint32 {
	var mask uint32 = 0
	for idx, val := range attrset {
		if true == val {
			switch idx {
			case 0:
				//ObjKey PowerConverter Sensor Name
			case 1:
				mask |= objects.POWER_CONVERTER_SENSOR_UPDATE_ADMIN_STATE
			case 2:
				mask |= objects.POWER_CONVERTER_SENSOR_UPDATE_HIGHER_ALARM_THRESHOLD
			case 3:
				mask |= objects.POWER_CONVERTER_SENSOR_UPDATE_HIGHER_WARN_THRESHOLD
			case 4:
				mask |= objects.POWER_CONVERTER_SENSOR_UPDATE_LOWER_WARN_THRESHOLD
			case 5:
				mask |= objects.POWER_CONVERTER_SENSOR_UPDATE_LOWER_ALARM_THRESHOLD
			}
		}
	}
	return mask
}

func (sMgr *SensorManager) UpdatePowerConverterSensorConfig(oldCfg *objects.PowerConverterSensorConfig, newCfg *objects.PowerConverterSensorConfig, attrset []bool) (bool, error) {
	if sMgr.plugin == nil {
		return false, errors.New("Invalid platform plugin")
	}
	powerConverterSensorCfgEnt, exist := sMgr.powerConverterConfigDB[newCfg.Name]
	if !exist {
		return false, errors.New("Invalid PowerConverterSensor Name")
	}
	mask := genPowerConverterSensorUpdateMask(attrset)
	if mask&objects.POWER_CONVERTER_SENSOR_UPDATE_ADMIN_STATE == objects.POWER_CONVERTER_SENSOR_UPDATE_ADMIN_STATE {
		powerConverterSensorCfgEnt.AdminState = newCfg.AdminState
	}
	if mask&objects.POWER_CONVERTER_SENSOR_UPDATE_HIGHER_ALARM_THRESHOLD == objects.POWER_CONVERTER_SENSOR_UPDATE_HIGHER_ALARM_THRESHOLD {
		powerConverterSensorCfgEnt.HigherAlarmThreshold = newCfg.HigherAlarmThreshold
	}
	if mask&objects.POWER_CONVERTER_SENSOR_UPDATE_HIGHER_WARN_THRESHOLD == objects.POWER_CONVERTER_SENSOR_UPDATE_HIGHER_WARN_THRESHOLD {
		powerConverterSensorCfgEnt.HigherWarningThreshold = newCfg.HigherWarningThreshold
	}
	if mask&objects.POWER_CONVERTER_SENSOR_UPDATE_LOWER_ALARM_THRESHOLD == objects.POWER_CONVERTER_SENSOR_UPDATE_LOWER_ALARM_THRESHOLD {
		powerConverterSensorCfgEnt.LowerAlarmThreshold = newCfg.LowerAlarmThreshold
	}
	if mask&objects.POWER_CONVERTER_SENSOR_UPDATE_LOWER_WARN_THRESHOLD == objects.POWER_CONVERTER_SENSOR_UPDATE_LOWER_WARN_THRESHOLD {
		powerConverterSensorCfgEnt.LowerWarningThreshold = newCfg.LowerWarningThreshold
	}
	sMgr.powerConverterConfigDB[newCfg.Name] = powerConverterSensorCfgEnt

	return true, nil
}
