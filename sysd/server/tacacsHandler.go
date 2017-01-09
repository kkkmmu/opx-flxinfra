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
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"models/objects"
	"os"
	"os/exec"
	"strings"
)

var COMMON_AUTH_CONFIG string = "auth [default=1 success=ignore] pam_succeed_if.so uid >= 1000 quiet\n" +
	"auth [authinfo_unavail=ignore success=2 default=ignore] " +
	"/usr/local/lib/security/pam_tacplus.so include=/etc/tacplus_servers " +
	"login=pap protocol=ssh service=shell\n"

var COMMON_ACCOUNT_CONFIG string = "account [default=1 success=ignore] pam_succeed_if.so uid > 1000 quiet\n" +
	"account	[authinfo_unavail=ignore success=2 default=ignore]  " +
	"/usr/local/lib/security/pam_tacplus.so include=/etc/tacplus_servers " +
	"login=pap protocol=ssh service=shell\n"

var COMMON_SESSION_CONFIG string = "session [default=1 success=ignore] pam_succeed_if.so uid > 1000 quiet\n" +
	"session [authinfo_unavail=ignore success=done default=ignore] " +
	"/usr/local/lib/security/pam_tacplus.so include=/etc/tacplus_servers " +
	"login=pap protocol=ssh service=shell\n"

var COMMON_SESSION_NONINT_CONFIG string = "session [default=1 success=ignore] pam_succeed_if.so uid > 1000 quiet\n" +
	"session [authinfo_unavail=ignore success=done default=ignore] " +
	"/usr/local/lib/security/pam_tacplus.so include=/etc/tacplus_servers " +
	"login=pap protocol=ssh service=shell\n"

var TACPLUS_SERVERS_CONFIG_PROLOGUE string = "# This is a common file used by audisp-tacplus, libpam_tacplus, and\n" +
	"# libtacplus_map config files as shipped.\n" +
	"#\n" +
	"# Any tac_plus client config can go here that is common to all users of this\n" +
	"# file, but typically it's just the TACACS+ server IP address(es) and shared\n" +
	"# secret(s)\n" +
	"\n"

var TACPLUS_SERVERS_CONFIG string = "secret=%s\nserver=%s\n"

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type TacacsConfig struct {
	ServerIp       string
	SourceIntf     string
	AuthService    string
	Secret         string
	Port           int16
	PrivilegeLevel int32
	Debug          int32
}

type TacacsState struct {
	ServerIp       string
	SourceIntf     string
	AuthService    string
	Secret         string
	Port           int16
	PrivilegeLevel int32
	Debug          int32
	ConnFailReason string
}

type TacacsGlobalConfig struct {
	ProfileName string
	Enable      string
	Timeout     int32
}

type TacacsGlobalState struct {
	ProfileName       string
	OperStatus        string
	NumActiveSessions int32
}

type TacacsManager struct {
	TacacsStateSlice     []*TacacsState
	TacacsStateMap       map[string]*TacacsState
	TacacsGlobalStateObj *TacacsGlobalState
}

func NewTacacsManager() *TacacsManager {
	mgr := &TacacsManager{}
	mgr.TacacsStateSlice = []*TacacsState{}
	mgr.TacacsStateMap = make(map[string]*TacacsState)
	mgr.TacacsGlobalStateObj = &TacacsGlobalState{}
	return mgr
}

func convertTacacsObjToLocalType(
	modelObj *objects.TacacsConfig) *TacacsConfig {

	return &TacacsConfig{
		ServerIp:       modelObj.ServerIp,
		SourceIntf:     modelObj.SourceIntf,
		AuthService:    modelObj.AuthService,
		Secret:         modelObj.Secret,
		Port:           modelObj.Port,
		PrivilegeLevel: modelObj.PrivilegeLevel,
		Debug:          modelObj.Debug,
	}
}

func convertTacacsGlobalObjToLocalType(
	modelObj *objects.TacacsGlobalConfig) *TacacsGlobalConfig {

	return &TacacsGlobalConfig{
		ProfileName: modelObj.ProfileName,
		Enable:      modelObj.Enable,
		Timeout:     modelObj.Timeout,
	}
}

func (server *SYSDServer) ReadTacacsConfigFromDB() error {
	server.logger.Info("Reading Tacacs Global Config From Db")
	if server.dbHdl != nil {
		var dbObj objects.TacacsGlobalConfig
		objList, err := server.dbHdl.GetAllObjFromDb(dbObj)
		if err != nil {
			server.logger.Err("DB query failed for TacacsGlobalConfig")
			return err
		}
		for idx := 0; idx < len(objList); idx++ {
			dbObject := objList[idx].(objects.TacacsGlobalConfig)
			localObj := convertTacacsGlobalObjToLocalType(&dbObject)
			server.CreateTacacsGlobalConfig(localObj)
		}
	}
	server.logger.Info("Reading Tacacs Global Config done")

	server.logger.Info("Reading Tacacs Config From Db")
	if server.dbHdl != nil {
		var dbObj objects.TacacsConfig
		objList, err := server.dbHdl.GetAllObjFromDb(dbObj)
		if err != nil {
			server.logger.Err("DB query failed for TacacsConfig")
			return err
		}
		for idx := 0; idx < len(objList); idx++ {
			dbObject := objList[idx].(objects.TacacsConfig)
			localObj := convertTacacsObjToLocalType(&dbObject)
			server.CreateTacacsConfig(localObj)
		}
	}
	server.logger.Info("Reading Tacacs Config done")
	return nil
}

func (server *SYSDServer) checkPackage(pkgName string) bool {
	c1 := exec.Command("dpkg", "-l")
	c2 := exec.Command("grep", pkgName)

	r, w := io.Pipe()
	c1.Stdout = w
	c2.Stdin = r

	var b2 bytes.Buffer
	c2.Stdout = &b2

	c1.Start()
	c2.Start()
	c1.Wait()
	w.Close()
	c2.Wait()
	return strings.Contains(b2.String(), pkgName)
}

func (server *SYSDServer) checkTacacsLib() bool {
	return server.checkPackage("libpam-tacplus") && server.checkPackage("libnss-tacplus") &&
		server.checkPackage("audisp-tacplus") && server.checkPackage("libsimple-tacacct")
}

// func addServer(serverConfig string, secret string, serverIp string) string {
// 	serverConfigLines := strings.Split(
// 		strings.TrimSpace(serverConfig), "\n")
// 	configLine1 := "secret=" + secret
// 	configLine2 := "server=" + serverIp
// 	//configIndex is the start of configuration
// 	var idx int = 0
// 	for idx = 0; idx < len(serverConfigLines) &&
// 		!strings.HasPrefix(serverConfigLines[idx], "secret="); idx++ {
// 	}
// 	foundPattern := false
// 	for i := idx; i < len(serverConfigLines); i++ {
// 		if strings.TrimSpace(serverConfigLines[i]) == configLine1 &&
// 			i+1 < len(serverConfigLines) &&
// 			strings.TrimSpace(serverConfigLines[i+1]) == configLine2 {

// 			foundPattern = true
// 		}
// 	}
// 	if !foundPattern {
// 		serverConfigLines = append(serverConfigLines, configLine1)
// 		serverConfigLines = append(serverConfigLines, configLine2)
// 	}
// 	return strings.Join(serverConfigLines, "\n") + "\n"
// }

// func removeServer(serverConfig string, secret string, serverIp string) string {
// 	serverConfigLines := strings.Split(
// 		strings.TrimSpace(serverConfig), "\n")
// 	configLine1 := "secret=" + secret
// 	configLine2 := "server=" + serverIp
// 	//configIndex is the start of configuration
// 	var idx int = 0
// 	for idx = 0; idx < len(serverConfigLines) &&
// 		!strings.HasPrefix(serverConfigLines[idx], "secret="); idx++ {
// 	}
// 	for i := idx; i < len(serverConfigLines); i++ {
// 		if strings.TrimSpace(serverConfigLines[i]) == configLine1 &&
// 			i+1 < len(serverConfigLines) &&
// 			strings.TrimSpace(serverConfigLines[i+1]) == configLine2 {

// 			serverConfigLines = append(
// 				serverConfigLines[:i], serverConfigLines[i+2:]...)
// 			break
// 		}
// 	}
// 	return strings.Join(serverConfigLines, "\n") + "\n"
// }

// // Not sure if this is really needed
// func checkServer(serverConfig string, secret string, serverIp string) bool {
// 	return true
// }

func (server *SYSDServer) stopAuditd() {
	cmd := "sudo service auditd stop"
	_, _ = exec.Command("sudo", "service", "auditd", "stop").Output()
	server.logger.Info(cmd)
}

func (server *SYSDServer) startAuditd() {
	cmd := "sudo service auditd start"
	_, _ = exec.Command("sudo", "service", "auditd", "start").Output()
	server.logger.Info(cmd)
}

func (server *SYSDServer) copyConfigFiles() {
	cmd := "cp /usr/local/etc/tacplus_servers /etc/tacplus_servers"
	_, _ = exec.Command("cp", "/usr/local/etc/tacplus_servers", "/etc/tacplus_servers").Output()
	server.logger.Info(cmd)

	cmd = "cp /usr/local/etc/tacplus_nss.conf /etc/tacplus_nss.conf"
	_, _ = exec.Command("cp", "/usr/local/etc/tacplus_nss.conf", "/etc/tacplus_nss.conf").Output()
	server.logger.Info(cmd)

	cmd = "cp /usr/local/etc/audisp/audisp-tac_plus.conf /etc/audisp/"
	_, _ = exec.Command("cp", "/usr/local/etc/audisp/audisp-tac_plus.conf", "/etc/audisp/").Output()
	server.logger.Info(cmd)

	cmd = "cp /usr/local/etc/audisp/plugins.d/audisp-tacplus.conf " +
		"/etc/audisp/plugins.d/"
	_, _ = exec.Command("cp", "/usr/local/etc/audisp/plugins.d/audisp-tacplus.conf",
		"/etc/audisp/plugins.d/").Output()
	server.logger.Info(cmd)

	cmd = "cp /usr/local/etc/audit/rules.d/audisp-tacplus.rules " +
		"/etc/audit/rules.d/"
	_, _ = exec.Command("cp", "/usr/local/etc/audit/rules.d/audisp-tacplus.rules",
		"/etc/audit/rules.d/").Output()
	server.logger.Info(cmd)

	cmd = "cp /usr/local/sbin/audisp-tacplus /sbin/"
	_, _ = exec.Command("cp", "/usr/local/sbin/audisp-tacplus", "/sbin/").Output()
	server.logger.Info(cmd)
}

func (server *SYSDServer) loadAugenRules() {
	cmd := "augenrules --load"
	_, _ = exec.Command("augenrules", "--load").Output()
	server.logger.Info(cmd)
}

func (server *SYSDServer) editCommonConfigFiles() bool {
	var numFiles int = 4
	filePaths := [...]string{
		"/etc/pam.d/common-auth",
		"/etc/pam.d/common-account",
		"/etc/pam.d/common-session",
		"/etc/pam.d/common-session-noninteractive",
	}
	commonConfigs := [...]string{
		COMMON_AUTH_CONFIG,
		COMMON_ACCOUNT_CONFIG,
		COMMON_SESSION_CONFIG,
		COMMON_SESSION_NONINT_CONFIG,
	}
	newConfigText := ""
	for i := 0; i < numFiles; i++ {
		buf, err := ioutil.ReadFile(filePaths[i])
		check(err)
		lines := strings.Split(string(buf), "\n")
		fileConfigured := false
		commonConfigLines := strings.Split(commonConfigs[i], "\n")
		for _, line := range lines {
			// Check first line to check if the config is already set
			if line == commonConfigLines[0] {
				fileConfigured = true
				break
			}
			if strings.Contains(line, "pam_unix") {
				newConfigText += commonConfigs[i]
			}
			newConfigText += line + "\n"
		}
		if !fileConfigured {
			f, err := os.Create(filePaths[i])
			check(err)
			_, err = f.WriteString(newConfigText)
			check(err)
			f.Close()
		}
	}
	return true
}

func (server *SYSDServer) editNssConfig() bool {
	buf, err := ioutil.ReadFile("/etc/nsswitch.conf")
	check(err)
	lines := strings.Split(string(buf), "\n")
	newLines := ""
	fileConfigured := false
	for _, line := range lines {
		if strings.Contains(line, "passwd:") &&
			!strings.Contains(line, "tacplus") {

			newLines += line + " tacplus\n"
		} else if strings.Contains(line, "passwd:") &&
			strings.Contains(line, "tacplus") {

			fileConfigured = true
			break
		} else {
			newLines += line + "\n"
		}
	}
	if !fileConfigured {
		f, err := os.Create("/etc/nsswitch.conf")
		check(err)
		_, err = f.WriteString(newLines)
		check(err)
	}
	return true
}

func (server *SYSDServer) addTacacsLinuxUsers() {
	cmd := "bash -c \"for ((a=1; a<=15; a++)) " +
		"do   /usr/sbin/useradd tacacs$a -m -s /bin/bash; done\""
	_, _ = exec.Command("bash", "-c", "for ((a=1; a<=15; a++)) "+
		"do   useradd tacacs$a -m -s /bin/bash; done").Output()
	server.logger.Info(cmd)

	cmd = "echo \"tacacs15    ALL=(ALL:ALL) ALL\" > " +
		"/etc/sudoers.d/tacplus"
	_, _ = exec.Command("echo", "tacacs15    ALL=(ALL:ALL) ALL", ">",
		"/etc/sudoers.d/tacplus").Output()
	server.logger.Info(cmd)
}

func (server *SYSDServer) setupTacacsConfig() {
	server.logger.Info("Setting up TACACS Config")
	server.stopAuditd()
	server.copyConfigFiles()
	server.startAuditd()
	server.loadAugenRules()
	server.editCommonConfigFiles()
	server.addTacacsLinuxUsers()
	server.editNssConfig()
}

func (server *SYSDServer) resetServers() {
	server.logger.Info("Resetting TACACS Servers")
	configStr := TACPLUS_SERVERS_CONFIG_PROLOGUE
	f, err := os.Create("/etc/tacplus_servers")
	check(err)
	_, err = f.WriteString(configStr)
	check(err)
	f.Close()
}

func (server *SYSDServer) reconfigureServers() {
	server.logger.Info("Reconfiguring all TACACS Servers")
	configStr := TACPLUS_SERVERS_CONFIG_PROLOGUE
	server.tacacsMgr.TacacsGlobalStateObj.NumActiveSessions = 0
	for _, tacacsState := range server.tacacsMgr.TacacsStateSlice {
		configStr += fmt.Sprintf(
			TACPLUS_SERVERS_CONFIG,
			tacacsState.Secret,
			tacacsState.ServerIp,
		)
	}
	f, err := os.Create("/etc/tacplus_servers")
	check(err)
	_, err = f.WriteString(configStr)
	check(err)
	f.Close()
	server.tacacsMgr.TacacsGlobalStateObj.NumActiveSessions =
		int32(len(server.tacacsMgr.TacacsStateSlice))
}

func (server *SYSDServer) CreateTacacsConfig(
	tacacsConfig *TacacsConfig) (bool, error) {

	tacacsStateObj := &TacacsState{}
	tacacsStateObj.ServerIp = tacacsConfig.ServerIp
	tacacsStateObj.SourceIntf = tacacsConfig.SourceIntf
	tacacsStateObj.AuthService = tacacsConfig.AuthService
	tacacsStateObj.Secret = tacacsConfig.Secret
	tacacsStateObj.Port = tacacsConfig.Port
	tacacsStateObj.PrivilegeLevel = tacacsConfig.PrivilegeLevel
	tacacsStateObj.Debug = tacacsConfig.Debug
	tacacsStateObj.ConnFailReason = "None"
	server.tacacsMgr.TacacsStateMap[tacacsConfig.ServerIp] =
		tacacsStateObj
	server.tacacsMgr.TacacsStateSlice =
		append(server.tacacsMgr.TacacsStateSlice, tacacsStateObj)

	if server.tacacsMgr.TacacsGlobalStateObj != nil &&
		server.tacacsMgr.TacacsGlobalStateObj.OperStatus == "UP" {

		server.reconfigureServers()
	}
	return true, nil
}

func (server *SYSDServer) UpdateTacacsConfig(
	tacacsConfig *TacacsConfig) (bool, error) {

	tacacsStateObj, ok := server.tacacsMgr.TacacsStateMap[tacacsConfig.ServerIp]
	if !ok {
		return false, nil
	}

	tacacsStateObj.ServerIp = tacacsConfig.ServerIp
	tacacsStateObj.SourceIntf = tacacsConfig.SourceIntf
	tacacsStateObj.AuthService = tacacsConfig.AuthService
	tacacsStateObj.Secret = tacacsConfig.Secret
	tacacsStateObj.Port = tacacsConfig.Port
	tacacsStateObj.PrivilegeLevel = tacacsConfig.PrivilegeLevel
	tacacsStateObj.Debug = tacacsConfig.Debug
	tacacsStateObj.ConnFailReason = "None"
	server.tacacsMgr.TacacsStateMap[tacacsConfig.ServerIp] =
		tacacsStateObj

	if server.tacacsMgr.TacacsGlobalStateObj != nil &&
		server.tacacsMgr.TacacsGlobalStateObj.OperStatus == "UP" {

		server.reconfigureServers()
	}
	return true, nil
}

func (server *SYSDServer) DeleteTacacsConfig(serverIp string) (bool, error) {
	_, ok := server.tacacsMgr.TacacsStateMap[serverIp]
	if !ok {
		return false, nil
	}
	sliceEntIdx := -1
	for i, tacacsState := range server.tacacsMgr.TacacsStateSlice {
		if tacacsState.ServerIp == serverIp {
			sliceEntIdx = i
			break
		}
	}
	if sliceEntIdx != -1 {
		server.tacacsMgr.TacacsStateSlice = append(server.tacacsMgr.TacacsStateSlice[:sliceEntIdx],
			server.tacacsMgr.TacacsStateSlice[sliceEntIdx+1:]...)
	}
	delete(server.tacacsMgr.TacacsStateMap, serverIp)
	if server.tacacsMgr.TacacsGlobalStateObj != nil &&
		server.tacacsMgr.TacacsGlobalStateObj.OperStatus == "UP" {

		server.reconfigureServers()
	}
	return true, nil
}

func (server *SYSDServer) GetTacacsState(
	serverIp string) (*TacacsState, error) {

	if val, ok := server.tacacsMgr.TacacsStateMap[serverIp]; ok {
		return val, nil
	} else {
		return nil, errors.New("Tacacs Object does not exist")
	}
}

func (server *SYSDServer) GetTacacsStateSlice(
	fromIdx, count int) (int, int, bool, []*TacacsState) {

	var nextIdx int
	var more bool
	var actualCount int
	length := len(server.tacacsMgr.TacacsStateSlice)
	if fromIdx < 0 || fromIdx >= length || count <= 0 {
		return 0, 0, false, []*TacacsState{}
	}
	if fromIdx+count >= length {
		actualCount = length - fromIdx
		nextIdx = 0
		more = false
	} else {
		actualCount = count
		nextIdx = fromIdx + count
		more = true
	}

	result := make([]*TacacsState, actualCount)
	copy(result, server.tacacsMgr.TacacsStateSlice[fromIdx:fromIdx+actualCount])
	return nextIdx, actualCount, more, result
}

func (server *SYSDServer) CreateTacacsGlobalConfig(
	tacacsGlobalConfig *TacacsGlobalConfig) (bool, error) {

	server.logger.Info("Checking TACACS library dependencies")
	if !server.checkTacacsLib() {
		return false, errors.New("Library dependency check for tacacs failed")
	}
	server.logger.Info("TACACS library dependencies satisified")
	server.setupTacacsConfig()

	server.tacacsMgr.TacacsGlobalStateObj = &TacacsGlobalState{}
	server.tacacsMgr.TacacsGlobalStateObj.ProfileName =
		tacacsGlobalConfig.ProfileName
	if tacacsGlobalConfig.Enable == "true" {
		server.tacacsMgr.TacacsGlobalStateObj.OperStatus = "UP"
	} else {
		server.tacacsMgr.TacacsGlobalStateObj.OperStatus = "DOWN"
	}
	server.tacacsMgr.TacacsGlobalStateObj.NumActiveSessions = 0
	return true, nil
}

func (server *SYSDServer) UpdateTacacsGlobalConfig(
	tacacsGlobalConfig *TacacsGlobalConfig) (bool, error) {

	server.tacacsMgr.TacacsGlobalStateObj.ProfileName =
		tacacsGlobalConfig.ProfileName
	if tacacsGlobalConfig.Enable == "true" {
		server.tacacsMgr.TacacsGlobalStateObj.OperStatus = "UP"
	} else {
		server.tacacsMgr.TacacsGlobalStateObj.OperStatus = "DOWN"
	}
	server.tacacsMgr.TacacsGlobalStateObj.NumActiveSessions = 0

	if server.tacacsMgr.TacacsGlobalStateObj.OperStatus == "UP" {
		server.reconfigureServers()
	} else {
		server.resetServers()
	}
	return true, nil
}

func (server *SYSDServer) GetTacacsGlobalState(
	profileName string) (*TacacsGlobalState, error) {

	if server.tacacsMgr.TacacsGlobalStateObj != nil {
		return server.tacacsMgr.TacacsGlobalStateObj, nil
	} else {
		return nil, errors.New("Tacacs Global Object does not exist")
	}
}

func (server *SYSDServer) GetTacacsGlobalStateSlice(
	fromIdx, count int) (int, int, bool, []*TacacsGlobalState) {

	globalStateSlice := []*TacacsGlobalState{}
	if server.tacacsMgr.TacacsGlobalStateObj != nil {
		globalStateSlice = append(
			globalStateSlice, server.tacacsMgr.TacacsGlobalStateObj)
	}
	var nextIdx int
	var more bool
	var actualCount int
	length := len(globalStateSlice)
	if fromIdx < 0 || fromIdx >= length || count <= 0 {
		return 0, 0, false, []*TacacsGlobalState{}
	}
	if fromIdx+count >= length {
		actualCount = length - fromIdx
		nextIdx = 0
		more = false
	} else {
		actualCount = count
		nextIdx = fromIdx + count
		more = true
	}

	result := make([]*TacacsGlobalState, actualCount)
	copy(result, globalStateSlice[fromIdx:fromIdx+actualCount])
	return nextIdx, actualCount, more, result
}
