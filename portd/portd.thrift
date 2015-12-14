namespace go portdServices
typedef i32 int
service PortService 
{
    int createV4Intf (1:string ipAddr, 2:i32 intf, 3:i32 vlanEnabled);
    int deleteV4Intf (1:string ipAddr, 2:i32 intf);
    int createV4Neighbor(1:string ipAddr, 2:string macAddr, 3:i32 vlanId, 4:i32 routerIntf);
    int deleteV4Neighbor(1:string ipAddr, 2:string macAddr, 3:i32 vlanId, 4:i32 routerIntf);
    int createVlan(1:i32 vlanId, 2:string ports, 3:string portTagType),
    int deleteVlan(1:i32 vlanId, 2:string ports, 3:string portTagType),
    int updateVlan(1:i32 vlanId, 2:string ports, 3:string portTagType),
	string getLinuxIfc(1:i32 ifType, 2:i32 ifIndex)
	void linkDown(1:i32 port)
	list<string> getVlanMembers(1:i32 vlanId)
}
