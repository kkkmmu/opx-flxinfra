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
	"encoding/json"
	"fmt"
	"models/events"
	"sync"
	"time"
	"utils/eventUtils"
	"utils/logging"
	"utils/ringBuffer"
)

type EventKey struct {
	DaemonId int
	EventId  int
}

type NonFaultDetail struct {
	IsClearingEvent bool
	FaultEventId    int
	FaultOwnerId    int
}

type FaultDetail struct {
	RaiseFault       bool
	ClearingEventId  int
	ClearingDaemonId int
	AlarmSeverity    string
	FaultOwnerName   string
	FaultEventName   string
	FaultSrcObjName  string
}

type EvtDetail struct {
	IsClearingEvent  bool
	IsFault          bool
	RaiseFault       bool
	ClearingEventId  int
	ClearingDaemonId int
	AlarmSeverity    string
	OwnerName        string
	EventName        string
	SrcObjName       string
}

type FaultRBEntry struct {
	OwnerId        int
	EventId        int
	ResolutionTime time.Time
	OccuranceTime  time.Time
	SrcObjKey      interface{}
	FaultSeqNumber uint64
	Description    string
	Resolved       bool
}

type AlarmRBEntry struct {
	OwnerId        int
	EventId        int
	ResolutionTime time.Time
	OccuranceTime  time.Time
	SrcObjKey      interface{}
	AlarmSeqNumber uint64
	Description    string
	Resolved       bool
}

type FaultData struct {
	FaultListIdx int
	//AlarmListIdx     int
	CreateAlarmTimer *time.Timer
	FaultSeqNumber   uint64
}

type AlarmData struct {
	AlarmListIdx     int
	AlarmSeqNumber   uint64
	RemoveAlarmTimer *time.Timer
}

type FaultObjKey string
type FaultDataMap map[FaultObjKey]FaultData
type AlarmDataMap map[FaultObjKey]AlarmData

type FaultManager struct {
	logger                     logging.LoggerIntf
	EventCh                    chan []byte
	FaultEventMap              map[EventKey]FaultDetail
	NonFaultEventMap           map[EventKey]NonFaultDetail
	FMapRWMutex                sync.RWMutex
	FaultMap                   map[EventKey]FaultDataMap
	AMapRWMutex                sync.RWMutex
	AlarmMap                   map[EventKey]AlarmDataMap
	FRBRWMutex                 sync.RWMutex
	FaultRB                    *ringBuffer.RingBuffer
	ARBRWMutex                 sync.RWMutex
	AlarmRB                    *ringBuffer.RingBuffer
	DaemonList                 []string
	FaultSeqNumber             uint64
	AlarmSeqNumber             uint64
	FaultToAlarmTransitionTime time.Duration
	AlarmTransitionTime        time.Duration
}

func NewFaultManager(logger logging.LoggerIntf) *FaultManager {
	fMgr := &FaultManager{}
	fMgr.logger = logger
	fMgr.EventCh = make(chan []byte, 1000)
	fMgr.FaultEventMap = make(map[EventKey]FaultDetail)
	fMgr.NonFaultEventMap = make(map[EventKey]NonFaultDetail)
	fMgr.FaultMap = make(map[EventKey]FaultDataMap) //Existing Faults
	fMgr.AlarmMap = make(map[EventKey]AlarmDataMap) //Existing Alarm
	fMgr.FaultRB = new(ringBuffer.RingBuffer)
	fMgr.FaultRB.SetRingBufferCapacity(100000) // Max 100000 entries in fault database
	fMgr.AlarmRB = new(ringBuffer.RingBuffer)
	fMgr.AlarmRB.SetRingBufferCapacity(100000) // Max 100000 entries in alarm database
	fMgr.FaultSeqNumber = 0
	fMgr.AlarmSeqNumber = 0
	fMgr.FaultToAlarmTransitionTime = time.Duration(3) * time.Second
	fMgr.AlarmTransitionTime = time.Duration(3) * time.Second
	return fMgr
}

func (fMgr *FaultManager) InitFaultManager() error {
	err := fMgr.initFMgrDS()
	if err != nil {
		fMgr.logger.Err(fmt.Sprintln("Error Initializing Fault Manager DS:", err))
		return err
	}
	go fMgr.EventProcessor()
	return err
}

func (fMgr *FaultManager) EventProcessor() {
	for {
		msg := <-fMgr.EventCh
		var evt eventUtils.Event
		err := json.Unmarshal(msg, &evt)
		if err != nil {
			fMgr.logger.Err(fmt.Sprintln("Unable to Unmarshal the byte stream", err))
			continue
		}
		fMgr.logger.Debug(fmt.Sprintln("OwnerId:", evt.OwnerId))
		fMgr.logger.Debug(fmt.Sprintln("OwnerName:", evt.OwnerName))
		fMgr.logger.Debug(fmt.Sprintln("EvtId:", evt.EvtId))
		fMgr.logger.Debug(fmt.Sprintln("EventName:", evt.EventName))
		fMgr.logger.Debug(fmt.Sprintln("Timestamp:", evt.TimeStamp))
		fMgr.logger.Debug(fmt.Sprintln("Description:", evt.Description))
		fMgr.logger.Debug(fmt.Sprintln("SrcObjName:", evt.SrcObjName))
		keyMap, exist := events.EventKeyMap[evt.OwnerName]
		if !exist {
			fMgr.logger.Err("Key map not found given Event")
			continue
		}
		obj, exist := keyMap[evt.SrcObjName]
		if !exist {
			fMgr.logger.Err("Src Object Name not found in Event Key Map")
			continue
		}
		obj = evt.SrcObjKey
		fMgr.logger.Debug(fmt.Sprintln("Src Obj Key", obj))

		fMgr.processEvents(evt)
	}
}

func (fMgr *FaultManager) initFMgrDS() error {
	evtMap := make(map[EventKey]EvtDetail)
	evtJson, err := eventUtils.ParseEventsJson()
	if err != nil {
		fMgr.logger.Err(fmt.Sprintln("Error Parsing the events.json", err))
		return err
	}
	for _, daemon := range evtJson.DaemonEvents {
		fMgr.logger.Debug(fmt.Sprintln("daemon.DaemonName:", daemon.DaemonName))
		fMgr.DaemonList = append(fMgr.DaemonList, daemon.DaemonName)
		for _, evt := range daemon.EventList {
			fId := EventKey{
				DaemonId: int(daemon.DaemonId),
				EventId:  int(evt.EventId),
			}
			evtEnt, exist := evtMap[fId]
			if exist {
				fMgr.logger.Err(fmt.Sprintln("Duplicate entry found"))
				continue
			}
			if evt.IsFault == true {
				evtEnt.IsFault = true
				evtEnt.IsClearingEvent = false
				evtEnt.RaiseFault = evt.Fault.RaiseFault
				evtEnt.ClearingEventId = evt.Fault.ClearingEventId
				evtEnt.ClearingDaemonId = evt.Fault.ClearingDaemonId
				evtEnt.AlarmSeverity = evt.Fault.AlarmSeverity
				evtEnt.OwnerName = daemon.DaemonName
				evtEnt.EventName = evt.EventName
				evtEnt.SrcObjName = evt.SrcObjName
			} else {
				evtEnt.IsFault = false
				evtEnt.IsClearingEvent = false
				evtEnt.RaiseFault = false
				evtEnt.ClearingEventId = -1
				evtEnt.ClearingDaemonId = -1
				evtEnt.AlarmSeverity = ""
				evtEnt.OwnerName = daemon.DaemonName
				evtEnt.EventName = evt.EventName
				evtEnt.SrcObjName = evt.SrcObjName
			}
			evtMap[fId] = evtEnt
		}
	}

	for fId, evt := range evtMap {
		if evt.IsFault == true {
			cFId := EventKey{
				DaemonId: evt.ClearingDaemonId,
				EventId:  evt.ClearingEventId,
			}
			cEvt, exist := evtMap[cFId]
			if !exist {
				fMgr.logger.Err(fmt.Sprintln("No clearing event found for fault:", fId))
				continue
			}

			cEvt.IsClearingEvent = true
			evtMap[cFId] = cEvt
		}
	}

	for fId, evt := range evtMap {
		if evt.IsFault == true {
			evtEnt, _ := fMgr.FaultEventMap[fId]
			evtEnt.RaiseFault = evt.RaiseFault
			evtEnt.ClearingEventId = evt.ClearingEventId
			evtEnt.ClearingDaemonId = evt.ClearingDaemonId
			evtEnt.FaultOwnerName = evt.OwnerName
			evtEnt.FaultEventName = evt.EventName
			evtEnt.FaultSrcObjName = evt.SrcObjName
			evtEnt.AlarmSeverity = evt.AlarmSeverity
			fMgr.FaultEventMap[fId] = evtEnt
			cFId := EventKey{
				DaemonId: evtEnt.ClearingDaemonId,
				EventId:  evtEnt.ClearingEventId,
			}
			cEvtEnt, _ := fMgr.NonFaultEventMap[cFId]
			cEvtEnt.FaultOwnerId = fId.DaemonId
			cEvtEnt.FaultEventId = fId.EventId
			fMgr.NonFaultEventMap[cFId] = cEvtEnt
		} else {
			evtEnt, _ := fMgr.NonFaultEventMap[fId]
			evtEnt.IsClearingEvent = evt.IsClearingEvent
			fMgr.NonFaultEventMap[fId] = evtEnt
		}
	}
	return nil
}
