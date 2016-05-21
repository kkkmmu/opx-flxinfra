//
//Copyright [2016] [SnapRoute Inc]
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//	 Unless required by applicable law or agreed to in writing, software
//	 distributed under the License is distributed on an "AS IS" BASIS,
//	 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	 See the License for the specific language governing permissions and
//	 limitations under the License.
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
	"fmt"
	"utils/commonDefs"
)

func (svr *SYSDServer) processAsicNotification(msg commonDefs.AsicdNotifyMsg) {
	switch msg.(type) {
	case commonDefs.IPv4IntfNotifyMsg:
		ipv4Msg := msg.(commonDefs.IPv4IntfNotifyMsg)
		svr.handleIPv4IntfNotification(ipv4Msg)
	}
}

func (svr *SYSDServer) handleIPv4IntfNotification(msg commonDefs.IPv4IntfNotifyMsg) {
	switch msg.MsgType {
	case commonDefs.NOTIFY_IPV4INTF_CREATE:

	case commonDefs.NOTIFY_IPV4INTF_DELETE:
	}
}

func (svr *SYSDServer) buildIPv4Info() {
	ipv4Info, err := svr.AsicPlugin.GetAllIPv4IntfState()
	if err != nil {
		svr.logger.Info(fmt.Sprintln("Failed to get information from asic plugin err:", err))
		return
	}
	svr.logger.Info(fmt.Sprintln("ipv4 information ", ipv4Info))
}
