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
)

type sflowDgramIdx struct {
	ifIndex int32
	key     int32
}

type sflowDgramInfo struct {
	idx   sflowDgramIdx
	dgram sflowDgram
}

type dgramSentRcpt struct {
	idx         sflowDgramIdx
	collectorId string
}

type sflowRecordSeqNum struct {
	flowSampleSeqNum uint32
	//Seq nums for other types to be added here
}

//Interface definition for dgrams sent to collector
type sflowDgram interface {
	GetBytes() []byte
	GetNumSflowSamples() int32
}

//Flow sample datagram
type flowSampleDgram struct {
	numSamples int32
	buf        bytes.Buffer
}

func (f *flowSampleDgram) GetBytes() []byte {
	return f.buf.Bytes()
}

func (f *flowSampleDgram) GetNumSflowSamples() int32 {
	return f.numSamples
}
