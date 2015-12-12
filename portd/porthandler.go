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
	"bytes"
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
	logger.Println("in brdge create for brname ", brname)
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

func addVlanLinkToBridge(vlanLink netlink.Link, bridgeLink *netlink.Bridge) (err error) {
	logger.Println("Add vlan link to bridge link")
	err = netlink.LinkSetMaster(vlanLink, bridgeLink)
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
		/*
		brname := sviBase + strconv.Itoa(int(intf))
		logger.Println("looking for bridge ", brname)
		link, err = netlink.LinkByName(brname)
		if link == nil {
			logger.Println("Could not find bridge err=", brname, err)
			return 0, err
		}*/
		//For now, assign ip on the first mmber interface of the vlan
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
	portTagType string) (rc portdServices.Int, err error) {
	logger.Println("create vlan")
	var brintfId IntfId
	//call asicd to create vlan and add members in the switch
	if asicdclnt.IsConnected == true {
		asicdclnt.ClientHdl.CreateVlan(vlanId, ports, portTagType)
	}
	
	//create bridgelink - SVI<vlan>
	brname := sviBase + strconv.Itoa(int(vlanId))
	/*logger.Println("looking for bridge ", brname)
	bridgeLink, err := netlink.LinkByName(brname)
	if bridgeLink == nil {
		bridgeLink, err = bridgeLinkCreate(brname)
		if bridgeLink == nil {
			logger.Println("Could not create bridge err=", brname, err)
			return 0, err
		}*/
		brintfId.ifType = portdCommonDefs.VLAN
		brintfId.ifIndex = int(vlanId)
		intfRecord := IntfRecord{ifName:brname, state:portdCommonDefs.LINK_STATE_UP}
		AsicLinuxIfMapTable[brintfId] = intfRecord
		logger.Println("Added entry type:index", brintfId.ifType, ":", brintfId.ifIndex, ":", brname)
	//}
	//go over the ports in the portlist
	for i := 0; i < len(ports); i++ {
		if ports[i] == '1' {
			//get the linux names from the map
			intfId := IntfId{ifType: portdCommonDefs.PHY, ifIndex: i}
			intfRecord, ok := AsicLinuxIfMapTable[intfId]
			if !ok {
				logger.Println("No linux mapping found for the front panel port err ", i, err)
				return 0, err
			}/*
			//create virtual vlan interface
			vlanLink, err := vlanLinkCreate(intfRecord.ifName, vlanId)
			if err != nil {
				logger.Println("Could not create vlan interface for port err ", i, err)
				return 0, err
			}
			//add the vlan interface to the bridge
			err = addVlanLinkToBridge(vlanLink, bridgeLink.(*netlink.Bridge))
			if err != nil {
				logger.Println("Could not add vlan interface ifName to bridge  err ", intfRecord.ifName, brname, err)
				return 0, err
			}*/
			brIntfRecord, ok := AsicLinuxIfMapTable[brintfId]
			if(!ok){
				return 0, nil
			} 
			if(len(brIntfRecord.memberIfList) == 0) {
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
func (m PortServiceHandler) LinkDown(ifIndex int32) (err error){
	logger.Println("Disable port ", ifIndex)
	processLinkDown(int(ifIndex))
	//intfId := IntfId{ifType: portdCommonDefs.PHY, ifIndex: int(ifIndex)}
	//ifName_, _ := AsicLinuxIfMapTable[intfId]
	//netlink call to disable link
	
/*	msgBuf := portdCommonDefs.LinkStateInfo{Port:uint8(ifIndex), LinkStatus:portdCommonDefs.LINK_STATE_DOWN}
    var msgBufPtr = unsafe.Pointer(&msgBuf)
	msgBufSlice := *((*[2]uint8)(msgBufPtr))
    var msg portdCommonDefs.PortdNotifyMsg
	msg.MsgType = portdCommonDefs.NOTIFY_LINK_STATE_CHANGE
	copy(msg.MsgBuf[:], msgBufSlice[:])
	var msgPtr = unsafe.Pointer(&msg)
	msgSlice := *((*[4]uint8)(msgPtr))
	var buf []byte
	buf = make([]byte, len(msgSlice))
    copy(buf[:], msgSlice[:])
	logger.Println("buf", buf)
   	PORT_PUB.Send(buf, nanomsg.DontWait)*/

/*	msgBuf := portdCommonDefs.LinkStateInfo{LinkType:portdCommonDefs.PHY,LinkId:uint8(ifIndex), LinkStatus:portdCommonDefs.LINK_STATE_DOWN}
	msgbufbytes, err := json.Marshal( msgBuf)
    msg := portdCommonDefs.PortdNotifyMsg {MsgType:portdCommonDefs.NOTIFY_LINK_STATE_CHANGE, MsgBuf: msgbufbytes}
	buf, err := json.Marshal( msg)
	if err != nil {
		logger.Println("Error in marshalling Json")
		return err
	}
	logger.Println("buf", buf)
   	PORT_PUB.Send(buf, nanomsg.DontWait)
*/
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
   logger.Println("Now install a dummy entry for eth0")
		intfId := IntfId{ifType: portdCommonDefs.PHY, ifIndex: 0}
		ifName := "eth0"
		intfRecord := IntfRecord{ifName:ifName, state:portdCommonDefs.LINK_STATE_UP}
		AsicLinuxIfMapTable[intfId] = intfRecord
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
