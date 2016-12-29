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
	"bytes"
	"github.com/Cistern/sflow"
)

//Key to datagram DB
type sflowDgramIdx struct {
	ifIndex int32
	key     int32
}

//Datagram sent receipt from collector
type dgramSentRcpt struct {
	idx         sflowDgramIdx
	collectorId string
}

//Sflow datagram definitions
//Interface definition for dgrams stored in dgram DB
type sflowDgram interface {
	GetBytes() []byte
	GetNumSflowSamples() int32
}

//Flow sample datagram sent to core dispatcher routine
type sflowDgramInfo struct {
	idx   sflowDgramIdx
	dgram sflowDgram
}

//Single container object that implements the sflow datagram interface
type sflowDgramObj struct {
	numSamples int32
	buf        bytes.Buffer
}

func (f *sflowDgramObj) GetBytes() []byte {
	return f.buf.Bytes()
}

func (f *sflowDgramObj) GetNumSflowSamples() int32 {
	return f.numSamples
}

//Sflow flow sample datagram generation
func (srvr *sflowServer) constructFlowSampleDgram(encoder *sflow.Encoder, flowSampleRcrd *flowSampleInfo) sflowDgram {
	var lenToSend int
	/*
	 - Any required packet parsing logic can be added here if/when needed
	 - Currently generating a flow record from byte stream as is
	 - Protocol is alway ethernet
	*/
	maxLen := int(srvr.sflowGblDB.maxSampledSize)
	if len(flowSampleRcrd.sflowData) < maxLen {
		lenToSend = len(flowSampleRcrd.sflowData)
	} else {
		lenToSend = maxLen + 1
	}
	rcrd := sflow.RawPacketFlow{
		Protocol:    ETHERNET_ISO88023,
		FrameLength: uint32(len(flowSampleRcrd.sflowData)),
		Stripped:    0,
		HeaderSize:  uint32(len(flowSampleRcrd.sflowData[:lenToSend])),
		Header:      flowSampleRcrd.sflowData[:lenToSend],
	}
	flowSample := &sflow.FlowSample{
		SequenceNum:      srvr.sflowSeqNum.flowSampleSeqNum,
		SourceIdType:     0,
		SourceIdIndexVal: uint32(flowSampleRcrd.ifIndex),
		SamplingRate:     uint32(srvr.sflowIntfDB[flowSampleRcrd.ifIndex].samplingRate),
		Input:            uint32(flowSampleRcrd.ifIndex),
		Records:          []sflow.Record{rcrd},
	}
	//Update flowsample seq num
	srvr.sflowSeqNum.flowSampleSeqNum += 1

	//Encode flow sample to generate byte stream
	var buf bytes.Buffer
	var dgram *sflowDgramObj
	err := encoder.Encode(&buf, []sflow.Sample{flowSample})
	if err == nil {
		dgram = &sflowDgramObj{
			buf:        buf,
			numSamples: int32(len(flowSample.Records)),
		}
	}
	return dgram
}

//Sflow counter sample datagram generation
func (srvr *sflowServer) constructCounterSampleDgram(encoder *sflow.Encoder, counterRcrd *counterInfo) sflowDgram {
	var rcrd sflow.GenericInterfaceCounters
	switch counterRcrd.ctrRcrdType {
	case GENERIC_IF_CTRS:
		rcrd = sflow.GenericInterfaceCounters{
			Index:               uint32(counterRcrd.ifIndex),
			Type:                0,
			Speed:               uint64(counterRcrd.ctrVals[GEN_IF_CTR_SPEED]),
			Direction:           uint32(counterRcrd.ctrVals[GEN_IF_CTR_DUPLEX]),
			Status:              uint32(counterRcrd.ctrVals[GEN_IF_CTR_OPERSTATUS]),
			InOctets:            uint64(counterRcrd.ctrVals[GEN_IF_CTR_IN_OCTETS]),
			InUnicastPackets:    uint32(counterRcrd.ctrVals[GEN_IF_CTR_IN_UCAST_PKTS]),
			InMulticastPackets:  uint32(counterRcrd.ctrVals[GEN_IF_CTR_IN_MCAST_PKTS]),
			InBroadcastPackets:  uint32(counterRcrd.ctrVals[GEN_IF_CTR_IN_BCAST_PKTS]),
			InDiscards:          uint32(counterRcrd.ctrVals[GEN_IF_CTR_IN_DISCARDS]),
			InErrors:            uint32(counterRcrd.ctrVals[GEN_IF_CTR_IN_ERRORS]),
			InUnknownProtocols:  uint32(counterRcrd.ctrVals[GEN_IF_CTR_IN_UNKNWN_PROTO]),
			OutOctets:           uint64(counterRcrd.ctrVals[GEN_IF_CTR_OUT_OCTETS]),
			OutUnicastPackets:   uint32(counterRcrd.ctrVals[GEN_IF_CTR_OUT_UCAST_PKTS]),
			OutMulticastPackets: uint32(counterRcrd.ctrVals[GEN_IF_CTR_OUT_MCAST_PKTS]),
			OutBroadcastPackets: uint32(counterRcrd.ctrVals[GEN_IF_CTR_OUT_BCAST_PKTS]),
			OutDiscards:         uint32(counterRcrd.ctrVals[GEN_IF_CTR_OUT_DISCARDS]),
			OutErrors:           uint32(counterRcrd.ctrVals[GEN_IF_CTR_OUT_ERRORS]),
			PromiscuousMode:     uint32(counterRcrd.ctrVals[GEN_IF_CTR_PROM_MODE]),
		}

	//Support for other types of counter datagrams can be added here

	default:
		logger.Err("Unsupported counter record type received, during construct dgram")
	}
	ctrSample := &sflow.CounterSample{
		SequenceNum:      srvr.sflowSeqNum.counterSampleSeqNum,
		SourceIdType:     0,
		SourceIdIndexVal: uint32(counterRcrd.ifIndex),
		Records:          []sflow.Record{rcrd},
	}
	//Update counter sample seq num
	srvr.sflowSeqNum.counterSampleSeqNum += 1

	//Encode counter sample to generate byte stream
	var buf bytes.Buffer
	var dgram *sflowDgramObj
	err := encoder.Encode(&buf, []sflow.Sample{ctrSample})
	if err == nil {
		dgram = &sflowDgramObj{
			buf:        buf,
			numSamples: int32(len(ctrSample.Records)),
		}
	}
	return dgram
}
