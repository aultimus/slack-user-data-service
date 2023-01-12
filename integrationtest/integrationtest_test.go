package integrationtest

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aultimus/slack-user-data-service/util"
	log "github.com/cocoonlife/timber"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"

	_ "github.com/lib/pq"
)

func init() {
	log.AddLogger(log.ConfigLogger{
		LogWriter: new(log.ConsoleWriter),
		Level:     log.DEBUG,
		Formatter: log.NewPatFormatter("[%D %T] [%L] %s %M"),
	})
}

type UsersResponse struct {
	Members []slack.User `json:"members"`
}

func wipeDatabase(dbConn *sqlx.DB) error {
	_, err := dbConn.Exec("DELETE FROM users")
	return err
}

func strToBool(s string) bool {
	if s == "true" {
		return true
	}
	return false
}

// this code is very brittle
func parseHTML(data string) ([]slack.User, error) {
	var out []slack.User

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(data))
	if err != nil {
		return out, err
	}

	var headings, row []string
	var rows [][]string
	firstRow := true

	// Find each table
	doc.Find("table").Each(func(index int, tablehtml *goquery.Selection) {
		tablehtml.Find("tr").Each(func(indextr int, rowhtml *goquery.Selection) {
			rowhtml.Find("th").Each(func(indexth int, tableheading *goquery.Selection) {
				headings = append(headings, tableheading.Text())
			})
			rowhtml.Find("td").Each(func(indexth int, tablecell *goquery.Selection) {
				row = append(row, tablecell.Text())
			})
			rows = append(rows, row)
			if !firstRow { // first row is headers
				out = append(out, slack.User{ID: row[0], Name: row[1],
					Deleted: strToBool(row[2]), RealName: row[3], TZ: row[4],
					Profile: slack.UserProfile{
						StatusText:  row[5],
						StatusEmoji: row[6],
						Image512:    row[7],
					},
				})
			}
			firstRow = false
			row = nil
		})
	})
	// we might want to use heading values at some point in future
	//fmt.Println("####### headings = ", len(headings), headings)
	return out, nil
}

func fetchUsers(httpClient *http.Client) ([]slack.User, error) {
	// request html form, parse html form to check results are correct
	resp, err := httpClient.Get("http://app:3000/users")
	if err != nil {
		return []slack.User{}, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	return parseHTML(string(b))
}

// This is written as one test as multiple go tests are known to run concurrently
// which can make testing against a single server difficult / unpredictable
func TestIntegration(t *testing.T) {
	flag.Parse()

	fmt.Println("integration test started")
	a := assert.New(t)

	token := os.Getenv("SLACK_VERIFICATION_TOKEN")
	if token == "" {
		a.FailNow("SLACK_VERIFICATION_TOKEN must be set")
	}

	// set up db
	dbStr := os.Getenv("DB_CONNECTION_STRING")
	dbConn, err := util.WaitForDB(dbStr)
	if err != nil {
		a.FailNow(err.Error())
	}
	defer dbConn.Close()

	err = dbConn.Ping()
	if err != nil {
		log.Errorf("failed to connect to db:" + err.Error())
	}

	err = wipeDatabase(dbConn)
	if err != nil {
		panic(err)
	}

	defer func() {
		wipeDatabase(dbConn)
		if err != nil {
			log.Error(err)
		}

	}()

	// run http server to mock slack api
	router := mux.NewRouter()
	server := &http.Server{
		Addr:         ":8081",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	waitChannel := make(chan struct{}, 1)

	// generate random user data
	numUsers := 1000
	userResponse := UsersResponse{
		Members: make([]slack.User, numUsers),
	}

	for i := 0; i < numUsers; i++ {
		userResponse.Members[i] = util.GenerateRandomUser("")
	}

	// usersListHandler mocks the slack api - writes random user data
	usersListHandler := func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		b, err := json.Marshal(userResponse)
		if err != nil {
			a.FailNow("failed to marshal response: ", err.Error())
		}
		w.Write(b)
		waitChannel <- struct{}{}
	}

	router.HandleFunc("/users.list", usersListHandler)

	go func(server *http.Server) {
		fmt.Println(server.ListenAndServe().Error())
	}(server)

	// wait until app has hit users.list to intialise users
	select {
	case <-time.After(time.Second * 20):
		// call timed out
		a.FailNow("timed out waiting for app to hit users.list endpoint")
	case <-waitChannel:
		fmt.Println("users.list hit")
	}
	// wait for server to update
	time.Sleep(time.Second * 2)

	// check server has correct information
	httpClient := &http.Client{
		Timeout: 2 * time.Second,
	}

	actual, err := fetchUsers(httpClient)
	if err != nil {
		a.FailNow(err.Error())
	}

	expected := userResponse.Members[:]

	sort.Slice(expected, func(i, j int) bool {
		return expected[i].ID < expected[j].ID
	})

	a.Equal(expected, actual)

	resp, err := httpClient.Get("http://app:3000/users")
	if err != nil {
		a.FailNow(err.Error())
	}
	defer resp.Body.Close()

	// send several user_change events, send them to the webhooks endpoint and
	// check they are reflected on the users page
	for i := 0; i < 25; i++ {
		userIndex := rand.Intn(len(expected))
		//fmt.Println()
		//spew.Dump(expected[userIndex])
		expected[userIndex] = util.GenerateRandomUser(expected[userIndex].ID)
		//spew.Dump(expected[userIndex])
		b := util.GenerateUpdateEvent(expected[userIndex], token)
		resp, err = httpClient.Post("http://app:3000/webhooks", "application/json", bytes.NewBuffer(b))
		a.Equal(200, resp.StatusCode)
		if err != nil {
			a.FailNow(err.Error())
		}
		time.Sleep(time.Millisecond * 150)

		actual, err = fetchUsers(httpClient)
		if err != nil {
			a.FailNow(err.Error())
		}
		a.Equal(expected, actual)
		//spew.Dump(expected[userIndex], actual[userIndex])
	}

	// add a new user via user_change event
	newUser := util.GenerateRandomUser("")
	b := util.GenerateUpdateEvent(newUser, token)
	resp, err = httpClient.Post("http://app:3000/webhooks", "application/json", bytes.NewBuffer(b))
	a.Equal(200, resp.StatusCode)
	if err != nil {
		a.FailNow(err.Error())
	}
	time.Sleep(time.Millisecond * 200)

	expected = append(expected, newUser)
	sort.Slice(expected, func(i, j int) bool {
		return expected[i].ID < expected[j].ID
	})

	actual, err = fetchUsers(httpClient)
	if err != nil {
		a.FailNow(err.Error())
	}
	a.Equal(expected, actual)

	// TODO: test that we ignore event types other than 'user_change'

	// send event without verification token and check it is not processed
	anotherUser := util.GenerateRandomUser("")
	b = util.GenerateUpdateEvent(anotherUser, "foo")
	resp, err = httpClient.Post("http://app:3000/webhooks", "application/json", bytes.NewBuffer(b))
	a.Equal(200, resp.StatusCode)
	if err != nil {
		a.FailNow(err.Error())
	}
	time.Sleep(time.Millisecond * 200)

	actual, err = fetchUsers(httpClient)
	if err != nil {
		a.FailNow(err.Error())
	}
	a.Equal(expected, actual)

}
