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

package server

import (
	"infra/statsd/objects"
)

//Sflow global object functions
func (srvr *DmnServer) ValidateCreateSflowGlobal(obj *objects.SflowGlobal) (bool, error) {
	return true, nil
}

func (srvr *DmnServer) createSflowGlobal(obj *objects.SflowGlobal) {
}

func (srvr *DmnServer) ValidateUpdateSflowGlobal(oldObj, newObj *objects.SflowGlobal, attrset []bool) (bool, error) {
	return true, nil
}

func (srvr *DmnServer) updateSflowGlobal(oldObj, newObj *objects.SflowGlobal, attrset []bool) {
}

func (srvr *DmnServer) ValidateDeleteSflowGlobal(obj *objects.SflowGlobal) (bool, error) {
	return true, nil
}

func (srvr *DmnServer) deleteSflowGlobal(obj *objects.SflowGlobal) {
}

//Sflow collector related functions
func (srvr *DmnServer) ValidateCreateSflowCollector(obj *objects.SflowCollector) (bool, error) {
	return true, nil
}

func (srvr *DmnServer) createSflowCollector(obj *objects.SflowCollector) {
}

func (srvr *DmnServer) ValidateUpdateSflowCollector(oldObj, newObj *objects.SflowCollector, attrset []bool) (bool, error) {
	return true, nil
}

func (srvr *DmnServer) updateSflowCollector(oldObj, newObj *objects.SflowCollector, attrset []bool) {
}

func (srvr *DmnServer) ValidateDeleteSflowCollector(obj *objects.SflowCollector) (bool, error) {
	return true, nil
}

func (srvr *DmnServer) deleteSflowCollector(obj *objects.SflowCollector) {
}

//Sflow interface related functions
func (srvr *DmnServer) ValidateCreateSflowIntf(obj *objects.SflowIntf) (bool, error) {
	return true, nil
}

func (srvr *DmnServer) createSflowIntf(obj *objects.SflowIntf) {
}

func (srvr *DmnServer) ValidateUpdateSflowIntf(oldObj, newObj *objects.SflowIntf, attrset []bool) (bool, error) {
	return true, nil
}

func (srvr *DmnServer) updateSflowIntf(oldObj, newObj *objects.SflowIntf, attrset []bool) {
}

func (srvr *DmnServer) ValidateDeleteSflowIntf(obj *objects.SflowIntf) (bool, error) {
	return true, nil
}

func (srvr *DmnServer) deleteSflowIntf(obj *objects.SflowIntf) {
}

//Sflow collector state object functions
func (srvr *DmnServer) getSflowCollectorState(ipAddr string) (*objects.SflowCollectorState, error) {
	return nil, nil
}

func (srvr *DmnServer) getBulkSflowCollectorState(fromIdx, count int) (*objects.SflowCollectorStateGetInfo, error) {
	return nil, nil
}

//Sflow interface state object functions
func (srvr *DmnServer) getSflowIntfState(intfRef string) (*objects.SflowIntfState, error) {
	return nil, nil
}

func (srvr *DmnServer) getBulkSflowIntfState(fromIdx, count int) (*objects.SflowIntfStateGetInfo, error) {
	return nil, nil
}
