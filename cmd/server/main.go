package main

import (
	"math/rand"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/slack-go/slack"

	"flag"

	"github.com/cocoonlife/timber"
	log "github.com/cocoonlife/timber"
	"github.com/workos-code-challenge/matthew-ault/bin/util"
	"github.com/workos-code-challenge/matthew-ault/db"
	"github.com/workos-code-challenge/matthew-ault/server"

	"net/http"
	_ "net/http/pprof"

	_ "github.com/lib/pq"
)

func init() {
	log.AddLogger(log.ConfigLogger{
		LogWriter: new(log.ConsoleWriter),
		Level:     log.DEBUG,
		Formatter: log.NewPatFormatter("[%D %T] [%L] %s %M"),
	})
	rand.Seed(time.Now().UnixNano())
}

func main() {
	log.Infof("server started")

	var portNum = *flag.String("port", server.DefaultPortNum,
		"TCP/IP port that this program listens on")
	flag.Parse()
	if portNum == "" {
		portNum = server.DefaultPortNum
	}

	slackAPIToken := os.Getenv("SLACK_API_TOKEN")
	if slackAPIToken == "" {
		log.Fatal("SLACK_API_TOKEN env var not set")
	}

	slackAPIURL := os.Getenv("SLACK_API_URL")
	if slackAPIURL == "" {
		slackAPIURL = slack.APIURL
	}

	// set up database
	dbStr := os.Getenv("DB_CONNECTION_STRING")
	dbConn, err := util.WaitForDB(dbStr)
	if err != nil {
		timber.Fatal("failed to connect to database" + err.Error())
	}
	defer dbConn.Close()

	err = dbConn.Ping()
	if err != nil {
		timber.Fatal(err)
	}
	postgres := db.NewPostgres(dbConn)

	// pprof - see: http://localhost:6060/debug/pprof/
	go func() {
		log.Errorf(http.ListenAndServe(":6060", nil))
	}()

	log.Infof("using slack api url %s", slackAPIURL)
	// TODO: make debug toggleable when starting server
	slackClient := slack.New(slackAPIToken, slack.OptionDebug(true))
	slack.OptionAPIURL(slackAPIURL)(slackClient)
	// set up app
	app := server.NewApp()

	err = app.Init(portNum, postgres, slackClient)
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = app.Run()
	if err != nil {
		log.Fatalf(err.Error())
	}
}
