package main

import (
	"flag"

	log "github.com/cocoonlife/timber"
	"github.com/workos-code-challenge/matthew-ault/server"

	"net/http"
	_ "net/http/pprof"
)

func init() {
	log.AddLogger(log.ConfigLogger{
		LogWriter: new(log.ConsoleWriter),
		Level:     log.DEBUG,
		Formatter: log.NewPatFormatter("[%D %T] [%L] %s %M"),
	})

}

func main() {
	log.Infof("server started")

	var portNum = *flag.String("port", server.DefaultPortNum,
		"TCP/IP port that this program listens on")
	if portNum == "" {
		portNum = server.DefaultPortNum
	}

	flag.Parse()

	// pprof - see: http://localhost:6060/debug/pprof/
	go func() {
		log.Errorf(http.ListenAndServe(":6060", nil))
	}()

	app := server.NewApp()
	err := app.Init(portNum)
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = app.Run()
	if err != nil {
		log.Fatalf(err.Error())
	}
}
