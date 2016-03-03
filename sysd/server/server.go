package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	nanomsg "github.com/op/go-nanomsg"
	"infra/sysd/sysdCommonDefs"
	"os"
	"os/signal"
	"syscall"
	"utils/logging"
)

type GlobalLoggingConfig struct {
	Enable bool
}

type ComponentLoggingConfig struct {
	Component string
	Level     logging.SRDebugLevel
}

type SYSDServer struct {
	logger                   *logging.Writer
	GlobalLoggingConfigCh    chan GlobalLoggingConfig
	ComponentLoggingConfigCh chan ComponentLoggingConfig
	sysdPubSocket            *nanomsg.PubSocket
	notificationCh           chan []byte
}

func NewSYSDServer(logger *logging.Writer) *SYSDServer {
	sysdServer := &SYSDServer{}
	sysdServer.logger = logger
	sysdServer.GlobalLoggingConfigCh = make(chan GlobalLoggingConfig)
	sysdServer.ComponentLoggingConfigCh = make(chan ComponentLoggingConfig)
	sysdServer.notificationCh = make(chan []byte)
	return sysdServer
}

func (server *SYSDServer) SigHandler() {
	server.logger.Info(fmt.Sprintln("Starting SigHandler"))
	sigChan := make(chan os.Signal, 1)
	signalList := []os.Signal{syscall.SIGHUP}
	signal.Notify(sigChan, signalList...)

	for {
		select {
		case signal := <-sigChan:
			switch signal {
			case syscall.SIGHUP:
				server.logger.Info("Received SIGHUP signal. Exiting")
				os.Exit(0)
			default:
				server.logger.Info(fmt.Sprintln("Unhandled signal : ", signal))
			}
		}
	}
}

func (server *SYSDServer) InitServer(paramFile string) {
	server.logger.Info(fmt.Sprintln("Starting Sysd Server"))
}

func (server *SYSDServer) InitPublisher(pub_str string) (pub *nanomsg.PubSocket) {
	server.logger.Info(fmt.Sprintln("Setting up ", pub_str, "publisher"))
	pub, err := nanomsg.NewPubSocket()
	if err != nil {
		server.logger.Info(fmt.Sprintln("Failed to open pub socket"))
		return nil
	}
	ep, err := pub.Bind(pub_str)
	if err != nil {
		server.logger.Info(fmt.Sprintln("Failed to bind pub socket - ", ep))
		return nil
	}
	err = pub.SetSendBuffer(1024)
	if err != nil {
		server.logger.Info(fmt.Sprintln("Failed to set send buffer size"))
		return nil
	}
	return pub
}

func (server *SYSDServer) PublishSysdNotifications() {
	server.sysdPubSocket = server.InitPublisher(sysdCommonDefs.PUB_SOCKET_ADDR)
	for {
		select {
		case event := <-server.notificationCh:
			server.logger.Info(fmt.Sprintln("Received call to notify session state", event))
			_, err := server.sysdPubSocket.Send(event, nanomsg.DontWait)
			if err == syscall.EAGAIN {
				server.logger.Err(fmt.Sprintln("Failed to publish event"))
			}
		}
	}
}

func (server *SYSDServer) ReadGlobalLoggingConfigFromDB(dbHdl *sql.DB) error {
	return nil
}

func (server *SYSDServer) ReadComponentLoggingConfigFromDB(dbHdl *sql.DB) error {
	return nil
}

func (server *SYSDServer) ReadConfigFromDB(dbHdl *sql.DB) error {
	var err error
	defer dbHdl.Close()
	// BfdGlobalConfig
	err = server.ReadGlobalLoggingConfigFromDB(dbHdl)
	if err != nil {
		return err
	}
	// BfdIntfConfig
	err = server.ReadComponentLoggingConfigFromDB(dbHdl)
	if err != nil {
		return err
	}
	return nil
}

func (server *SYSDServer) ProcessGlobalLoggingConfig(gLogConf GlobalLoggingConfig) error {
	msg := sysdCommonDefs.GlobalLogging{
		Enable: gLogConf.Enable,
	}
	msgBuf, err := json.Marshal(msg)
	if err != nil {
		server.logger.Err("Failed to marshal Global logging message")
		return err
	}
	notification := sysdCommonDefs.Notification{
		Type:    uint8(sysdCommonDefs.G_LOG),
		Payload: msgBuf,
	}
	notificationBuf, err := json.Marshal(notification)
	if err != nil {
		server.logger.Err("Failed to marshal Global logging message")
		return err
	}
	server.notificationCh <- notificationBuf
	return nil
}

func (server *SYSDServer) ProcessComponentLoggingConfig(cLogConf ComponentLoggingConfig) error {
	msg := sysdCommonDefs.ComponentLogging{
		Name:  cLogConf.Component,
		Level: cLogConf.Level,
	}
	msgBuf, err := json.Marshal(msg)
	if err != nil {
		server.logger.Err("Failed to marshal Global logging message")
		return err
	}
	notification := sysdCommonDefs.Notification{
		Type:    uint8(sysdCommonDefs.C_LOG),
		Payload: msgBuf,
	}
	notificationBuf, err := json.Marshal(notification)
	if err != nil {
		server.logger.Err("Failed to marshal Global logging message")
		return err
	}
	server.notificationCh <- notificationBuf
	return nil
}

func (server *SYSDServer) StartServer(paramFile string, dbHdl *sql.DB) {
	// Start signal handler first
	go server.SigHandler()
	// Initialize BFD server from params file
	server.InitServer(paramFile)
	// Start notification publish thread
	go server.PublishSysdNotifications()
	// Read BFD configurations already present in DB
	go server.ReadConfigFromDB(dbHdl)

	// Now, wait on below channels to process
	for {
		select {
		case gLogConf := <-server.GlobalLoggingConfigCh:
			server.logger.Info(fmt.Sprintln("Received call for performing Global logging Configuration", gLogConf))
			server.ProcessGlobalLoggingConfig(gLogConf)
		case compLogConf := <-server.ComponentLoggingConfigCh:
			server.logger.Info(fmt.Sprintln("Received call for performing Component logging Configuration", compLogConf))
			server.ProcessComponentLoggingConfig(compLogConf)
		}
	}
}
