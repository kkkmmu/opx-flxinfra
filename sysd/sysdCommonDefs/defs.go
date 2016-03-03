package sysdCommonDefs

import (
	"utils/logging"
)

const (
	PUB_SOCKET_ADDR = "ipc:///tmp/sysd.ipc"
)

const (
	G_LOG uint8 = 1
	C_LOG uint8 = 2
)

type GlobalLogging struct {
	Enable bool
}

type ComponentLogging struct {
	Name  string
	Level logging.SRDebugLevel
}

type Notification struct {
	Type    uint8
	Payload []byte
}
