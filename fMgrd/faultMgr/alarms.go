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

package faultMgr

import (
	"fmt"
	//"errors"
	"infra/fMgrd/objects"
	"time"
	"utils/eventUtils"
)

func (fMgr *FaultManager) GetBulkAlarmState(fromIdx int, count int) (*objects.AlarmStateGetInfo, error) {
	fMgr.logger.Info("Inside GetBulkAlarmState ...")
	var retObj objects.AlarmStateGetInfo

	alarms := fMgr.AlarmRB.GetListOfEntriesFromRingBuffer()
	length := len(alarms)
	aState := make([]objects.AlarmState, count)

	var i int
	var j int

	for i, j = 0, fromIdx; i < count && j < length; j++ {
		aIntf := alarms[length-j-1]
		alarm := aIntf.(AlarmRBEntry)
		var aObj objects.AlarmState
		aObj.OwnerId = int32(alarm.OwnerId)
		aObj.EventId = int32(alarm.EventId)
		evtKey := EventKey{
			DaemonId: alarm.OwnerId,
			EventId:  alarm.EventId,
		}
		fEnt, exist := fMgr.FaultEventMap[evtKey]
		if !exist {
			j--
			continue
		}
		aObj.OwnerName = fEnt.FaultOwnerName
		aObj.EventName = fEnt.FaultEventName
		aObj.SrcObjName = fEnt.FaultSrcObjName
		aObj.Severity = fEnt.AlarmSeverity
		aObj.Description = alarm.Description
		aObj.OccuranceTime = alarm.OccuranceTime.String()
		aObj.SrcObjKey = fmt.Sprintf("%v", alarm.SrcObjKey)
		if alarm.Resolved == true {
			aObj.ResolutionTime = alarm.ResolutionTime.String()
		} else {
			aObj.ResolutionTime = "N/A"
		}
		aState[i] = aObj
		i++
	}
	retObj.EndIdx = j
	retObj.Count = i
	if j != length {
		retObj.More = true
	}
	retObj.List = aState

	return &retObj, nil
}

func (fMgr *FaultManager) StartAlarmTimer(evt eventUtils.Event) *time.Timer {
	evtKey := EventKey{
		DaemonId: int(evt.OwnerId),
		EventId:  int(evt.EvtId),
	}

	alarmFunc := func() {
		if fMgr.AlarmMap[evtKey] == nil {
			fMgr.logger.Info("Alarm Database does not exist, hence skipping alarm generation")
			fMgr.AlarmMap[evtKey] = make(map[FaultObjKey]AlarmData)
		}

		aDataMapEnt, _ := fMgr.AlarmMap[evtKey]
		fObjKey := generateFaultObjKey(evt.OwnerName, evt.SrcObjName, evt.SrcObjKey)
		if fObjKey == "" {
			fMgr.logger.Info("Fault Obj key, hence skipping alarm generation")
			return
		}
		aDataEnt, exist := aDataMapEnt[fObjKey]
		if exist {
			fMgr.logger.Info("Alarm Data entry already exist, hence skipping this")
			return
		}
		aDataEnt.AlarmListIdx = fMgr.AddAlarmEntryInRB(evt)
		aDataEnt.AlarmSeqNumber = fMgr.AlarmSeqNumber
		fMgr.AlarmSeqNumber++
		aDataMapEnt[fObjKey] = aDataEnt
		fMgr.AlarmMap[evtKey] = aDataMapEnt
	}

	return time.AfterFunc(fMgr.FaultToAlarmTransitionTime, alarmFunc)
}

func (fMgr *FaultManager) AddAlarmEntryInRB(evt eventUtils.Event) int {
	aRBEnt := AlarmRBEntry{
		OwnerId:        int(evt.OwnerId),
		EventId:        int(evt.EvtId),
		OccuranceTime:  time.Now(),
		SrcObjKey:      evt.SrcObjKey,
		AlarmSeqNumber: fMgr.AlarmSeqNumber,
	}

	return fMgr.AlarmRB.InsertIntoRingBuffer(aRBEnt)
}

func (fMgr *FaultManager) StartAlarmRemoveTimer(evt eventUtils.Event) *time.Timer {
	evtKey := EventKey{
		DaemonId: int(evt.OwnerId),
		EventId:  int(evt.EvtId),
	}

	cFEnt, exist := fMgr.NonFaultEventMap[evtKey]
	if !exist {
		fMgr.logger.Info("Error finding the fault for fault clearing event")
		return nil
	}
	fEvtKey := EventKey{
		DaemonId: cFEnt.FaultOwnerId,
		EventId:  cFEnt.FaultEventId,
	}

	fObjKey := generateFaultObjKey(evt.OwnerName, evt.SrcObjName, evt.SrcObjKey)
	if fObjKey == "" {
		fMgr.logger.Info("Error generating fault object key")
		return nil
	}

	alarmFunc := func() {
		aDataMapEnt, exist := fMgr.AlarmMap[fEvtKey]
		if !exist {
			fMgr.logger.Info("Alarm Database does not exist, hence skipping removal of Alarm")
			return
		}
		fObjKey := generateFaultObjKey(evt.OwnerName, evt.SrcObjName, evt.SrcObjKey)
		if fObjKey == "" {
			fMgr.logger.Info("Fault Obj key, hence skipping alarm removal")
			return
		}
		aDataEnt, exist := aDataMapEnt[fObjKey]
		if !exist {
			fMgr.logger.Info("Alarm Data entry doesnot exist, hence skipping this")
			return
		}
		aIntf := fMgr.AlarmRB.GetEntryFromRingBuffer(aDataEnt.AlarmListIdx)
		aRBData := aIntf.(AlarmRBEntry)
		if aRBData.AlarmSeqNumber == aDataEnt.AlarmSeqNumber {
			aRBData.ResolutionTime = time.Now()
			fMgr.AlarmRB.UpdateEntryInRingBuffer(aRBData, aDataEnt.AlarmListIdx)
			delete(aDataMapEnt, fObjKey)
			fMgr.AlarmMap[fEvtKey] = aDataMapEnt
		}
	}

	return time.AfterFunc(fMgr.AlarmTransitionTime, alarmFunc)
}
