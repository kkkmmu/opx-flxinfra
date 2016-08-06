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

package server

import (
	//	"encoding/json"
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	//	"io/ioutil"
	//	"models/events"
	"time"
	//	"utils/eventUtils"
	"utils/logging"
	//	"utils/ringBuffer"
	"infra/fMgrd/faultMgr"
)

//type FaultDataMap map[FaultObjKey]FaultData

/* ---Ashu
type FaultDatabaseKey struct {
	FaultId        FaultId
	FObjKey        FaultObjKey
	Resolved       bool
	FaultSeqNumber uint64
	OwnerName      string
	EventName      string
	OccuranceTime  time.Time
	ResolutionTime time.Time
	Description    string
}

type AlarmDatabaseKey struct {
	FaultId  FaultId
	FObjKey  FaultObjKey
	FaultIdx int
}
*/

type FMGRServer struct {
	Logger logging.LoggerIntf
	dbHdl  redis.Conn
	subHdl redis.PubSubConn
	fMgr   *faultMgr.FaultManager
	//	FaultEventMap              map[FaultId]FaultDetail
	//	NonFaultEventMap           map[FaultId]NonFaultData
	//FaultDatabase  map[FaultId]FaultDataMap
	//FaultList      *ringBuffer.RingBuffer
	//AlarmList      *ringBuffer.RingBuffer
	//FaultSeqNumber uint64
	InitDone chan bool
	//daemonList     []string
	//	eventProcessCh             chan []byte
	//FaultToAlarmTransitionTime time.Duration
	ReqChan   chan *ServerRequest
	ReplyChan chan interface{}
}

func NewFMGRServer(logger *logging.Writer) *FMGRServer {
	fMgrServer := &FMGRServer{}
	fMgrServer.Logger = logger
	fMgrServer.InitDone = make(chan bool)
	fMgrServer.ReqChan = make(chan *ServerRequest)
	fMgrServer.ReplyChan = make(chan interface{})
	//fMgrServer.FaultDatabase = make(map[FaultId]FaultDataMap)
	//fMgrServer.FaultList = new(ringBuffer.RingBuffer)
	//fMgrServer.FaultList.SetRingBufferCapacity(100000) //Max 100000 entries in fault database
	//fMgrServer.AlarmList = new(ringBuffer.RingBuffer)
	//fMgrServer.AlarmList.SetRingBufferCapacity(100000) //Max 100000 entries in fault database
	//fMgrServer.FaultSeqNumber = 0
	//fMgrServer.FaultToAlarmTransitionTime = time.Duration(3) * time.Second
	return fMgrServer
}

func (server *FMGRServer) dial() (redis.Conn, error) {
	retryCount := 0
	ticker := time.NewTicker(2 * time.Second)
	for _ = range ticker.C {
		retryCount += 1
		dbHdl, err := redis.Dial("tcp", ":6379")
		if err != nil {
			if retryCount%100 == 0 {
				server.Logger.Err(fmt.Sprintln("Failed to dail out to Redis server. Retrying connection. Num of retries = ", retryCount))
			}
		} else {
			return dbHdl, nil
		}
	}
	err := errors.New("Error opening db handler")
	return nil, err
}

func (server *FMGRServer) Subscriber() {
	for {
		switch n := server.subHdl.Receive().(type) {
		case redis.Message:
			server.fMgr.EventCh <- n.Data
		case redis.Subscription:
			if n.Count == 0 {
				return
			}
		case error:
			server.Logger.Err(fmt.Sprintf("error: %v\n", n))
			return
		}
	}
	server.subHdl.Unsubscribe()
	server.subHdl.PUnsubscribe()
}

func (server *FMGRServer) InitSubscriber() error {
	var errMsg string
	for _, daemon := range server.fMgr.DaemonList {
		err := server.subHdl.Subscribe(daemon)
		if err != nil {
			errMsg = fmt.Sprintf("%s : %s", errMsg, err)
		}
	}

	if errMsg == "" {
		return nil
	}

	return errors.New(fmt.Sprintln("Error Initializing Subscriber:", errMsg))
}

func (server *FMGRServer) InitServer() error {
	server.fMgr = faultMgr.NewFaultManager(server.Logger)
	err := server.fMgr.InitFaultManager()
	if err != nil {
		server.Logger.Err(fmt.Sprintln(err))
		return err
	}
	server.dbHdl, err = server.dial()
	if err != nil {
		server.Logger.Err(fmt.Sprintln(err))
		return err
	}
	//defer server.dbHdl.Close()
	server.subHdl = redis.PubSubConn{Conn: server.dbHdl}

	err = server.InitSubscriber()
	if err != nil {
		server.Logger.Err(fmt.Sprintln("Error in Initializing Subscriber", err))
		return err
	}
	go server.Subscriber()
	return err
}

func (server *FMGRServer) handleRPCRequest(req *ServerRequest) {
	server.Logger.Info(fmt.Sprintln("Calling handle RPC Request for:", *req))
	switch req.Op {
	case GET_BULK_FAULT_STATE:
		var retObj GetBulkFaultStateOutArgs
		if val, ok := req.Data.(*GetBulkInArgs); ok {
			retObj.BulkInfo, retObj.Err = server.getBulkFaultState(val.FromIdx, val.Count)
		}
		server.ReplyChan <- interface{}(&retObj)
	case GET_BULK_ALARM_STATE:
		var retObj GetBulkAlarmStateOutArgs
		if val, ok := req.Data.(*GetBulkInArgs); ok {
			retObj.BulkInfo, retObj.Err = server.getBulkAlarmState(val.FromIdx, val.Count)
		}
		server.ReplyChan <- interface{}(&retObj)
	default:
		server.Logger.Err(fmt.Sprintln("Error: Server received unrecognized request - ", req.Op))
	}
}

func (server *FMGRServer) StartServer() {
	server.InitServer()
	server.InitDone <- true
	for {
		select {
		case req := <-server.ReqChan:
			server.Logger.Info(fmt.Sprintln("Server request received - ", *req))
			server.handleRPCRequest(req)
		}
	}
}
