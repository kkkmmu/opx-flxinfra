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
	"bytes"
	"github.com/Cistern/sflow"
	"infra/statsd/objects"
	"net"
)

const (
	SFLOW_SUB_AGENT_ID        = 0
	ETHERNET_ISO88023  uint32 = 1 //Sflow Dgram Protocol ID from Sflowv5 spec
)

func (srvr *sflowServer) constructFlowSampleDgram(encoder *sflow.Encoder,
	rcrdSeqNum *sflowRecordSeqNum, sflowRcrd sflowRecordInfo) sflowDgram {
	var lenToSend int

	/*
	 - Any required packet parsing logic can be added here if/when needed
	 - Currently generating a flow record from byte stream as is
	 - Protocol is alway ethernet
	*/
	maxLen := int(srvr.sflowGblDB.maxSampledSize)
	if len(sflowRcrd.sflowData) < maxLen {
		lenToSend = len(sflowRcrd.sflowData)
	} else {
		lenToSend = maxLen + 1
	}
	rcrd := sflow.RawPacketFlow{
		Protocol:    ETHERNET_ISO88023,
		FrameLength: uint32(len(sflowRcrd.sflowData)),
		Stripped:    0,
		HeaderSize:  uint32(srvr.sflowGblDB.maxSampledSize),
		Header:      sflowRcrd.sflowData[:lenToSend],
	}
	flowSample := &sflow.FlowSample{
		SequenceNum:      rcrdSeqNum.flowSampleSeqNum,
		SourceIdType:     0,
		SourceIdIndexVal: uint32(sflowRcrd.ifIndex),
		SamplingRate:     uint32(srvr.sflowIntfDB[sflowRcrd.ifIndex].samplingRate),
		Input:            uint32(sflowRcrd.ifIndex),
		Records:          []sflow.Record{rcrd},
	}
	//Update flowsample seq num
	rcrdSeqNum.flowSampleSeqNum += 1

	//Encode flow sample to generate byte stream
	var buf bytes.Buffer
	var dgram *flowSampleDgram
	err := encoder.Encode(&buf, []sflow.Sample{flowSample})
	if err == nil {
		dgram = &flowSampleDgram{
			buf:        buf,
			numSamples: int32(len(flowSample.Records)),
		}
	}
	return dgram
}

/*
 - Drain channels populated by interface pollers
 - Construct sflow datagrams
 - Post to sflow transmit channel
*/
func (srvr *sflowServer) sflowCoreRx() {
	//Internal DgramDB seqnum
	var intSeqNum int32
	//Sequence numbers for various record types
	var rcrdSeqNum sflowRecordSeqNum
	//sflow encoder
	var encoder *sflow.Encoder

	for {
		select {
		case sflowRcrd := <-srvr.sflowIntfRecordCh:
			/*
			 - Construct SFLOW v5 datagram and save in DgramDB
			 - FIXME: Currently, One sample per datagram
			*/
			dgram := srvr.constructFlowSampleDgram(encoder, &rcrdSeqNum, sflowRcrd)
			if dgram != nil {
				//Save record to DGRAM DB
				if srvr.sflowDgramDB[sflowRcrd.ifIndex] == nil {
					srvr.sflowDgramDB[sflowRcrd.ifIndex] = make(map[int32]sflowDgram)
				} else {
					srvr.sflowDgramDB[sflowRcrd.ifIndex][intSeqNum] = dgram
					//Dispatch dgram info to listener
					srvr.sflowDgramRdy <- &sflowDgramInfo{
						idx: sflowDgramIdx{
							ifIndex: sflowRcrd.ifIndex,
							key:     intSeqNum,
						},
						dgram: dgram,
					}
					//Bump up seq number
					intSeqNum += 1
				}
			}

		case cfg := <-srvr.gblCfgCh:
			switch cfg.op {
			case objects.SFLOW_GLOBAL_UPDATE_ATTR_AGENT_IPADDR:
				//Free earlier encoder instance
				encoder = nil
				agentIpAddr := net.ParseIP(srvr.sflowGblDB.agentIpAddr)
				encoder = sflow.NewEncoder(agentIpAddr, SFLOW_SUB_AGENT_ID, 0)
			}
		}
	}
}

/*
 - Drain sflow datagrams
 - Dispatch to all collectors
 - Cleanup datagram memory after all collectors have sent
*/
func (srvr *sflowServer) sflowCoreTx() {
	//Refcounter used to track when to free dgrams from DB
	var refCounter map[sflowDgramIdx]map[string]bool = make(map[sflowDgramIdx]map[string]bool)

	for {
		select {
		case dgram := <-srvr.sflowDgramRdy:
			logger.Debug("SflowCoreTx: Received sflow dgram ready. Sending to collectors")
			refCounter[dgram.idx] = make(map[string]bool)
			for collectorId, ch := range srvr.sflowDgramToCollector {
				ch <- dgram
				//Insert id into map to aid with refCnt
				refCounter[dgram.idx][collectorId] = true
				logger.Debug("SflowCoreTx: Sent sflow dgram to collector : ", collectorId)
			}

		case rcpt := <-srvr.sflowDgramSentReceiptCh:
			if m, ok := refCounter[rcpt.idx]; ok {
				if _, ok = m[rcpt.collectorId]; ok {
					//Valid ref found, delete map entry
					delete(m, rcpt.collectorId)
					//Check if all references to dgram idx are freed. If yes, free dgram from DB
					if len(refCounter[rcpt.idx]) == 0 {
						//All references are cleaned up, cleanup dgram from DB
						delete(srvr.sflowDgramDB[rcpt.idx.ifIndex], rcpt.idx.key)
						logger.Debug("SflowCoreTx: Cleaned up reference to dgram in DB, key: ", rcpt.idx.ifIndex, rcpt.idx.key)
					}
				} else {
					logger.Warning("SflowCoreTx: Received dgram sent receipt, with invalid collector ID")
				}
			} else {
				logger.Warning("SflowCoreTx: Received dgram sent receipt, with invalid idx", rcpt)
			}

		case collectorId := <-srvr.collectorTerminatedCh:
			var found bool
			logger.Debug("SflowCoreTx: Bye message received from collector : ", collectorId)
			//Received a bye msg from a collector. Cleanup all references held for this collector
			for idx, m := range refCounter {
				for cid, _ := range m {
					if cid == collectorId {
						found = true
						break
					}
				}
				if found {
					//Delete reference to collector
					delete(m, collectorId)
					//If all references are cleaned up, cleanup dgram DB also
					if len(m) == 0 {
						logger.Debug("SflowCoreTx: Cleaned up reference to dgram in DB, key: ", idx.ifIndex, idx.key)
						delete(srvr.sflowDgramDB[idx.ifIndex], idx.key)
					}
					found = false
				}
			}
		}
	}
}
