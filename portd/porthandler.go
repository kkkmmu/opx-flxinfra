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
)

type PortServiceHandler struct {
}
type IntfId struct {
	ifType  int
	ifIndex int
}

var AsicLinuxIfMapTable map[IntfId]string

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
		asicdclnt.ClientHdl.CreateIPv4Intf(ipAddr, intf) //need to pass vlanEnabled here to asic
	}
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
	if vlanEnabled == 1 {
		//set the ip interface on bridge<vlan>
		brname := sviBase + strconv.Itoa(int(intf))
		logger.Println("looking for bridge ", brname)
		link, err = netlink.LinkByName(brname)
		if link == nil {
			logger.Println("Could not find bridge err=", brname, err)
			return 0, err
		}
	} else {
		//set ip interface on the actual interface derived from looking up the asictolinuxmap
		intfId := IntfId{ifType: portdCommonDefs.PHY, ifIndex: int(intf)}
		ifName, ok := AsicLinuxIfMapTable[intfId]
		if !ok {
			logger.Println(" could not find SVI mapping for port")
			return 0, err
		}
		link, err = netlink.LinkByName(ifName)
		if link == nil {
			logger.Println("Could not find interface ", ifName)
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

func (m PortServiceHandler) CreateVlan(vlanId int32,
	ports string,
	portTagType string) (rc portdServices.Int, err error) {
	logger.Println("create vlan")
	//call asicd to create vlan and add members in the switch
	if asicdclnt.IsConnected == true {
		asicdclnt.ClientHdl.CreateVlan(vlanId, ports, portTagType)
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
		intfId := IntfId{ifType: portdCommonDefs.VLAN, ifIndex: int(vlanId)}
		AsicLinuxIfMapTable[intfId] = brname
		logger.Println("Added entry type:index", intfId.ifType, ":", intfId.ifIndex, ":", brname)
	}
	//go over the ports in the portlist
	for i := 0; i < len(ports); i++ {
		if ports[i] == '1' {
			//get the linux names from the map
			intfId := IntfId{ifType: portdCommonDefs.PHY, ifIndex: i}
			ifName, ok := AsicLinuxIfMapTable[intfId]
			if !ok {
				logger.Println("No linux mapping found for the front panel port err ", i, err)
				return 0, err
			}
			//create virtual vlan interface
			vlanLink, err := vlanLinkCreate(ifName, vlanId)
			if err != nil {
				logger.Println("Could not create vlan interface for port err ", i, err)
				return 0, err
			}
			//add the vlan interface to the bridge
			err = addVlanLinkToBridge(vlanLink, bridgeLink.(*netlink.Bridge))
			if err != nil {
				logger.Println("Could not add vlan interface ifName to bridge  err ", ifName, brname, err)
				return 0, err
			}
		}
	}
	return 0, nil
}

func (m PortServiceHandler) DeleteVlan(vlanId int32,
	ports string,
	portTagType string) (rc portdServices.Int, err error) {
	if asicdclnt.IsConnected == true {
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
	ifName, _ = AsicLinuxIfMapTable[intfId]
	logger.Println("iftype", ifType, "idindex", ifIndex, "ifname", ifName)
	return ifName, nil
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
		logger.Println("Failed to Open Transport", transport, protocolFactory)
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
			logger.Printf("found asicd at port %d", client.Port)
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
			logger.Printf("found ribd at port %d", client.Port)
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
	var portCfgList []PortConfigJson
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
	AsicLinuxIfMapTable = make(map[IntfId]string)
	for _, v := range portCfgList {
		intfId := IntfId{ifType: portdCommonDefs.PHY, ifIndex: v.Port}
		AsicLinuxIfMapTable[intfId] = v.Ifname
	}

}
func NewPortServiceHandler(paramsDir string) *PortServiceHandler {
	AsicLinuxIfMapTable = make(map[IntfId]string)
	configFile := paramsDir + "/clients.json"
	ConnectToClients(configFile)
	portCfgFile := paramsDir + "/portd.json"
	BuildAsicToLinuxMap(portCfgFile)
	linkAttrs = netlink.NewLinkAttrs()
	return &PortServiceHandler{}
}
