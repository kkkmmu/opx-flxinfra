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
	SNAPSHOT_LEN int32         = 65549
	PROMISCUOUS  bool          = false
	PCAP_TIMEOUT time.Duration = time.Duration(1) * time.Second
)

func (intf *sflowIntf) intfPoller(intfRef string, sflowIntfRecordCh chan sflowRecordInfo) {
	//Poll interface and send data to Dgram creation engine
	pcapHdl, err := pcap.OpenLive(intfRef, SNAPSHOT_LEN, PROMISCUOUS, PCAP_TIMEOUT)
	if err != nil || pcapHdl == nil {
		intf.operstate = objects.ADMIN_STATE_DOWN
		intf.initCompleteCh <- false
		logger.Err("intfPoller(): Unable to open Pcap Handle.", err, pcapHdl)
		return
	}
	intf.operstate = objects.ADMIN_STATE_UP
	intf.initCompleteCh <- true
	logger.Debug("IntfPoller: Init complete for interface, starting poller : ", intfRef, intf.operstate)
	intf.startPolling(pcapHdl, sflowIntfRecordCh)
}

func (intf *sflowIntf) startPolling(pcapHdl *pcap.Handle, sflowIntfRecordCh chan sflowRecordInfo) {
	src := gopacket.NewPacketSource(pcapHdl, pcapHdl.LinkType())
	in := src.Packets()
	for {
		select {
		case packet, ok := <-in:
			if ok {
				intf.numSflowSamplesExported++
				sflowIntfRecordCh <- sflowRecordInfo{
					ifIndex:   intf.ifIndex,
					sflowData: packet.Data(),
				}
			}
		case <-intf.shutdownCh:
			logger.Debug("IntfPoller: Received shutdown for interface : ", intf.ifIndex)
			intf.operstate = objects.ADMIN_STATE_DOWN
			pcapHdl.Close()
			return
		}
	}
}
