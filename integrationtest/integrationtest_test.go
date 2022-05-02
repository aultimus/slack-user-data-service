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
	log "github.com/cocoonlife/timber"
	"github.com/google/uuid"
	"github.com/goombaio/namegenerator"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/slack-go/slack"
	"github.com/stretchr/testify/assert"
	"github.com/workos-code-challenge/matthew-ault/bin/util"

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
	ok      bool         `json:"ok"`
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

func generateUpdateEvent(user slack.User) []byte {
	updateEventTemplate := `
	{
		"event": {
			"type": "user_change",
			"user": {
				"id": "%s",
				"name": "%s",
				"deleted": %s,
				"profile": {
					"image_512": "%s",
					"status_emoji": "%s",
					"status_text": "%s"
				},
				"real_name": "%s",
				"tz": "%s"
			}
		},
		"type": "event_callback"
	}
	`
	deleted := "false"
	if user.Deleted {
		deleted = "true"
	}
	s := fmt.Sprintf(updateEventTemplate, user.ID, user.Name, deleted,
		user.Profile.Image512, user.Profile.StatusEmoji, user.Profile.StatusText,
		user.RealName, user.TZ)
	return []byte(s)
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

func generateRandomUser(id string) slack.User {
	if id == "" {
		id = uuid.NewString()
	}

	seed := time.Now().UTC().UnixNano() // TODO: should only do this once
	nameGenerator := namegenerator.NewNameGenerator(seed)
	emojis := []string{":lol:", ":work:", ":smiling:", ":house:"}
	statusTexts := []string{"out eating", "out exercising", "out shopping", "doing programming"}
	timezones := []string{"EST", "PST", "BST", "GMT"}

	name := nameGenerator.Generate()

	randomNum := rand.Intn(2)
	var deleted bool
	if randomNum == 1 {
		deleted = true
	}

	return slack.User{
		ID:       id,
		Name:     name,
		RealName: name + " real",
		Deleted:  deleted,
		TZ:       timezones[rand.Intn(len(timezones))],
		Profile: slack.UserProfile{
			Image512:    "http://imgur.com/" + name + ".png",
			StatusEmoji: emojis[rand.Intn(len(emojis))],
			StatusText:  statusTexts[rand.Intn(len(statusTexts))],
		},
	}
}

func mutateUser(user *slack.User) slack.User {
	return generateRandomUser(user.ID)
}

// This is written as one test as multiple go tests are known to run concurrently
// which can make testing against a single server difficult / unpredictable
func TestIntegration(t *testing.T) {
	flag.Parse()

	fmt.Println("integration test started")
	a := assert.New(t)

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

	// TODO sanity check db connection?

	// run http server to mock slack api
	router := mux.NewRouter()

	// this server will mock the slack api
	server := &http.Server{
		Addr:         ":8081",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	waitChannel := make(chan struct{}, 1)

	// TODO: randomly generate data at runtime
	// TODO: Test with a lot of users - maybe 1000?

	numUsers := 1000
	userResponse := UsersResponse{
		Members: make([]slack.User, numUsers),
	}

	for i := 0; i < numUsers; i++ {
		userResponse.Members[i] = generateRandomUser("")
	}

	// usersListHandler mocks the slack api
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

	// send event to the webhooks endpoint to update user
	resp, err := httpClient.Get("http://app:3000/users")
	if err != nil {
		a.FailNow(err.Error())
	}
	defer resp.Body.Close()

	expected[0].Profile.StatusText = "building"
	expected[0].Profile.StatusEmoji = ":building:"

	b := generateUpdateEvent(expected[0])

	resp, err = httpClient.Post("http://app:3000/webhooks", "application/json", bytes.NewBuffer(b))
	a.Equal(200, resp.StatusCode)
	if err != nil {
		a.FailNow(err.Error())
	}
	time.Sleep(time.Second * 2)

	actual, err = fetchUsers(httpClient)
	if err != nil {
		a.FailNow(err.Error())
	}
	a.Equal(expected, actual)
	//spew.Dump(actual[0])

	// TODO: send event to the webhooks endpoint to update user
	// TODO: request html form, parse html form to check results are correct

	// TODO: send event to the webhooks endpoint to update user
	// TODO: request html form, parse html form to check results are correct
}
