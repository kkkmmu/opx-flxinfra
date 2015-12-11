package portdCommonDefs

const (
      VLAN  = 0
      PHY     = 1
      PUB_SOCKET_ADDR = "ipc:///tmp/portd.ipc"	
	  NOTIFY_LINK_STATE_CHANGE = 1
	  LINK_STATE_DOWN = 0
      LINK_STATE_UP = 1
	  DEFAULT_NOTIFICATION_SIZE = 128
)
type PortdNotifyMsg struct {
    MsgType uint16
    MsgBuf []byte
}

type LinkStateInfo struct {
    Port uint8
    LinkStatus uint8
}
