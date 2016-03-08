package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"infra/sysd/rpc"
	"infra/sysd/server"
	"utils/logging"
)

func main() {
	fmt.Println("Starting system daemon")
	paramsDir := flag.String("params", "./params", "Params directory")
	flag.Parse()

	dbName := *paramsDir + "/UsrConfDb.db"
	fmt.Println("Opening Config DB: ", dbName)
	dbHdl, err := sql.Open("sqlite3", dbName)
	if err != nil {
		fmt.Println("Failed to open connection to DB. ", err, " Exiting!!")
		return
	}

	logger, err := logging.NewLogger(*paramsDir, "sysd", "SYSTEM", dbHdl)
	if err != nil {
		fmt.Println("Failed to start the logger. Exiting!!")
		return
	}
	logger.Info("Started the logger successfully.")

	fileName := *paramsDir
	if fileName[len(fileName)-1] != '/' {
		fileName = fileName + "/"
	}
	fileName = fileName + "clients.json"

	logger.Info(fmt.Sprintln("Starting Sysd Server..."))
	sysdServer := server.NewSYSDServer(logger)
	go sysdServer.StartServer(fileName, dbHdl)

	logger.Info(fmt.Sprintln("Starting Config listener..."))
	confIface := rpc.NewSYSDHandler(logger, sysdServer)
	rpc.StartServer(logger, confIface, fileName)
}
