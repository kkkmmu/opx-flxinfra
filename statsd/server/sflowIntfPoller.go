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

import (
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"infra/statsd/objects"
	"time"
)

const (
	//Flow sample poller related defs
	SNAPSHOT_LEN int32         = 65549
	PROMISCUOUS  bool          = false
	PCAP_TIMEOUT time.Duration = time.Duration(1) * time.Second

	//Counter sample poller related defs
	SFLOW_INTF_OPERSTATE_UP          uint64 = 1
	SFLOW_INTF_OPERSTATE_DOWN               = 0
	SFLOW_INTF_DUPLEX_MODE_FULL             = 1
	SFLOW_INTF_DUPLEX_MODE_HALF             = 2
	SFLOW_INTF_PROM_MODE_UNSUPPORTED        = 0
)

type intfPollerInitParams struct {
	intfRef              string
	netDevName           string
	sflowIntfRecordCh    chan *flowSampleInfo
	sflowIntfCtrRecordCh chan *counterInfo
	pollInterval         int32
}

func (intf *sflowIntf) intfPoller(initParams *intfPollerInitParams) {
	//Poll interface and send data to Dgram creation engine
	pcapHdl, err := pcap.OpenLive(initParams.netDevName, SNAPSHOT_LEN, PROMISCUOUS, PCAP_TIMEOUT)
	if err != nil || pcapHdl == nil {
		intf.operstate = objects.ADMIN_STATE_DOWN
		intf.initCompleteCh <- false
		logger.Err("IntfPoller: Unable to open Pcap Handle.", err, pcapHdl)
		return
	}
	intf.operstate = objects.ADMIN_STATE_UP
	intf.initCompleteCh <- true
	logger.Debug("IntfPoller: Init complete for interface, starting poller : ", initParams.intfRef, intf.operstate)
	intf.startPolling(pcapHdl, initParams.intfRef, initParams.sflowIntfRecordCh, initParams.sflowIntfCtrRecordCh, initParams.pollInterval)
}

func (intf *sflowIntf) startPolling(pcapHdl *pcap.Handle, intfRef string, sflowIntfRecordCh chan *flowSampleInfo, sflowIntfCtrRecordCh chan *counterInfo, pollInterval int32) {
	//Create channel to chain shutdown msg for counter poller
	var ctrShutdownCh chan bool = make(chan bool)

	//Spawn interface counter poller
	go intf.startCounterPoller(intfRef, sflowIntfCtrRecordCh, ctrShutdownCh, pollInterval)

	src := gopacket.NewPacketSource(pcapHdl, pcapHdl.LinkType())
	in := src.Packets()
	for {
		select {
		case packet, ok := <-in:
			if ok {
				intf.numSflowSamplesExported++
				sflowIntfRecordCh <- &flowSampleInfo{
					ifIndex:   intf.ifIndex,
					sflowData: packet.Data(),
				}
			}

		case <-intf.shutdownCh:
			logger.Debug("IntfPoller: Received shutdown for interface : ", intf.ifIndex)
			//First shutdown counter poller
			logger.Debug("IntfPoller: Shutting down counter poller")
			ctrShutdownCh <- true
			intf.operstate = objects.ADMIN_STATE_DOWN
			pcapHdl.Close()
			return
		}
	}
}

//Utility routine to construct a counterinfo object
func constructGenericIfCtrInfo(ifIndex int32, cfgObj *objects.SflowIntfCfgInfo, ctrObj *objects.SflowIntfCounterInfo) *counterInfo {
	ctrRcrd := new(counterInfo)
	ctrRcrd.ifIndex = ifIndex
	ctrRcrd.ctrRcrdType = GENERIC_IF_CTRS
	ctrRcrd.ctrVals = make(map[uint32]uint64)

	//Insert counter entries into value map
	ctrRcrd.ctrVals[GEN_IF_CTR_SPEED] = uint64(cfgObj.Speed)

	//FIXME: Following encoding for duplex and operstate is based on sflow v5 spec, this code
	//should be moved to the DGRAM construction routine
	if cfgObj.FullDuplex {
		ctrRcrd.ctrVals[GEN_IF_CTR_DUPLEX] = uint64(SFLOW_INTF_DUPLEX_MODE_FULL)
	} else {
		ctrRcrd.ctrVals[GEN_IF_CTR_DUPLEX] = uint64(SFLOW_INTF_DUPLEX_MODE_HALF)
	}
	if ctrObj.OperState {
		ctrRcrd.ctrVals[GEN_IF_CTR_OPERSTATUS] = uint64(SFLOW_INTF_OPERSTATE_UP)
	} else {
		ctrRcrd.ctrVals[GEN_IF_CTR_OPERSTATUS] = uint64(SFLOW_INTF_OPERSTATE_DOWN)
	}
	ctrRcrd.ctrVals[GEN_IF_CTR_IN_OCTETS] = ctrObj.IfInOctets
	ctrRcrd.ctrVals[GEN_IF_CTR_IN_UCAST_PKTS] = ctrObj.IfInUcastPkts
	ctrRcrd.ctrVals[GEN_IF_CTR_IN_MCAST_PKTS] = ctrObj.IfInMcastPkts
	ctrRcrd.ctrVals[GEN_IF_CTR_IN_BCAST_PKTS] = ctrObj.IfInBcastPkts
	ctrRcrd.ctrVals[GEN_IF_CTR_IN_DISCARDS] = ctrObj.IfInDiscards
	ctrRcrd.ctrVals[GEN_IF_CTR_IN_ERRORS] = ctrObj.IfInErrors
	ctrRcrd.ctrVals[GEN_IF_CTR_IN_UNKNWN_PROTO] = ctrObj.IfInUnknownProtos
	ctrRcrd.ctrVals[GEN_IF_CTR_OUT_OCTETS] = ctrObj.IfOutOctets
	ctrRcrd.ctrVals[GEN_IF_CTR_OUT_UCAST_PKTS] = ctrObj.IfOutUcastPkts
	ctrRcrd.ctrVals[GEN_IF_CTR_OUT_MCAST_PKTS] = ctrObj.IfOutMcastPkts
	ctrRcrd.ctrVals[GEN_IF_CTR_OUT_BCAST_PKTS] = ctrObj.IfOutBcastPkts
	ctrRcrd.ctrVals[GEN_IF_CTR_OUT_DISCARDS] = ctrObj.IfOutDiscards
	ctrRcrd.ctrVals[GEN_IF_CTR_OUT_ERRORS] = ctrObj.IfOutErrors
	ctrRcrd.ctrVals[GEN_IF_CTR_PROM_MODE] = SFLOW_INTF_PROM_MODE_UNSUPPORTED
	return ctrRcrd
}

func (intf *sflowIntf) startCounterPoller(intfRef string, sflowIntfCtrRecordCh chan *counterInfo, shutdownCh chan bool, pollInterval int32) {
	var ticker *time.Ticker
	if pollInterval != 0 {
		ticker = time.NewTicker(time.Duration(pollInterval) * time.Second)
	}
	for {
		select {
		case intvl := <-intf.restartCtrPollTicker:
			//Counter poll interval has been modified, restart ticker
			if ticker != nil {
				ticker.Stop()
				ticker = nil
			}
			if intvl != 0 {
				ticker = time.NewTicker(time.Duration(intvl) * time.Second)
			}

		case _ = <-ticker.C:
			//On tick poll interface info and post to core rx
			ctrObj := hwHdl.SflowGetIntfCounters(intfRef)
			if ctrObj != nil {
				cfgObj := hwHdl.SflowGetIntfCfgInfo(intfRef)
				if cfgObj != nil {
					//Populate counter record info and send to core
					sflowIntfCtrRecordCh <- constructGenericIfCtrInfo(intf.ifIndex, cfgObj, ctrObj)
				}
			}

		case _ = <-shutdownCh:
			logger.Debug("IntfCounterPoller: Received shutdown for interface : ", intf.ifIndex)
			return
		}
	}
}

func (intf *sflowIntf) shutdownIntfPoller() {
	intf.shutdownCh <- true
}
