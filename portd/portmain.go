package main 
import ("os"
		  "log"
		"portdServices"
        "git.apache.org/thrift.git/lib/go/thrift"
        )

var logger *log.Logger

func main () {
    var transport thrift.TServerTransport
    var err error
	 var addr = "localhost:9090"

    logger = log.New(os.Stdout, "PORTD :", log.Ldate|log.Ltime|log.Lshortfile)                                

    transport, err = thrift.NewTServerSocket(addr)
	 if err != nil {
		  logger.Println("Failed to create Socket with:", addr)
	 }
    handler := NewPortServiceHandler()
    processor := portdServices.NewPortServiceProcessor(handler)
    transportFactory := thrift.NewTBufferedTransportFactory(8192) 
    protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
    server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
    logger.Println("Starting PORT daemon")
    server.Serve()
}
