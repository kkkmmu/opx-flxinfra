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
// _______   __       __________   ___      _______.____    __    ____  __  .___________.  ______  __    __
// |   ____||  |     |   ____\  \ /  /     /       |\   \  /  \  /   / |  | |           | /      ||  |  |  |
// |  |__   |  |     |  |__   \  V  /     |   (----  \   \/    \/   /  |  |  ---|  |----    ,---- |  |__|  |
// |   __|  |  |     |   __|   >   <       \   \      \            /   |  |     |  |        |     |   __   |
// |  |     |  `----.|  |____ /  .  \  .----)   |      \    /\    /    |  |     |  |        `----.|  |  |  |
// |__|     |_______||_______/__/ \__\ |_______/        \__/  \__/     |__|     |__|      \______||__|  |__|
//

package server

const (
	//Defining various channel sizes
	INTF_POLLER_TO_SVR_CHAN_SIZE     = 100
	DGRAM_RDY_NOTIFICATION_CHAN_SIZE = 100
	DGRAM_SENT_RECEIPT_CHAN_SIZE     = 100
	COLLECTOR_TERMINATED_CHAN_SIZE   = 100
)

func (srvr *sflowServer) initSflowServer() {
	srvr.sflowIntfDB = make(map[int32]*sflowIntf)
	srvr.sflowCollectorDB = make(map[string]*sflowCollector)
	srvr.netDevInfo = make(map[int32]netDevData)
	srvr.sflowDgramDB = make(map[int32]map[int32]sflowDgram)
	srvr.sflowIntfRecordCh = make(chan *flowSampleInfo, INTF_POLLER_TO_SVR_CHAN_SIZE)
	srvr.sflowIntfCtrRecordCh = make(chan *counterInfo, INTF_POLLER_TO_SVR_CHAN_SIZE)
	srvr.sflowDgramRdy = make(chan *sflowDgramInfo, DGRAM_RDY_NOTIFICATION_CHAN_SIZE)
	srvr.regUnregCollectorCh = make(chan *collectorChanInfo)
	srvr.collectorChRegUnregAck = make(chan bool)
	srvr.sflowDgramSentReceiptCh = make(chan *dgramSentRcpt, DGRAM_SENT_RECEIPT_CHAN_SIZE)
	srvr.collectorTerminatedCh = make(chan string, COLLECTOR_TERMINATED_CHAN_SIZE)
	srvr.gblCfgCh = make(chan *gblCfgUpdateInfo)
}

func (srvr *sflowServer) constructSflowInfra() error {
	logger.Debug("Constructing Sflow infrastructure")
	netdevInfo, err := hwHdl.GetSflowNetdevInfo()
	if err == nil {
		for _, val := range netdevInfo {
			srvr.netDevInfo[val.IfIndex] = netDevData{
				intfRef:    val.IntfRef,
				netDevName: val.NetDevName,
			}
		}
	}
	logger.Debug("===============SFLOW NETDEVINFO START============")
	logger.Debug("Sflow infra dump : ", srvr.netDevInfo)
	logger.Debug("===============SFLOW NETDEVINFO END==============")
	return err
}
