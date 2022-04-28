package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/cocoonlife/timber"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
)

const (
	ContentType    = "Content-Type"
	DefaultPortNum = "3000"
	MimeTypeJSON   = "application/json"
)

func NewApp() *App {
	return &App{}
}

type Storer interface {
}

type App struct {
	server *http.Server
	db     Storer
}

func (a *App) Init(portNum string, storer Storer) error {
	log.Infof("init")
	router := mux.NewRouter()

	server := &http.Server{
		Addr:           ":" + portNum,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	router.HandleFunc("/", a.Hello)
	router.HandleFunc("/webhooks", a.Webhooks)

	a.server = server
	a.db = storer
	return nil
}

func (a *App) Run() error {
	log.Infof("running server on %s", a.server.Addr)
	return a.server.ListenAndServe()
}

func (a *App) Hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Hello world today!")
}

// Event represents a slack event data type which is received by the webhooks
// endpoint. This struct was initally generated using https://mholt.github.io/json-to-go/
// id, name, deleted, real_name, tz, profile object (status_text, status_emoji, image_512).

type Event struct {
	APIAppID       string `json:"api_app_id"`
	Authorizations []struct {
		EnterpriseID        interface{} `json:"enterprise_id"`
		IsBot               bool        `json:"is_bot"`
		IsEnterpriseInstall bool        `json:"is_enterprise_install"`
		TeamID              string      `json:"team_id"`
		UserID              string      `json:"user_id"`
	} `json:"authorizations"`
	Event struct {
		CacheTs int    `json:"cache_ts"`
		EventTs string `json:"event_ts"`
		Type    string `json:"type"`
		User    struct {
			Color             string `json:"color"`
			Deleted           bool   `json:"deleted" db:"deleted"`
			ID                string `json:"id" db:"id"`
			IsAdmin           bool   `json:"is_admin"`
			IsAppUser         bool   `json:"is_app_user"`
			IsBot             bool   `json:"is_bot"`
			IsEmailConfirmed  bool   `json:"is_email_confirmed"`
			IsOwner           bool   `json:"is_owner"`
			IsPrimaryOwner    bool   `json:"is_primary_owner"`
			IsRestricted      bool   `json:"is_restricted"`
			IsUltraRestricted bool   `json:"is_ultra_restricted"`
			Locale            string `json:"locale"`
			Name              string `json:"name" db:"name"`
			Profile           struct {
				AvatarHash             string        `json:"avatar_hash"`
				DisplayName            string        `json:"display_name"`
				DisplayNameNormalized  string        `json:"display_name_normalized"`
				Email                  string        `json:"email"`
				Fields                 interface{}   `json:"fields"`
				FirstName              string        `json:"first_name"`
				Image192               string        `json:"image_192"`
				Image24                string        `json:"image_24"`
				Image32                string        `json:"image_32"`
				Image48                string        `json:"image_48"`
				Image512               string        `json:"image_512" db:"image_512"`
				Image72                string        `json:"image_72"`
				LastName               string        `json:"last_name"`
				Phone                  string        `json:"phone"`
				RealName               string        `json:"real_name"`
				RealNameNormalized     string        `json:"real_name_normalized"`
				Skype                  string        `json:"skype"`
				StatusEmoji            string        `json:"status_emoji" db:"profile_status_emoji"`
				StatusEmojiDisplayInfo []interface{} `json:"status_emoji_display_info"`
				StatusExpiration       int           `json:"status_expiration"`
				StatusText             string        `json:"status_text" db:"profile_status_text"`
				StatusTextCanonical    string        `json:"status_text_canonical"`
				Team                   string        `json:"team"`
				Title                  string        `json:"title"`
			} `json:"profile"`
			RealName               string `json:"real_name" db:"real_name"`
			TeamID                 string `json:"team_id"`
			Tz                     string `json:"tz" db:"tz"`
			TzLabel                string `json:"tz_label"`
			TzOffset               int    `json:"tz_offset"`
			Updated                int    `json:"updated"`
			WhoCanShareContactCard string `json:"who_can_share_contact_card"`
		} `json:"user"`
	} `json:"event"`
	EventID            string `json:"event_id"`
	EventTime          int    `json:"event_time"`
	IsExtSharedChannel bool   `json:"is_ext_shared_channel"`
	TeamID             string `json:"team_id"`
	Token              string `json:"token"`
	Type               string `json:"type"`
}

func (a *App) Webhooks(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	b, err := io.ReadAll(req.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		log.Fatalln(err)
	}
	var eventObj Event
	err = json.Unmarshal(b, &eventObj)
	if err != nil {
		log.Errorf("error unmarshalling json: %v", err)
		return
	}

	spew.Dump(eventObj)
}
