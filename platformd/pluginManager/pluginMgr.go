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

package pluginManager

import (
	"infra/platformd/pluginManager/onlp"
	"infra/platformd/pluginManager/openBMC"
	"infra/platformd/pluginManager/pluginCommon"
	"strings"
)

type PluginIntf interface {
	Init() error
	DeInit() error
}

func NewPlugin(pluginName string, initParams *pluginCommon.PluginInitParams) PluginIntf {
	var plugin PluginIntf
	pluginName = strings.ToLower(pluginName)
	switch pluginName {
	case pluginCommon.ONLP_PLUGIN:
		plugin = onlp.NewONLPPlugin(initParams)
	case pluginCommon.OpenBMC_PLUGIN:
		plugin = openBMC.NewOpenBMCPlugin(initParams)
	default:
	}
	return plugin
}
