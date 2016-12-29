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

package objects

const (
	ADMIN_STATE_UP   = "UP"
	ADMIN_STATE_DOWN = "DOWN"
)

const (
	SFLOW_GLOBAL_ATTR_ADMIN_STATE_IDX           = 0x1
	SFLOW_GLOBAL_ATTR_AGENT_IPADDR_IDX          = 0x2
	SFLOW_GLOBAL_ATTR_MAX_SAMPLE_SIZE_IDX       = 0x3
	SFLOW_GLOBAL_ATTR_COUNTER_POLL_INTERVAL_IDX = 0x4
	SFLOW_GLOBAL_ATTR_MAX_DATAGRAM_SIZE_IDX     = 0x5

	SFLOW_GLOBAL_UPDATE_ATTR_ADMIN_STATE           = 0x1
	SFLOW_GLOBAL_UPDATE_ATTR_AGENT_IPADDR          = 0x2
	SFLOW_GLOBAL_UPDATE_ATTR_MAX_SAMPLE_SIZE       = 0x4
	SFLOW_GLOBAL_UPDATE_ATTR_COUNTER_POLL_INTERVAL = 0x8
	SFLOW_GLOBAL_UPDATE_ATTR_MAX_DATAGRAM_SIZE     = 0x10
)

type SflowGlobal struct {
	Vrf                 string
	AdminState          string
	AgentIpAddr         string
	MaxSampledSize      int32
	CounterPollInterval int32
	MaxDatagramSize     int32
}

const (
	SFLOW_COLLECTOR_ATTR_UDP_PORT_IDX    = 0x1
	SFLOW_COLLECTOR_ATTR_ADMIN_STATE_IDX = 0x2

	SFLOW_COLLECTOR_UPDATE_ATTR_UDP_PORT    = 0x1
	SFLOW_COLLECTOR_UPDATE_ATTR_ADMIN_STATE = 0x2
)

type SflowCollector struct {
	IpAddr     string
	UdpPort    int32
	AdminState string
}

type SflowCollectorState struct {
	IpAddr                  string
	OperState               string
	NumSflowSamplesExported int32
	NumDatagramExported     int32
}

type SflowCollectorStateGetInfo struct {
	EndIdx int
	Count  int
	More   bool
	List   []*SflowCollectorState
}

const (
	SFLOW_INTF_ATTR_ADMIN_STATE_IDX   = 0x1
	SFLOW_INTF_ATTR_SAMPLING_RATE_IDX = 0x2

	SFLOW_INTF_UPDATE_ATTR_ADMIN_STATE   = 0x1
	SFLOW_INTF_UPDATE_ATTR_SAMPLING_RATE = 0x2
)

type SflowIntf struct {
	IntfRef      string
	AdminState   string
	SamplingRate int32
}

type SflowIntfState struct {
	IntfRef                 string
	OperState               string
	NumSflowSamplesExported int32
}

type SflowIntfStateGetInfo struct {
	EndIdx int
	Count  int
	More   bool
	List   []*SflowIntfState
}

type SflowNetdevInfo struct {
	IfIndex    int32
	IntfRef    string
	NetDevName string
}

type SflowIntfCounterInfo struct {
	OperState         bool
	IfInOctets        uint64
	IfInUcastPkts     uint64
	IfInMcastPkts     uint64
	IfInBcastPkts     uint64
	IfInDiscards      uint64
	IfInErrors        uint64
	IfInUnknownProtos uint64
	IfOutOctets       uint64
	IfOutUcastPkts    uint64
	IfOutMcastPkts    uint64
	IfOutBcastPkts    uint64
	IfOutDiscards     uint64
	IfOutErrors       uint64
}

type SflowIntfCfgInfo struct {
	Speed      int32
	FullDuplex bool
}
