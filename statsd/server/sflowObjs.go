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

package server

import (
	"infra/statsd/objects"
	"sync"
)

//Server op code definitions
const (
	CREATE_SFLOW_GLOBAL ServerOpId = iota
	UPDATE_SFLOW_GLOBAL
	DELETE_SFLOW_GLOBAL
	CREATE_SFLOW_COLLECTOR
	UPDATE_SFLOW_COLLECTOR
	DELETE_SFLOW_COLLECTOR
	GET_SFLOW_COLLECTOR_STATE
	GET_BULK_SFLOW_COLLECTOR_STATE
	CREATE_SFLOW_INTF
	UPDATE_SFLOW_INTF
	DELETE_SFLOW_INTF
	GET_SFLOW_INTF_STATE
	GET_BULK_SFLOW_INTF_STATE
)

type CreateSflowGlobalInArgs struct {
	Obj *objects.SflowGlobal
}

type UpdateSflowGlobalInArgs struct {
	OldObj  *objects.SflowGlobal
	NewObj  *objects.SflowGlobal
	AttrSet []bool
}

type DeleteSflowGlobalInArgs struct {
	Obj *objects.SflowGlobal
}

type CreateSflowCollectorInArgs struct {
	Obj *objects.SflowCollector
}

type UpdateSflowCollectorInArgs struct {
	OldObj  *objects.SflowCollector
	NewObj  *objects.SflowCollector
	AttrSet []bool
}

type DeleteSflowCollectorInArgs struct {
	Obj *objects.SflowCollector
}

type GetSflowCollectorStateInArgs struct {
	IpAddr string
}

type GetSflowCollectorStateOutArgs struct {
	Obj *objects.SflowCollectorState
	Err error
}

type GetBulkInArgs struct {
	FromIdx int
	Count   int
}

type GetBulkSflowCollectorStateOutArgs struct {
	BulkObj *objects.SflowCollectorStateGetInfo
	Err     error
}

type CreateSflowIntfInArgs struct {
	Obj *objects.SflowIntf
}

type UpdateSflowIntfInArgs struct {
	OldObj  *objects.SflowIntf
	NewObj  *objects.SflowIntf
	AttrSet []bool
}

type DeleteSflowIntfInArgs struct {
	Obj *objects.SflowIntf
}

type GetSflowIntfStateInArgs struct {
	IntfRef string
}

type GetSflowIntfStateOutArgs struct {
	Obj *objects.SflowIntfState
	Err error
}

type GetBulkSflowIntfStateOutArgs struct {
	BulkObj *objects.SflowIntfStateGetInfo
	Err     error
}

//Internal server struct definitions
type sflowGbl struct {
	vrf                 string
	adminState          string
	agentIpAddr         string
	maxSampledSize      int32
	counterPollInterval int32
	maxDatagramSize     int32
}

type sflowCollector struct {
	ipAddr                  string
	udpPort                 int32
	adminState              string
	operstate               string
	numSflowSamplesExported int32
	numDatagramExported     int32
	shutdownCh              chan bool
	dgramRcvCh              chan *sflowDgramIdx
}

type sflowIntf struct {
	ifIndex                 int32
	adminState              string
	operstate               string
	samplingRate            int32
	operState               string
	numSflowSamplesExported int32
	shutdownCh              chan bool
}

//Typedef for db keys
type netDevData struct {
	intfRef string
}

type sflowRecordInfo struct {
	ifIndex   int32
	sflowData []byte
}

type sflowDgramIdx struct {
	ifIndex int32
	key     int32
}

type sflowDgram struct {
}

type sflowServer struct {
	dbMutex                  sync.RWMutex
	netDevInfo               map[int32]netDevData
	sflowGblDB               *sflowGbl
	sflowCollectorDB         map[string]*sflowCollector
	sflowCollectorDBKeyCache []string
	sflowIntfDB              map[int32]*sflowIntf
	sflowIntfDBKeyCache      []int32
	//Data store for Sflow encoded datagrams, keyed by interface and running key count
	sflowDgramDB map[int32]map[int32]sflowDgram
	//Channel for interface poller to send sflow records to Dgram processor
	sflowIntfRecordCh chan sflowRecordInfo
	//Channel to send Dgram ready for dispatch
	sflowDgramRdy chan sflowDgramIdx
	//Channel list to send sflowRecord to collector
	sflowDgramToCollector map[string]chan sflowDgramIdx
	//Response channel from collector routines
	sflowDgramSentReceipt chan sflowDgramIdx
}
