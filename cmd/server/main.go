package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/go-playground/errors"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/slack-go/slack"

	"flag"

	"github.com/cocoonlife/timber"
	log "github.com/cocoonlife/timber"
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

func WaitForDB(dbConnStr string) (*sqlx.DB, error) {
	maxRetries := 10
	var handle *sqlx.DB
	var err error

	for i := 0; i < maxRetries; i++ {
		handle, err = sqlx.Connect("postgres", dbConnStr)
		if err != nil {
			if i < maxRetries-1 {
				log.Infof("sleeping for db connect")
				time.Sleep(time.Second)
				continue
			}
			return nil, err
		}
		break
	}
	log.Infof("connected to db")

	// if db is spinning up, let's be fault tolerant and try a few times
	for i := 0; i < maxRetries; i++ {
		if err = handle.Ping(); err != nil {
			if i < maxRetries-1 {
				log.Infof("sleeping for db ping")
				time.Sleep(time.Second)
				continue
			}
			return nil, err
		}
		break
	}
	log.Infof("pinged db")
	return handle, nil
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
	dbConn, err := WaitForDB(dbStr)
	if err != nil {
		timber.Fatal(errors.Wrap(err, "failed to connect to database").AddTag("connection_string", dbStr))
	}
	defer dbConn.Close()

	err = dbConn.Ping()
	if err != nil {
		timber.Fatal(errors.Wrap(err, "failed to connect to db"))
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
