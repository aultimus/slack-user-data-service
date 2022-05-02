package server

import (
	"encoding/json"
	"flag"
	"testing"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/stretchr/testify/assert"
)

/*
ParseEvent
*/

func TestParseEvent(t *testing.T) {
	flag.Parse()

	a := assert.New(t)

	innerEvent := slackevents.EventsAPIEvent{
		Data: slack.User{ID: "foo", Name: "foobar"},
		Type: "user_change",
	}
	b, err := json.Marshal(innerEvent)
	a.NoError(err)
	rawMsg := json.RawMessage(b)
	e := slackevents.EventsAPICallbackEvent{
		Type:       "event_callback",
		InnerEvent: &rawMsg,
	}
	b, err = json.Marshal(e)
	a.NoError(err)

	b = []byte(`
	{
		"event": {
			"type": "user_change",
			"user": {
				"name": "matthew.ault",
				"deleted": false,
				"profile": {
					"image_512": "https://secure.gravatar.com/avatar/d71bef6a8706532e51bed9361a4ff3ce.jpg?s=512&d=https%3A%2F%2Fa.slack-edge.com%2Fdf10d%2Fimg%2Favatars%2Fava_0014-512.png",
					"status_emoji": ":house_with_garden:",
					"status_text": "Working remotely"
				},
				"real_name": "Matthew Ault",
				"tz": "America/New_York"
			}
		},
		"type": "event_callback"
	}
	`)

	ev, err := slackevents.ParseEvent(b, slackevents.OptionNoVerifyToken())
	userChangeEvent, ok := ev.InnerEvent.Data.(*slack.UserChangeEvent)
	a.True(ok)
	//fmt.Println(string(b))
	//spew.Dump(event)
	a.Equal("Matthew Ault", userChangeEvent.User.RealName)

	//userChangeEvent, ok := ev.InnerEvent.Data.(*slack.UserChangeEvent)
	//a.True(ok)
	//fmt.Println(string(b))
	//spew.Dump(event)
	//spew.Dump(ev)
	//spew.Dump(userChangeEvent.User)
}
