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
