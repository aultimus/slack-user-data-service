package util

import (
	"fmt"
	"math/rand"
	"time"

	log "github.com/cocoonlife/timber"
	"github.com/google/uuid"
	"github.com/goombaio/namegenerator"
	"github.com/jmoiron/sqlx"
	"github.com/slack-go/slack"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
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

func GenerateUpdateEvent(user slack.User, token string) []byte {
	updateEventTemplate := `
	{
		"token": "%s",
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
	s := fmt.Sprintf(updateEventTemplate, token, user.ID, user.Name, deleted,
		user.Profile.Image512, user.Profile.StatusEmoji, user.Profile.StatusText,
		user.RealName, user.TZ)
	return []byte(s)
}

func GenerateRandomUser(id string) slack.User {
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

func MutateUser(user *slack.User) slack.User {
	return GenerateRandomUser(user.ID)
}
