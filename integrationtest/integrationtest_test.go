package integrationtest

import (
	"flag"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

// captured via running 'curl -X POST -H "Authorization: Bearer $SLACK_API_TOKEN" https://slack.com/api/users.list'
// TODO: use more interesting users than the ones already in the db
var usersListJSON = `
{
    "cache_ts": 1651434036,
    "members": [
        {
            "color": "757575",
            "deleted": false,
            "id": "USLACKBOT",
            "is_admin": false,
            "is_app_user": false,
            "is_bot": false,
            "is_email_confirmed": false,
            "is_owner": false,
            "is_primary_owner": false,
            "is_restricted": false,
            "is_ultra_restricted": false,
            "name": "slackbot",
            "profile": {
                "always_active": true,
                "avatar_hash": "sv41d8cd98f0",
                "display_name": "Slackbot",
                "display_name_normalized": "Slackbot",
                "fields": null,
                "first_name": "slackbot",
                "image_192": "https://a.slack-edge.com/80588/marketing/img/avatars/slackbot/avatar-slackbot.png",
                "image_24": "https://a.slack-edge.com/80588/img/slackbot_24.png",
                "image_32": "https://a.slack-edge.com/80588/img/slackbot_32.png",
                "image_48": "https://a.slack-edge.com/80588/img/slackbot_48.png",
                "image_512": "https://a.slack-edge.com/80588/img/slackbot_512.png",
                "image_72": "https://a.slack-edge.com/80588/img/slackbot_72.png",
                "last_name": "",
                "phone": "",
                "real_name": "Slackbot",
                "real_name_normalized": "Slackbot",
                "skype": "",
                "status_emoji": "",
                "status_emoji_display_info": [],
                "status_expiration": 0,
                "status_text": "",
                "status_text_canonical": "",
                "team": "T03D3SWA4DA",
                "title": ""
            },
            "real_name": "Slackbot",
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
            "id": "U03DE5PR65P",
            "is_admin": true,
            "is_app_user": false,
            "is_bot": false,
            "is_email_confirmed": true,
            "is_owner": true,
            "is_primary_owner": true,
            "is_restricted": false,
            "is_ultra_restricted": false,
            "name": "challenge-onsite",
            "profile": {
                "avatar_hash": "g423be4b81b8",
                "display_name": "",
                "display_name_normalized": "",
                "email": "challenge-onsite@workos.com",
                "fields": null,
                "first_name": "WorkOS",
                "image_192": "https://secure.gravatar.com/avatar/423be4b81b880abaf7a4f5b3ebbe9c13.jpg?s=192&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0023-192.png",
                "image_24": "https://secure.gravatar.com/avatar/423be4b81b880abaf7a4f5b3ebbe9c13.jpg?s=24&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0023-24.png",
                "image_32": "https://secure.gravatar.com/avatar/423be4b81b880abaf7a4f5b3ebbe9c13.jpg?s=32&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0023-32.png",
                "image_48": "https://secure.gravatar.com/avatar/423be4b81b880abaf7a4f5b3ebbe9c13.jpg?s=48&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0023-48.png",
                "image_512": "https://secure.gravatar.com/avatar/423be4b81b880abaf7a4f5b3ebbe9c13.jpg?s=512&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0023-512.png",
                "image_72": "https://secure.gravatar.com/avatar/423be4b81b880abaf7a4f5b3ebbe9c13.jpg?s=72&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0023-72.png",
                "last_name": "",
                "phone": "",
                "real_name": "WorkOS",
                "real_name_normalized": "WorkOS",
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
            "id": "U03DEQ55EHZ",
            "is_admin": false,
            "is_app_user": false,
            "is_bot": false,
            "is_email_confirmed": true,
            "is_owner": false,
            "is_primary_owner": false,
            "is_restricted": false,
            "is_ultra_restricted": false,
            "name": "matthew.ault.backup",
            "profile": {
                "avatar_hash": "9d52980a0dca",
                "display_name": "Matthew Ault",
                "display_name_normalized": "Matthew Ault",
                "email": "matthew.ault.backup@gmail.com",
                "fields": null,
                "first_name": "Matthew",
                "image_1024": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_1024.png",
                "image_192": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_192.png",
                "image_24": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_24.png",
                "image_32": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_32.png",
                "image_48": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_48.png",
                "image_512": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_512.png",
                "image_72": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_72.png",
                "image_original": "https://avatars.slack-edge.com/2022-04-28/3480574803088_9d52980a0dcaeb2176e4_original.png",
                "is_custom_image": true,
                "last_name": "Ault",
                "phone": "",
                "real_name": "Matthew Ault",
                "real_name_normalized": "Matthew Ault",
                "skype": "",
                "status_emoji": "",
                "status_emoji_display_info": [],
                "status_expiration": 0,
                "status_text": "",
                "status_text_canonical": "",
                "team": "T03D3SWA4DA",
                "title": ""
            },
            "real_name": "Matthew Ault",
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
            "id": "U03DR9SNFU0",
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
                "display_name": "matt",
                "display_name_normalized": "matt",
                "email": "matthew.ault@gmail.com",
                "fields": null,
                "first_name": "Matthew",
                "huddle_state": "default_unset",
                "image_192": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=192&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-192.png",
                "image_24": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=24&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-24.png",
                "image_32": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=32&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-32.png",
                "image_48": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=48&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-48.png",
                "image_512": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=512&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-512.png",
                "image_72": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=72&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-72.png",
                "last_name": "Ault",
                "phone": "",
                "real_name": "Matthew Ault",
                "real_name_normalized": "Matthew Ault",
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
            "id": "U03E4HE66HE",
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
                "display_name": "Brandi Goulding",
                "display_name_normalized": "Brandi Goulding",
                "email": "brandi.a.goulding@gmail.com",
                "fields": null,
                "first_name": "Brandi",
                "image_1024": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_1024.jpg",
                "image_192": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_192.jpg",
                "image_24": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_24.jpg",
                "image_32": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_32.jpg",
                "image_48": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_48.jpg",
                "image_512": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_512.jpg",
                "image_72": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_72.jpg",
                "image_original": "https://avatars.slack-edge.com/2022-04-28/3450197819382_79d3a67d61151da98221_original.jpg",
                "is_custom_image": true,
                "last_name": "Goulding",
                "phone": "",
                "real_name": "Brandi Goulding",
                "real_name_normalized": "Brandi Goulding",
                "skype": "",
                "status_emoji": "",
                "status_emoji_display_info": [],
                "status_expiration": 0,
                "status_text": "",
                "status_text_canonical": "",
                "team": "T03D3SWA4DA",
                "title": ""
            },
            "real_name": "Brandi Goulding",
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

// This is written as one test as multiple go tests are known to run concurrently
// which can make testing against a single server difficult / unpredictable
func TestIntegration(t *testing.T) {
	flag.Parse()

	fmt.Println("integration test started")
	a := assert.New(t)
	a.True(true)

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

	router.HandleFunc("/users.list", usersListHandler)
	go func(server *http.Server) {
		fmt.Println(server.ListenAndServe().Error())
	}(server)

	// wait until app has hit users.list to intialise users
	time.Sleep(time.Hour)

	// TODO: send event to the webhooks endpoint to update user
	// TODO: request html form, parse html form to check results are correct

	// TODO: send event to the webhooks endpoint to update user
	// TODO: request html form, parse html form to check results are correct

	// TODO: send event to the webhooks endpoint to update user
	// TODO: request html form, parse html form to check results are correct
}

// usersListHandler mocks the slack api
func usersListHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(usersListJSON))
}
