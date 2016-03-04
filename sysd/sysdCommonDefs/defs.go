package sysdCommonDefs

import ()

const (
	PUB_SOCKET_ADDR = "ipc:///tmp/sysd.ipc"
)

const (
	G_LOG uint8 = 1
	C_LOG uint8 = 2
)

//Logging levels
type SRDebugLevel uint8

const (
	CRIT   SRDebugLevel = 0
	ERR    SRDebugLevel = 1
	WARN   SRDebugLevel = 2
	ALERT  SRDebugLevel = 3
	EMERG  SRDebugLevel = 4
	NOTICE SRDebugLevel = 5
	INFO   SRDebugLevel = 6
	DEBUG  SRDebugLevel = 7
	TRACE  SRDebugLevel = 8
)

type GlobalLogging struct {
	Enable bool
}

type ComponentLogging struct {
	Name  string
	Level SRDebugLevel
}

type Notification struct {
	Type    uint8
	Payload []byte
}
