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
//   This is a auto-generated file, please do not edit!
// _______   __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----  \   \/    \/   /  |  |  ---|  |----    ,---- |  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |        |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |        `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//

package rpc

import (
	"infra/statsd/objects"
	"statsd"
)

func convertFromRPCFmtSflowGlobal(obj *statsd.SflowGlobal) *objects.SflowGlobal {
	return &objects.SflowGlobal{
		Vrf:                 obj.Vrf,
		AdminState:          obj.AdminState,
		AgentIpAddr:         obj.AgentIpAddr,
		MaxSampledSize:      obj.MaxSampledSize,
		CounterPollInterval: obj.CounterPollInterval,
	}
}

func convertFromRPCFmtSflowCollector(obj *statsd.SflowCollector) *objects.SflowCollector {
	return &objects.SflowCollector{
		IpAddr:          obj.IpAddr,
		UdpPort:         obj.UdpPort,
		AdminState:      obj.AdminState,
		MaxDatagramSize: obj.MaxDatagramSize,
	}
}

func convertFromRPCFmtSflowIntf(obj *statsd.SflowIntf) *objects.SflowIntf {
	return &objects.SflowIntf{
		IntfRef:      obj.IntfRef,
		AdminState:   obj.AdminState,
		SamplingRate: obj.SamplingRate,
	}
}

func convertToRPCFmtSflowCollectorState(obj *objects.SflowCollectorState) *statsd.SflowCollectorState {
	return &statsd.SflowCollectorState{
		IpAddr:                  obj.IpAddr,
		OperState:               obj.OperState,
		NumSflowSamplesExported: obj.NumSflowSamplesExported,
		NumDatagramExported:     obj.NumDatagramExported,
	}
}

func convertToRPCFmtSflowIntfState(obj *objects.SflowIntfState) *statsd.SflowIntfState {
	return &statsd.SflowIntfState{
		IntfRef:            obj.IntfRef,
		OperState:          obj.OperState,
		NumRecordsExported: obj.NumRecordsExported,
	}
}
