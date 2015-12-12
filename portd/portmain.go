package main

import (
	"flag"
	"git.apache.org/thrift.git/lib/go/thrift"
	"log"
	"log/syslog"
	"os"
	"portdServices"
)

var logger *log.Logger

func main() {
	var transport thrift.TServerTransport
	var err error
	var addr = "localhost:5050"

	logger = log.New(os.Stdout, "PORTD :", log.Ldate|log.Ltime|log.Lshortfile)

	syslogger, err := syslog.New(syslog.LOG_NOTICE|syslog.LOG_INFO|syslog.LOG_DAEMON, "PORTD")
	if err == nil {
		syslogger.Info("### PORT Daemon started")
		logger.SetOutput(syslogger)
	}

	transport, err = thrift.NewTServerSocket(addr)
	if err != nil {
		logger.Println("Failed to create Socket with:", addr)
	}
	paramsDir := flag.String("params", "", "Directory Location for config files")
	flag.Parse()
	handler := NewPortServiceHandler(*paramsDir)
	processor := portdServices.NewPortServiceProcessor(handler)
	transportFactory := thrift.NewTBufferedTransportFactory(8192)
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	logger.Println("Starting PORT daemon")
	server.Serve()
}
