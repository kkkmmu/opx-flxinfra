namespace go portdServices
typedef i32 int
service PortService 
{
    int createV4Intf (1:int ipAddr, 2:int intf);
    int deleteV4Intf (1:int ipAddr);
}
