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
	"fmt"
	"infra/statsd/objects"
	"net"
)

func (c *sflowCollector) collectorTx(receiptChan chan *dgramSentRcpt, termCh chan string) {
	collectorAddrStr := fmt.Sprintf("%s:%d", c.ipAddr, c.udpPort)
	collectorAddr, err := net.ResolveUDPAddr("udp", collectorAddrStr)
	if err != nil {
		logger.Err("Error Resolving the UDP Address", err)
		c.operstate = objects.ADMIN_STATE_DOWN
		c.initCompleteCh <- false
		return
	}
	conn, err := net.DialUDP("udp", nil, collectorAddr)
	if err != nil {
		logger.Err("Error opening UDP connection for", collectorAddr)
		c.operstate = objects.ADMIN_STATE_DOWN
		c.initCompleteCh <- false
		return
	}
	logger.Debug("collectorTx: UDP init complete starting serve loop - ", collectorAddr)
	c.operstate = objects.ADMIN_STATE_UP
	c.initCompleteCh <- true
	for {
		select {
		case sflowDgramInfo := <-c.dgramRcvCh:
			_, err := conn.Write(sflowDgramInfo.dgram.GetBytes())
			if err != nil {
				logger.Err("Error sending data to collector:", collectorAddrStr, err)
			} else {
				c.numDatagramExported++
				c.numSflowSamplesExported = c.numSflowSamplesExported +
					sflowDgramInfo.dgram.GetNumSflowSamples()
			}
			receiptChan <- &dgramSentRcpt{
				idx: sflowDgramIdx{
					ifIndex: sflowDgramInfo.idx.ifIndex,
					key:     sflowDgramInfo.idx.key,
				},
				collectorId: c.ipAddr,
			}
		case <-c.shutdownCh:
			logger.Debug("collectorTx: Received shutdown for collector : ", c.ipAddr)
			c.operstate = objects.ADMIN_STATE_DOWN
			conn.Close()
			//Post bye msg on term channel
			termCh <- c.ipAddr
			return
		}
	}
}

func (c *sflowCollector) shutdownCollectorTx() {
	c.shutdownCh <- true
}
