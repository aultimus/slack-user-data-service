package integrationtest

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	log "github.com/cocoonlife/timber"
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

// captured via running 'curl -X POST -H "Authorization: Bearer $SLACK_API_TOKEN" https://slack.com/api/users.list'
// TODO: use more interesting users than the ones already in the db
var usersListJSON = `
{
    "cache_ts": 1651434036,
    "members": [
        {
            "color": "757575",
            "deleted": false,
            "id": "user1",
            "is_admin": false,
            "is_app_user": false,
            "is_bot": false,
            "is_email_confirmed": false,
            "is_owner": false,
            "is_primary_owner": false,
            "is_restricted": false,
            "is_ultra_restricted": false,
            "name": "benderbends@gmail.com",
            "profile": {
                "always_active": true,
                "avatar_hash": "sv41d8cd98f0",
                "display_name": "Bender",
                "display_name_normalized": "Bender",
                "fields": null,
                "first_name": "Bender",
                "image_192": "https://a.slack-edge.com/80588/marketing/img/avatars/slackbot/avatar-slackbot.png",
                "image_24": "https://a.slack-edge.com/80588/img/slackbot_24.png",
                "image_32": "https://a.slack-edge.com/80588/img/slackbot_32.png",
                "image_48": "https://a.slack-edge.com/80588/img/slackbot_48.png",
                "image_512": "https://a.slack-edge.com/80588/img/slackbot_512.png",
                "image_72": "https://a.slack-edge.com/80588/img/slackbot_72.png",
                "last_name": "the Robot",
                "phone": "",
                "real_name": "Bender",
                "real_name_normalized": "Bender",
                "skype": "",
                "status_emoji": "",
                "status_emoji_display_info": [],
                "status_expiration": 0,
                "status_text": "",
                "status_text_canonical": "",
                "team": "T03D3SWA4DA",
                "title": ""
            },
            "real_name": "Bender",
            "team_id": "T03D3SWA4DA",
            "tz": "America/Los_Angeles",
            "tz_label": "Pacific Daylight Time",
            "tz_offset": -25200,
            "updated": 0,
            "who_can_share_contact_card": "EVERYONE"
        },
        {
            "color": "9f69e7",
            "deleted": false,
            "has_2fa": false,
            "id": "user3",
            "is_admin": true,
            "is_app_user": false,
            "is_bot": false,
            "is_email_confirmed": true,
            "is_owner": true,
            "is_primary_owner": true,
            "is_restricted": false,
            "is_ultra_restricted": false,
            "name": "rafael the turtle",
            "profile": {
                "avatar_hash": "g423be4b81b8",
                "display_name": "",
                "display_name_normalized": "",
                "email": "ilovepizza@gmail.com",
                "fields": null,
                "first_name": "Rafael",
                "image_192": "https://secure.gravatar.com/avatar/423be4b81b880abaf7a4f5b3ebbe9c13.jpg?s=192&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0023-192.png",
                "image_24": "https://secure.gravatar.com/avatar/423be4b81b880abaf7a4f5b3ebbe9c13.jpg?s=24&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0023-24.png",
                "image_32": "https://secure.gravatar.com/avatar/423be4b81b880abaf7a4f5b3ebbe9c13.jpg?s=32&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0023-32.png",
                "image_48": "https://secure.gravatar.com/avatar/423be4b81b880abaf7a4f5b3ebbe9c13.jpg?s=48&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0023-48.png",
                "image_512": "https://secure.gravatar.com/avatar/423be4b81b880abaf7a4f5b3ebbe9c13.jpg?s=512&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0023-512.png",
                "image_72": "https://secure.gravatar.com/avatar/423be4b81b880abaf7a4f5b3ebbe9c13.jpg?s=72&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0023-72.png",
                "last_name": "the Turtle",
                "phone": "",
                "real_name": "Rafael",
                "real_name_normalized": "Rafael",
                "skype": "",
                "status_emoji": "",
                "status_emoji_display_info": [],
                "status_expiration": 0,
                "status_text": "",
                "status_text_canonical": "",
                "team": "T03D3SWA4DA",
                "title": ""
            },
            "real_name": "WorkOS",
            "team_id": "T03D3SWA4DA",
            "tz": "America/Los_Angeles",
            "tz_label": "Pacific Daylight Time",
            "tz_offset": -25200,
            "updated": 1651002109,
            "who_can_share_contact_card": "EVERYONE"
        },
        {
            "color": "e7392d",
            "deleted": false,
            "has_2fa": false,
            "id": "user2",
            "is_admin": false,
            "is_app_user": false,
            "is_bot": false,
            "is_email_confirmed": true,
            "is_owner": false,
            "is_primary_owner": false,
            "is_restricted": false,
            "is_ultra_restricted": false,
            "name": "joe.bloggs",
            "profile": {
                "avatar_hash": "9d52980a0dca",
                "display_name": "Joe Bloggs",
                "display_name_normalized": "Joe Bloggs",
                "email": "joe.bloggs.backup@gmail.com",
                "fields": null,
                "first_name": "Joe",
                "image_1024": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_1024.png",
                "image_192": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_192.png",
                "image_24": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_24.png",
                "image_32": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_32.png",
                "image_48": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_48.png",
                "image_512": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_512.png",
                "image_72": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_72.png",
                "image_original": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_original.png",
                "is_custom_image": true,
                "last_name": "Bloggs",
                "phone": "",
                "real_name": "Joe Bloggs",
                "real_name_normalized": "Joe Bloggs",
                "skype": "",
                "status_emoji": "",
                "status_emoji_display_info": [],
                "status_expiration": 0,
                "status_text": "",
                "status_text_canonical": "",
                "team": "T03D3SWA4DA",
                "title": ""
            },
            "real_name": "Joe Bloggs",
            "team_id": "T03D3SWA4DA",
            "tz": "America/New_York",
            "tz_label": "Eastern Daylight Time",
            "tz_offset": -14400,
            "updated": 1651183620,
            "who_can_share_contact_card": "EVERYONE"
        },
        {
            "color": "4bbe2e",
            "deleted": false,
            "has_2fa": false,
            "id": "user4",
            "is_admin": true,
            "is_app_user": false,
            "is_bot": false,
            "is_email_confirmed": false,
            "is_owner": false,
            "is_primary_owner": false,
            "is_restricted": false,
            "is_ultra_restricted": false,
            "name": "matthew.ault",
            "profile": {
                "avatar_hash": "gd71bef6a870",
                "display_name": "ash",
                "display_name_normalized": "ash",
                "email": "matthew.ault@gmail.com",
                "fields": null,
                "first_name": "Ash",
                "huddle_state": "default_unset",
                "image_192": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=192&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-192.png",
                "image_24": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=24&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-24.png",
                "image_32": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=32&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-32.png",
                "image_48": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=48&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-48.png",
                "image_512": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=512&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-512.png",
                "image_72": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=72&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-72.png",
                "last_name": "Ketchem",
                "phone": "",
                "real_name": "Ash Ketchem",
                "real_name_normalized": "Ash Ketchem",
                "skype": "",
                "status_emoji": ":100:",
                "status_emoji_display_info": [],
                "status_expiration": 1651463999,
                "status_text": "Working remotely",
                "status_text_canonical": "Working remotely",
                "team": "T03D3SWA4DA",
                "title": ""
            },
            "real_name": "Matthew Ault",
            "team_id": "T03D3SWA4DA",
            "tz": "America/Regina",
            "tz_label": "Central Standard Time",
            "tz_offset": -21600,
            "updated": 1651428990,
            "who_can_share_contact_card": "EVERYONE"
        },
        {
            "color": "3c989f",
            "deleted": false,
            "has_2fa": false,
            "id": "user5",
            "is_admin": false,
            "is_app_user": false,
            "is_bot": false,
            "is_email_confirmed": true,
            "is_owner": false,
            "is_primary_owner": false,
            "is_restricted": false,
            "is_ultra_restricted": false,
            "name": "brandi.a.goulding",
            "profile": {
                "avatar_hash": "79d3a67d6115",
                "display_name": "Bob The Builder",
                "display_name_normalized": "Bob The Builder",
                "email": "brandi.a.goulding@gmail.com",
                "fields": null,
                "first_name": "Bobby",
                "image_1024": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_1024.jpg",
                "image_192": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_192.jpg",
                "image_24": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_24.jpg",
                "image_32": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_32.jpg",
                "image_48": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_48.jpg",
                "image_512": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_512.jpg",
                "image_72": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_72.jpg",
                "image_original": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_original.jpg",
                "is_custom_image": true,
                "last_name": "The Builder",
                "phone": "",
                "real_name": "Bob The Builder",
                "real_name_normalized": "Bobby Builder",
                "skype": "",
                "status_emoji": "",
                "status_emoji_display_info": [],
                "status_expiration": 0,
                "status_text": "",
                "status_text_canonical": "",
                "team": "T03D3SWA4DA",
                "title": ""
            },
            "real_name": "Bob The Builder",
            "team_id": "T03D3SWA4DA",
            "tz": "America/New_York",
            "tz_label": "Eastern Daylight Time",
            "tz_offset": -14400,
            "updated": 1651183952,
            "who_can_share_contact_card": "EVERYONE"
        }
    ],
    "ok": true,
    "response_metadata": {
        "next_cursor": ""
    }
}`

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
	userResponse := UsersResponse{
		Members: []slack.User{
			{ID: "user1", Name: "Bob the Builder", Deleted: false, RealName: "Bob Humphries"},
			{ID: "user3", Name: "Rafael the Ninja Turtle", Deleted: false, RealName: "Rafael Tortuga",
				Profile: slack.UserProfile{StatusText: "eating pizza", StatusEmoji: ":pizza:", Image512: "http://somesite/turtle.png"}},
			{ID: "user2", Name: "Ash Ketchum", Deleted: false, RealName: "Ash Ketchum"},
			{ID: "user4", Name: "Dora the Deleted", Deleted: true, RealName: "Dora Dorries"},
			{ID: "user5", Name: "Jane", Deleted: false, RealName: "Jane Adams"},
		},
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
	time.Sleep(time.Second * 5)

	// check server has correct information
	httpClient := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := httpClient.Get("http://app:3000/users")
	if err != nil {
		a.FailNow(err.Error())
	}
	defer resp.Body.Close()

	expected := userResponse.Members[:]

	sort.Slice(expected, func(i, j int) bool {
		return expected[i].ID < expected[j].ID
	})
	// this code relies upon column ordering but we could parse headers out

	b, err := io.ReadAll(resp.Body)
	fmt.Println(string(b))
	actual, err := parseHTML(string(b))
	if err != nil {
		a.FailNow(err.Error())
	}

	a.Equal(expected, actual)

	// TODO: send event to the webhooks endpoint to update user
	// TODO: request html form, parse html form to check results are correct

	// TODO: send event to the webhooks endpoint to update user
	// TODO: request html form, parse html form to check results are correct

	// TODO: send event to the webhooks endpoint to update user
	// TODO: request html form, parse html form to check results are correct
}
