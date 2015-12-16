package main

import (
	"asicdServices"
	"encoding/json"
	_ "errors"
	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/vishvananda/netlink"
	"infra/portd/portdCommonDefs"
	"io/ioutil"
	"l3/rib/ribdCommonDefs"
	"net"
	"portdServices"
	"ribd"
	"strconv"
	"time"
	"asicd/asicdConstDefs"
//	"unsafe"
	"github.com/op/go-nanomsg"
	"encoding/binary"
//	"os"
	"syscall"
	"os/exec"
	"bytes"
	"strings"
)

type PortServiceHandler struct {
}

const (
	SUB_ASICD = 0
)

type IntfId struct {
	ifType  int
	ifIndex int
}
type IntfRecord struct {
	ifName string
	state  int  
	memberIfList []int
	activeIfCount int8
	parentId int
}

var AsicLinuxIfMapTable map[IntfId]IntfRecord

type PortClientBase struct {
	Address            string
	Transport          thrift.TTransport
	PtrProtocolFactory *thrift.TBinaryProtocolFactory
	IsConnected        bool
}

type AsicdClient struct {
	PortClientBase
	ClientHdl *asicdServices.AsicdServiceClient
}

type RibClient struct {
	PortClientBase
	ClientHdl *ribd.RouteServiceClient
}

type ClientJson struct {
	Name string `json:Name`
	Port int    `json:Port`
}

type PortConfigJson struct {
	Port   int    `json:Port`
	Ifname string `json:Ifname`
}

var asicdclnt AsicdClient
var ribdclnt RibClient
var sviBase = "SVI"
var linkAttrs netlink.LinkAttrs
var dummyLinkAttrs netlink.LinkAttrs
var PORT_PUB  *nanomsg.PubSocket
var AsicdSub *nanomsg.SubSocket

func parsePortRange(portStr string) (int, int, error) {
	portNums := strings.Split(portStr, "-")
	startPort, err := strconv.Atoi(portNums[0])
	if err != nil {
		return 0, 0, err
	}
	endPort, err := strconv.Atoi(portNums[1])
	if err != nil {
		return 0, 0, err
	}
	return startPort, endPort, nil
}

/*
 * Utility function to parse from a user specified port string to a port bitmap.
 * Supported formats for port string shown below:
 * - 1,2,3,10 (comma separate list of ports)
 * - 1-10,24,30-31 (hypen separated port ranges)
 * - 00011 (direct port bitmap)
 */
func parseUsrPortStrToPbm(usrPortStr string) (string, error) {
	//FIXME: Assuming max of 256 ports, create common def (another instance in main.go)
	var portList [256]int
	var pbmStr string = ""
	//Handle ',' separated strings
	if strings.Contains(usrPortStr, ",") {
		commaSepList := strings.Split(usrPortStr, ",")
		for _, subStr := range commaSepList {
			//Substr contains '-' separated range
			if strings.Contains(subStr, "-") {
				startPort, endPort, err := parsePortRange(subStr)
				if err != nil {
					return pbmStr, err
				}
				for port := startPort; port <= endPort; port++ {
					portList[port] = 1
				}
			} else {
				//Substr is a port number
				port, err := strconv.Atoi(subStr)
				if err != nil {
					return pbmStr, err
				}
				portList[port] = 1
			}
		}
	} else if strings.Contains(usrPortStr, "-") {
		//Handle '-' separated range
		startPort, endPort, err := parsePortRange(usrPortStr)
		if err != nil {
			return pbmStr, err
		}
		for port := startPort; port <= endPort; port++ {
			portList[port] = 1
		}
	} else {
        if len(usrPortStr) > 1 {
            //Port bitmap directly specified
            return usrPortStr, nil
        } else {
            //Handle single port number
            port, err := strconv.Atoi(usrPortStr)
            if err != nil {
                return pbmStr, err
            }
            portList[port] = 1
        }
	}
	//Convert portList to port bitmap string
	var zeroStr string = ""
	for _, port := range portList {
		if port == 1 {
			pbmStr += zeroStr
			pbmStr += "1"
			zeroStr = ""
		} else {
			zeroStr += "0"
		}
	}
	return pbmStr, nil
}

func vlanLinkCreate(ifName string, vlanId int32) (link netlink.Link, err error) {
	linkAttrs.Name = ifName + "." + strconv.Itoa(int(vlanId))
	//get the parent link's index
	parentIfLink, err := netlink.LinkByName(ifName)
	if err != nil {
		logger.Println("Error getting link info for ", ifName)
		return link, err
	}
	ParentlinkAttrs := parentIfLink.Attrs()
	linkAttrs.ParentIndex = ParentlinkAttrs.Index
	logger.Printf("parentIndex %d for ifName %s", linkAttrs.ParentIndex, ifName)
	vlanlink := &netlink.Vlan{linkAttrs, int(vlanId)}
	err = netlink.LinkAdd(vlanlink)
	if(err != nil) {
		logger.Println("err from linkAdd = ", err)
	   return vlanlink, err
	}
	err = netlink.LinkSetUp(vlanlink)
	if(err != nil) {
		logger.Println("err from linkSetup = ", err)
	   return vlanlink, err
	}
	linkAttrs = dummyLinkAttrs
	return vlanlink, err
}
func bridgeLinkCreate(brname string) (link netlink.Link, err error) {
	logger.Println("in bridge create for brname ", brname)
	linkAttrs.Name = brname
	logger.Println("linkAttrs.Name=", linkAttrs.Name)
	mybridge := &netlink.Bridge{linkAttrs}
	err = netlink.LinkAdd(mybridge)
	if(err != nil) {
		logger.Println("err from linkAdd = ", err)
	   return mybridge, err
	}
	err = netlink.LinkSetUp(mybridge)
	if(err != nil) {
		logger.Println("err from linkSetup = ", err)
	   return mybridge, err
	}
	linkAttrs = dummyLinkAttrs
	return mybridge, err
}

func OsCommand(binary string, args []string, env []string) {
	err := syscall.Exec(binary, args, env)
	//err = exec.Command(binary, command).Run()
	if(err != nil) {
		logger.Println("command for bridge link returned err", err, "when running args", args)
	}
	logger.Println("Completed OSCommand")
}
func addVlanLinkToBridge(vlanLink netlink.Link, bridgeLink *netlink.Bridge, vlanId int32) (err error) {
	logger.Println("Add vlan link to bridge link")
/*	err = netlink.LinkSetMaster(vlanLink, bridgeLink)
	if(err != nil) {
		logger.Println("Err ", err, "when setting master index for vlanlink")
	}
*/	
	parentIfLink, err := netlink.LinkByIndex(vlanLink.Attrs().ParentIndex)
	if err != nil {
		logger.Println("Error getting parent link info " )
		return err
	}
    err = netlink.LinkSetMaster(parentIfLink, bridgeLink)
	if(err != nil) {
		logger.Println("Err ", err, "when setting master index for parentlink")
	}

    brname := bridgeLink.Attrs().Name

	//echo 1 > /sys/class/net/<brname>/bridge/vlan_filtering 
	fileName := "/sys/class/net/"+brname+"/bridge/vlan_filtering"
	logger.Println("Reading file ", fileName)
    _, err = ioutil.ReadFile(fileName)
    if(err != nil) {
		logger.Println("Error ", err, "reading from file ", fileName)
		return  err
	}	
	data := "1"

	if err = ioutil.WriteFile(fileName, []byte(data), 0644); err != nil {
		logger.Println("Error ", err, "writing to file ", fileName)
	}

   /*** Using the netlink way ******/
    //bridge vlan add dev <vlanlink> vid <vid> pvid untagged
    /*err = netlink.LinkSetPvid(vlanLink, int(vlanId))
	if(err != nil) {
		logger.Println("linksetpvid for vlan link returned err=", err)
		return err
	}
	err = netlink.LinkSetUntagged(vlanLink)
	if(err != nil) {
		logger.Println("LinkSetUntagged for vlan link returned err=", err)
	}
    //bridge vlan add dev <bridgelink> vid <vid> self pvid untagged
    err = netlink.LinkSetPvid(bridgeLink, int(vlanId))
	if(err != nil) {
		logger.Println("linksetpvid for vlan link returned err=", err)
	}
	err = netlink.LinkSetUntagged(bridgeLink)
	if(err != nil) {
		logger.Println("LinkSetUntagged for bridge link returned err=", err)
	}
	err = netlink.LinkSetSelf(bridgeLink)
	if(err != nil) {
		logger.Println("LinkSelf for bridge link returned err=", err)
	}*/
	
	/*** temporary hack to call exec command****/
	binary, lookErr := exec.LookPath("bridge")
	if lookErr != nil {
		logger.Println("bridge not found lookerr = ", lookErr)
		return lookErr
	}
	logger.Println("path search for bridge found as ", binary)
	/*args1 := []string{"bridge", "vlan", "add", "dev", brname, "vid", strconv.Itoa(int(vlanId)), "self", "pvid", "untagged"}
	env := os.Environ()
	go OsCommand (binary, args1, env)
	logger.Println("configured bridge successfully")
//	vlanName := vlanLink.Attrs().Name
//	args2 := []string{"bridge", "vlan", "add", "dev", vlanName, "vid", strconv.Itoa(int(vlanId)),  "pvid", "untagged"}
	//err = exec.Command(binary, command).Run()
//	env = os.Environ()
//	go OsCommand (binary, args2, env)
//	logger.Println("configured vlanLink successfully")*/
    cmd := exec.Command(binary, "vlan", "add", "dev", brname, "vid", strconv.Itoa(int(vlanId)), "self", "pvid", "untagged")
	err = cmd.Run()
	if(err != nil) {
		logger.Println("Error executing bridge vlan command for bridge")
	}
/*
	vlanName := vlanLink.Attrs().Name
    cmd = exec.Command(binary, "vlan", "add", "dev", vlanName, "vid", strconv.Itoa(int(vlanId)), "pvid", "untagged")
	err = cmd.Run()
	if(err != nil) {
		logger.Println("Error executing command ")
	}
*/
	parentIfName := parentIfLink.Attrs().Name
    logger.Printf("Found the parent interface as %s, now add this as untagged member\n", parentIfName)

    cmd = exec.Command(binary, "vlan", "add", "dev", parentIfName, "vid", strconv.Itoa(int(vlanId)), "pvid", "untagged")
	err = cmd.Run()
	if(err != nil) {
		logger.Println("Error executing bridge vlan command for parentiflink")
	}

	return err
}

func (m PortServiceHandler) CreateV4Intf(ipAddr string,
	intf int32,
	vlanEnabled int32) (rc portdServices.Int, err error) {
	logger.Println("Received create intf request")
	var ipMask net.IP
	var link netlink.Link
	if asicdclnt.IsConnected == true {
		_,err = asicdclnt.ClientHdl.CreateIPv4Intf(ipAddr, intf) //need to pass vlanEnabled here to asic
		if(err != nil) {
			logger.Println("asicd returned error ", err)
		}
	}
	logger.Println("Finished calling asicd")
	if ribdclnt.IsConnected == true {
		var nextHopIfType int
		ip, ipNet, err := net.ParseCIDR(ipAddr)
		if err != nil {
			return -1, err
		}
		ipMask = make(net.IP, 4)
		copy(ipMask, ipNet.Mask)
		ipAddrStr := ip.String()
		ipMaskStr := net.IP(ipMask).String()
		if vlanEnabled == 1 {
			nextHopIfType = portdCommonDefs.VLAN
		} else {
			nextHopIfType = portdCommonDefs.PHY
		}
		ribdclnt.ClientHdl.CreateV4Route(ipAddrStr, ipMaskStr, 0, "0.0.0.0", ribd.Int(nextHopIfType), ribd.Int(intf), ribdCommonDefs.CONNECTED)
	}
	logger.Println("Finished calling ribd")
	if vlanEnabled == 1 {
		//set the ip interface on bridge<vlan>
		
		brname := sviBase + strconv.Itoa(int(intf))
		logger.Println("looking for bridge ", brname)
		link, err = netlink.LinkByName(brname)
		if link == nil {
			logger.Println("Could not find bridge err=", brname, err)
			return 0, err
		}
/*		//For now, assign ip on the first mmber interface of the vlan
		vlanintfId := IntfId{ifType: portdCommonDefs.VLAN, ifIndex: int(intf)}
		vlanintfRecord, ok := AsicLinuxIfMapTable[vlanintfId]
		if !ok {
			logger.Println(" could not find SVI mapping for vlan ", intf)
			return 0, err
		}
		intfId := IntfId{ifType:portdCommonDefs.PHY, ifIndex:int(vlanintfRecord.memberIfList[0])}
		linkName := AsicLinuxIfMapTable[intfId].ifName
		link, err = netlink.LinkByName(linkName)
		if link == nil {
			logger.Println("Could not find interface ", linkName)
			return 0, err
		}
	*/	
	} else {
		//set ip interface on the actual interface derived from looking up the asictolinuxmap
		intfId := IntfId{ifType: portdCommonDefs.PHY, ifIndex: int(intf)}
		intfRecord, ok := AsicLinuxIfMapTable[intfId]
		if !ok {
			logger.Println(" could not find SVI mapping for port")
			return 0, err
		}
		link, err = netlink.LinkByName(intfRecord.ifName)
		if link == nil {
			logger.Println("Could not find interface ", intfRecord.ifName)
			return 0, err
		}
	}

	addr, err := netlink.ParseAddr(ipAddr)
	if err != nil {
		logger.Println("error while parsing ip address ", err, ipAddr)
		return 0, err
	}
	err = netlink.AddrAdd(link, addr)
	if err != nil {
		logger.Println("error while assigning ip address to SVI ", err)
		return 0, err
	}
	return 0, err
}

func (m PortServiceHandler) DeleteV4Intf(ipAddr string,
	intf int32) (rc portdServices.Int, err error) {
	logger.Println("Received Intf Delete request")
	var ipMask net.IP
	if asicdclnt.IsConnected == true {
		asicdclnt.ClientHdl.DeleteIPv4Intf(ipAddr, intf)
	}
	if ribdclnt.IsConnected == true {
		ip, ipNet, err := net.ParseCIDR(ipAddr)
		if err != nil {
			return -1, err
		}
		ipMask = make(net.IP, 4)
		copy(ipMask, ipNet.Mask)
		ipAddrStr := ip.String()
		ipMaskStr := net.IP(ipMask).String()
		ribdclnt.ClientHdl.DeleteV4Route(ipAddrStr, ipMaskStr, ribdCommonDefs.CONNECTED)
	}
	return 0, err
}

func (m PortServiceHandler) CreateV4Neighbor(
	ipAddr string,
	macAddr string,
	vlanId int32,
	routerIntf int32) (rc portdServices.Int, err error) {
	logger.Println("Received create neighbor intf request")
	if asicdclnt.IsConnected == true {
		asicdclnt.ClientHdl.CreateIPv4Neighbor(ipAddr, macAddr, vlanId, routerIntf)
	}
	return 0, err
}

func (m PortServiceHandler) DeleteV4Neighbor(ipAddr string, macAddr string, vlanId int32, routerIntf int32) (rc portdServices.Int, err error) {
	logger.Println("Received delete neighbor intf request")
	if asicdclnt.IsConnected == true {
		asicdclnt.ClientHdl.DeleteIPv4Neighbor(ipAddr, macAddr, vlanId, routerIntf)
	}
	return 0, err
}
func (m PortServiceHandler) GetVlanMembers(vlanId int32) (ports []string, err error) {
	
	ports = make([]string,0)
	intfId := IntfId{ifType: portdCommonDefs.VLAN, ifIndex: int(vlanId)}
	intfRecord, ok := AsicLinuxIfMapTable[intfId]
	if(!ok) {
		logger.Printf("vlan %d not in the map", int(vlanId))
		return ports, err
	}
	logger.Println("iftype", portdCommonDefs.VLAN, "ifindex", vlanId, "ifname", intfRecord.ifName)
/*	bridgeLink, err := netlink.LinkByName(intfRecord.ifName)
	return ports, nil
	if(err != nil) {
		logger.Printf("bridge %s not configured\n", intfRecord.ifName)
		return ports, err
	}
	for _,v := range AsicLinuxIfMapTable {
		link, err:= netlink.LinkByName(v.ifName)
		if(err == nil) {
			if(link.Attrs().MasterIndex == bridgeLink.Attrs().Index){
				ports = append(ports, v.ifName)
			}
		}
	}*/
	for i:=0;i<len(intfRecord.memberIfList);i++ {
		getIntfId := IntfId{ifType:portdCommonDefs.PHY, ifIndex:intfRecord.memberIfList[i]}
		getIfRecord, ok := AsicLinuxIfMapTable[getIntfId]
		if(!ok){
			continue
		}
		ports = append(ports, getIfRecord.ifName)
	}
	return ports, nil
}

func (m PortServiceHandler) CreateVlan(vlanId int32,
	ports string,
	untaggedPorts string) (rc portdServices.Int, err error) {
	logger.Println("create vlan")
    portPbmStr, err := parseUsrPortStrToPbm(ports)
    if err != nil {
        return 0, err
    }
	var brintfId IntfId
	//call asicd to create vlan and add members in the switch
	if asicdclnt.IsConnected == true {
		asicdclnt.ClientHdl.CreateVlan(vlanId, ports, untaggedPorts)
	}
	
	//create bridgelink - SVI<vlan>
	brname := sviBase + strconv.Itoa(int(vlanId))
	logger.Println("looking for bridge ", brname)
	bridgeLink, err := netlink.LinkByName(brname)
	if bridgeLink == nil {
		bridgeLink, err = bridgeLinkCreate(brname)
		if bridgeLink == nil {
			logger.Println("Could not create bridge err=", brname, err)
			return 0, err
		}
		brintfId.ifType = portdCommonDefs.VLAN
		brintfId.ifIndex = int(vlanId)
		intfRecord := IntfRecord{ifName:brname, state:portdCommonDefs.LINK_STATE_UP}
		AsicLinuxIfMapTable[brintfId] = intfRecord
		logger.Println("Added entry type:index", brintfId.ifType, ":", brintfId.ifIndex, ":", brname)
	}
	//go over the ports in the portlist
	for i := 0; i < len(portPbmStr); i++ {
		if portPbmStr[i] == '1' {
			//get the linux names from the map
			intfId := IntfId{ifType: portdCommonDefs.PHY, ifIndex: i}
			intfRecord, ok := AsicLinuxIfMapTable[intfId]
			if !ok {
				logger.Println("No linux mapping found for the front panel port err ", i, err)
				return 0, err
			}
			//create virtual vlan interface
			vlanLink, err := vlanLinkCreate(intfRecord.ifName, vlanId)
			if err != nil {
				logger.Println("Could not create vlan interface for port err ", i, err)
				return 0, err
			}
			//add the vlan interface to the bridge
			err = addVlanLinkToBridge(vlanLink, bridgeLink.(*netlink.Bridge), vlanId)
			if err != nil {
				logger.Println("Could not add vlan interface ifName to bridge  err ", intfRecord.ifName, brname, err)
				return 0, err
			}
			logger.Println("Added vlanlink to bridge")
			brIntfRecord, ok := AsicLinuxIfMapTable[brintfId]
			if(!ok){
				return 0, nil
			} 
			if(len(brIntfRecord.memberIfList) == 0) {
				logger.Println("Making the memberlist because this is the first member")
				brIntfRecord.memberIfList = make([]int, 0)
			}
			logger.Printf("Adding member port %d to vlan %d\n", i,  vlanId)
			brIntfRecord.memberIfList = append(brIntfRecord.memberIfList, i)
			if(intfRecord.state == portdCommonDefs.LINK_STATE_UP) {
				logger.Println("adding a link up member")
				brIntfRecord.activeIfCount++
			}
			AsicLinuxIfMapTable[brintfId] = brIntfRecord
			logger.Printf("activeIfCount for intfId %d:%d is %d\n", brintfId.ifType,brintfId.ifIndex, brIntfRecord.activeIfCount)
			
			intfRecord.parentId = int(vlanId)
			AsicLinuxIfMapTable[intfId] = intfRecord
		}
	}
	return 0, nil
}

func (m PortServiceHandler) DeleteVlan(vlanId int32,
	ports string,
	portTagType string) (rc portdServices.Int, err error) {
	logger.Println("Delete vlan")
	if asicdclnt.IsConnected == true {
		logger.Println("call deletevlan from asicd")
		asicdclnt.ClientHdl.DeleteVlan(vlanId, ports, portTagType)
	}
	return 0, nil
}

func (m PortServiceHandler) UpdateVlan(vlanId int32,
	ports string,
	portTagType string) (rc portdServices.Int, err error) {
	if asicdclnt.IsConnected == true {
		asicdclnt.ClientHdl.UpdateVlan(vlanId, ports, portTagType)
	}
	return 0, nil
}

func (m PortServiceHandler) GetLinuxIfc(ifType int32, ifIndex int32) (ifName string, err error) {
	intfId := IntfId{ifType: int(ifType), ifIndex: int(ifIndex)}
	intfRecord, _ := AsicLinuxIfMapTable[intfId]
	logger.Println("iftype", ifType, "idindex", ifIndex, "ifname", intfRecord.ifName)
	return intfRecord.ifName, nil
}
func processLinkDown(intfId int) {
	logger.Println("processLinkDown for port ", intfId)
	intfIndex := IntfId{ifType:portdCommonDefs.PHY, ifIndex: intfId}
	intfRecord, ok := AsicLinuxIfMapTable[intfIndex]
	if(!ok) {
		logger.Println("Could not find port", intfId)
		return
	}
	intfRecord.state = portdCommonDefs.LINK_STATE_DOWN
	logger.Println("set state for intf record")
	AsicLinuxIfMapTable[intfIndex] = intfRecord
	parentIntfId := IntfId {ifType:portdCommonDefs.VLAN, ifIndex:intfRecord.parentId}
	parentIntfRecord, ok := AsicLinuxIfMapTable[parentIntfId]
	if(!ok) {
		logger.Println("Vlan ",intfRecord.parentId, "not found")
		return
	}
	parentIntfRecord.activeIfCount--
	if(parentIntfRecord.activeIfCount == 0) {
		parentIntfRecord.state = portdCommonDefs.LINK_STATE_DOWN
	}
	AsicLinuxIfMapTable[parentIntfId] = parentIntfRecord
	logger.Printf("Set parent link %d:%d state activeIfCount = \n", parentIntfId.ifType, parentIntfId.ifIndex, parentIntfRecord.activeIfCount)
	if(parentIntfRecord.activeIfCount == 0) {
		//publish link down event for the vlan
	
	   msgBuf := portdCommonDefs.LinkStateInfo{LinkType:portdCommonDefs.VLAN, LinkId:uint8(parentIntfId.ifIndex), LinkStatus:portdCommonDefs.LINK_STATE_DOWN}
	   msgbufbytes, err := json.Marshal( msgBuf)
       msg := portdCommonDefs.PortdNotifyMsg {MsgType:portdCommonDefs.NOTIFY_LINK_STATE_CHANGE, MsgBuf: msgbufbytes}
	   buf, err := json.Marshal( msg)
	   if err != nil {
		 logger.Println("Error in marshalling Json")
		 return
	   }
	   logger.Println("buf", buf)
   	   PORT_PUB.Send(buf, nanomsg.DontWait)
	}
}
func processLinkUp(intfId int) {
	logger.Println("processLinkUp for port ", intfId)
	intfIndex := IntfId{ifType:portdCommonDefs.PHY, ifIndex: intfId}
	intfRecord, ok := AsicLinuxIfMapTable[intfIndex]
	if(!ok) {
		logger.Println("Could not find port", intfId)
		return
	}
	intfRecord.state = portdCommonDefs.LINK_STATE_UP
	logger.Println("set state for intf record")
	AsicLinuxIfMapTable[intfIndex] = intfRecord

	parentIntfId := IntfId {ifType:portdCommonDefs.VLAN, ifIndex:intfRecord.parentId}
	parentIntfRecord, ok := AsicLinuxIfMapTable[parentIntfId]
	if(!ok) {
		logger.Println("Vlan ",intfRecord.parentId, "not found")
		return
	}
	if(parentIntfRecord.activeIfCount == 0) {
		parentIntfRecord.state = portdCommonDefs.LINK_STATE_UP
	}
	parentIntfRecord.activeIfCount++
	AsicLinuxIfMapTable[parentIntfId] = parentIntfRecord
	logger.Printf("Set parent link %d:%d state activeIfCount = \n", parentIntfId.ifType, parentIntfId.ifIndex, parentIntfRecord.activeIfCount)
	if(parentIntfRecord.activeIfCount == 1) {
		//publish link up event for the vlan
	
	   msgBuf := portdCommonDefs.LinkStateInfo{LinkType:portdCommonDefs.VLAN, LinkId:uint8(parentIntfId.ifIndex), LinkStatus:portdCommonDefs.LINK_STATE_UP}
	   msgbufbytes, err := json.Marshal( msgBuf)
       msg := portdCommonDefs.PortdNotifyMsg {MsgType:portdCommonDefs.NOTIFY_LINK_STATE_CHANGE, MsgBuf: msgbufbytes}
	   buf, err := json.Marshal( msg)
	   if err != nil {
		 logger.Println("Error in marshalling Json")
		 return
	   }
	   logger.Println("buf", buf)
   	   PORT_PUB.Send(buf, nanomsg.DontWait)
	}
}

func (m PortServiceHandler) LinkDown(ifIndex int32) (err error){
	logger.Println("Disable port ", ifIndex)
	processLinkDown(int(ifIndex))
	return err
}

func CreateIPCHandles(address string) (thrift.TTransport, *thrift.TBinaryProtocolFactory) {
	var transportFactory thrift.TTransportFactory
	var transport thrift.TTransport
	var protocolFactory *thrift.TBinaryProtocolFactory
	var err error

	protocolFactory = thrift.NewTBinaryProtocolFactoryDefault()
	transportFactory = thrift.NewTTransportFactory()
	transport, err = thrift.NewTSocket(address)
	transport = transportFactory.GetTransport(transport)
	if err = transport.Open(); err != nil {
	//	logger.Println("Failed to Open Transport", transport, protocolFactory)
		return nil, nil
	}
	return transport, protocolFactory
}

func connectToClient(client ClientJson) {
	var timer *time.Timer
	logger.Printf("in go routine ConnectToClient for connecting to %s\n", client.Name)
	for {
		timer = time.NewTimer(time.Second * 10)
		<-timer.C
		if client.Name == "asicd" {
		//	logger.Printf("found asicd at port %d", client.Port)
			asicdclnt.Address = "localhost:" + strconv.Itoa(client.Port)
			asicdclnt.Transport, asicdclnt.PtrProtocolFactory = CreateIPCHandles(asicdclnt.Address)
			if asicdclnt.Transport != nil && asicdclnt.PtrProtocolFactory != nil {
				logger.Println("connecting to asicd")
				asicdclnt.ClientHdl = asicdServices.NewAsicdServiceClientFactory(asicdclnt.Transport, asicdclnt.PtrProtocolFactory)
				asicdclnt.IsConnected = true
				timer.Stop()
				return
			}
		}
		if client.Name == "ribd" {
		//	logger.Printf("found ribd at port %d", client.Port)
			ribdclnt.Address = "localhost:" + strconv.Itoa(client.Port)
			ribdclnt.Transport, ribdclnt.PtrProtocolFactory = CreateIPCHandles(ribdclnt.Address)
			if ribdclnt.Transport != nil && ribdclnt.PtrProtocolFactory != nil {
				logger.Println("connecting to ribd")
				ribdclnt.ClientHdl = ribd.NewRouteServiceClientFactory(ribdclnt.Transport, ribdclnt.PtrProtocolFactory)
				ribdclnt.IsConnected = true
				timer.Stop()
				return
			}
		}
	}
}

func ConnectToClients(paramsFile string) {
	var clientsList []ClientJson

	bytes, err := ioutil.ReadFile(paramsFile)
	if err != nil {
		logger.Println("Error in reading client configuration file")
		return
	}

	err = json.Unmarshal(bytes, &clientsList)
	if err != nil {
		logger.Println("Error in Unmarshalling Json")
		return
	}

	for _, client := range clientsList {
		logger.Println("#### Client name is ", client.Name)
		if client.Name == "asicd" {
			logger.Printf("found asicd at port %d", client.Port)
			asicdclnt.Address = "localhost:" + strconv.Itoa(client.Port)
			asicdclnt.Transport, asicdclnt.PtrProtocolFactory = CreateIPCHandles(asicdclnt.Address)
			if asicdclnt.Transport != nil && asicdclnt.PtrProtocolFactory != nil {
				logger.Println("connecting to asicd")
				asicdclnt.ClientHdl = asicdServices.NewAsicdServiceClientFactory(asicdclnt.Transport, asicdclnt.PtrProtocolFactory)
				asicdclnt.IsConnected = true
			} else {
				go connectToClient(client)
			}
		}
		if client.Name == "ribd" {
			logger.Printf("found ribd at port %d", client.Port)
			ribdclnt.Address = "localhost:" + strconv.Itoa(client.Port)
			ribdclnt.Transport, ribdclnt.PtrProtocolFactory = CreateIPCHandles(ribdclnt.Address)
			if ribdclnt.Transport != nil && ribdclnt.PtrProtocolFactory != nil {
				logger.Println("connecting to ribd")
				ribdclnt.ClientHdl = ribd.NewRouteServiceClientFactory(ribdclnt.Transport, ribdclnt.PtrProtocolFactory)
				ribdclnt.IsConnected = true
			} else {
				go connectToClient(client)
			}
		}
	}
}

func BuildAsicToLinuxMap(cfgFile string) {
/*	var portCfgList []PortConfigJson
	bytes, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		logger.Println("Error in reading port configuration file")
		return
	}
	err = json.Unmarshal(bytes, &portCfgList)
	if err != nil {
		logger.Println("Error in Unmarshalling Json, err=", err)
		return
	}
	for _, v := range portCfgList {
		intfId := IntfId{ifType: portdCommonDefs.PHY, ifIndex: v.Port}
		intfRecord := IntfRecord{ifName:v.Ifname, state:portdCommonDefs.LINK_STATE_UP}
		AsicLinuxIfMapTable[intfId] = intfRecord
	}
*/
	for i:=1;i<73;i++ {
		intfId := IntfId{ifType: portdCommonDefs.PHY, ifIndex: i}
		ifName := "fpPort-"+strconv.Itoa(i)
		intfRecord := IntfRecord{ifName:ifName, state:portdCommonDefs.LINK_STATE_UP}
		AsicLinuxIfMapTable[intfId] = intfRecord
   }
/*   logger.Println("Now install a dummy entry for eth0")
		intfId := IntfId{ifType: portdCommonDefs.PHY, ifIndex: 0}
		ifName := "eth0"
		intfRecord := IntfRecord{ifName:ifName, state:portdCommonDefs.LINK_STATE_UP}
		AsicLinuxIfMapTable[intfId] = intfRecord*/
}
func InitPublisher()(pub *nanomsg.PubSocket) {
	pub, err := nanomsg.NewPubSocket()
    if err != nil {
        logger.Println("Failed to open pub socket")
        return nil
    }
    ep, err := pub.Bind(portdCommonDefs.PUB_SOCKET_ADDR)
    if err != nil {
        logger.Println("Failed to bind pub socket - ", ep)
        return nil
    }
    err = pub.SetSendBuffer(1024*1024)
    if err != nil {
        logger.Println("Failed to set send buffer size")
        return nil
    }
	return pub
}
func processAsicdEvents(sub *nanomsg.SubSocket) {
	
	logger.Println("in process Asicd events")
    for {
	  logger.Println("In for loop")
      rcvdMsg,err := sub.Recv(0)
	  if(err != nil) {
	     logger.Println("Error in receiving ", err)
		 return	
	  }
	  logger.Println("After recv rcvdMsg buf", rcvdMsg)
      buf := bytes.NewReader(rcvdMsg)
      var MsgType asicdConstDefs.AsicdNotifyMsg
      err = binary.Read(buf, binary.LittleEndian, &MsgType)
      if err != nil {
	     logger.Println("Error in reading msgtype ", err)
		 return	
      }
      switch MsgType {
        case asicdConstDefs.NOTIFY_LINK_STATE_CHANGE:
           var msg asicdConstDefs.LinkStateInfo
           err = binary.Read(buf, binary.LittleEndian, &msg)
           if err != nil {
    	     logger.Println("Error in reading msg ", err)
		     return	
           }
		    logger.Printf("Msg linkstatus = %d msg port = %d\n", msg.LinkStatus, msg.Port)
		    if(msg.LinkStatus == asicdConstDefs.LINK_STATE_DOWN) {
				processLinkDown(int(msg.Port))
			} else {
				processLinkUp(int(msg.Port))
			}
       }
	}
}

func processEvents(sub *nanomsg.SubSocket, subType ribd.Int) {
	logger.Println("in process events for sub ", subType)
	if(subType == SUB_ASICD){
		logger.Println("process Asicd events")
		processAsicdEvents(sub)
	} 
}
func setupEventHandler(sub *nanomsg.SubSocket, address string, subtype ribd.Int) {
	logger.Println("Setting up event handlers for sub type ", subtype)
	sub, err := nanomsg.NewSubSocket()
	 if err != nil {
        logger.Println("Failed to open sub socket")
        return
    }
	logger.Println("opened socket")
	ep, err := sub.Connect(address)
	if err != nil {
        logger.Println("Failed to connect to pub socket - ", ep)
        return
    }
	logger.Println("Connected to ", ep.Address)
	err = sub.Subscribe("")
	if(err != nil) {
		logger.Println("Failed to subscribe to all topics")
		return 
	}
	logger.Println("Subscribed")
	err = sub.SetRecvBuffer(1024 * 1204)
    if err != nil {
        logger.Println("Failed to set recv buffer size")
        return
    }
		//processPortdEvents(sub)
	processEvents(sub, subtype)
}

func NewPortServiceHandler(paramsDir string) *PortServiceHandler {
	AsicLinuxIfMapTable = make(map[IntfId]IntfRecord)
	configFile := paramsDir + "/clients.json"
	ConnectToClients(configFile)
	portCfgFile := paramsDir + "/portd.json"
	BuildAsicToLinuxMap(portCfgFile)
	PORT_PUB = InitPublisher()
	go setupEventHandler(AsicdSub, asicdConstDefs.PUB_SOCKET_ADDR, SUB_ASICD)
	linkAttrs = netlink.NewLinkAttrs()
	return &PortServiceHandler{}
}
