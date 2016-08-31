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

package objects

type QsfpState struct {
	Location           string
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

type QsfpStateGetInfo struct {
	EndIdx int
	Count  int
	More   bool
	List   []*QsfpState
}

type QsfpConfig struct {
	Location                 string
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

type QsfpConfigGetInfo struct {
	EndIdx int
	Count  int
	More   bool
	List   []*QsfpConfig
}
