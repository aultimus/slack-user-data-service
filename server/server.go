package server

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"text/template"
	"time"

	log "github.com/cocoonlife/timber"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/workos-code-challenge/matthew-ault/models"
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
	CreateUser(user models.User) error
	CreateUsers(user []models.User) error
	UpdateUser(user models.User) error
	GetAllUsers() ([]models.User, error)
}

type App struct {
	server      *http.Server
	db          Storer
	slackClient *slack.Client
}

func (a *App) Init(portNum string, storer Storer, slackAPIToken string) error {
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

	router.HandleFunc("/", a.RootHandler)
	router.HandleFunc("/users", a.UsersHandler).Methods(http.MethodGet)
	router.HandleFunc("/webhooks", a.WebhooksHandler)

	a.server = server
	a.db = storer

	// TODO: make debug toggleable when starting server
	a.slackClient = slack.New(slackAPIToken, slack.OptionDebug(true))
	a.FetchUsers()

	return nil
}

// TODO: do something with long running requests and use context

func (a *App) Run() error {
	log.Infof("running server on %s", a.server.Addr)
	return a.server.ListenAndServe()
}

func (a *App) RootHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Hello world today!")
}

func (a *App) UsersHandler(w http.ResponseWriter, req *http.Request) {
	users, err := a.db.GetAllUsers()
	if err != nil {
		log.Errorf("db GetAllUsers returned error: %v", err)
		return
	}
	usersStruct := struct {
		Users []models.User
	}{Users: users}

	// it may make sense to have a user facing User struct definition at some point
	tmpl, _ := template.ParseFiles("./html/users.html")
	tmpl.Execute(w, usersStruct)
}

func (a *App) WebhooksHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	b, err := io.ReadAll(req.Body)
	if err != nil {
		log.Errorf("failed to parse slack event body: %v", err)
		return
	}

	//var eventObj models.Event
	//err = json.Unmarshal(b, &eventObj)
	//if err != nil {
	//	log.Errorf("error unmarshalling json: %v", err)
	//	return
	//}
	// TODO: use verification
	event, err := slackevents.ParseEvent(b, slackevents.OptionNoVerifyToken())
	if err != nil {
		log.Errorf("failed slackevents.ParseEvent: %v", err)
		return
	}
	log.Debugf("received %s type event", event.InnerEvent.Type)
	switch event.InnerEvent.Type {
	// Note: go falls through by default
	case "user_change":
		// https://api.slack.com/events/user_change
		log.Debugf("processing %s event", event.InnerEvent.Type)
		switch event.InnerEvent.Data.(type) {
		case *slack.UserChangeEvent:
			//fmt.Println(string(b))
			//spew.Dump(event)
			apiUser := event.InnerEvent.Data.(*slack.UserChangeEvent).User
			dbUser := APIToDBUser(apiUser)
			err = a.db.UpdateUser(dbUser)
			if err != nil {
				log.Errorf("error during UpdateUser: %s, user: %s", err.Error(), spew.Sdump(dbUser))
				return
			}
		default: // something went horribly wrong
			log.Errorf("user_change event has inner data of type %v ", reflect.TypeOf(event.InnerEvent.Data))
		}
	default: // unregonised event type
		log.Debugf("ignoring event of event type %s", event.InnerEvent.Type)
	}
}

func APIToDBUser(in slack.User) models.User {
	// we could either implement this function via marshalling and unmarshalling
	// or via mapping. marshalling and unmarshalling is more extensible
	// but less can go wrong with a mapping function like this
	return models.User{
		Deleted:            in.Deleted,
		ID:                 in.ID,
		Name:               in.Name,
		ProfileImage512:    in.Profile.Image512,
		ProfileStatusEmoji: in.Profile.StatusEmoji,
		ProfileStatusText:  in.Profile.StatusText,
		RealName:           in.RealName,
		Tz:                 in.TZ,
	}
}

func APIToDBUsers(in []slack.User) []models.User {
	out := make([]models.User, len(in))
	for i := 0; i < len(in); i++ {
		out[i] = APIToDBUser(in[i])
	}
	return out
}

// FetchUsers retrieves the initial set of users used to initialise the database
func (a *App) FetchUsers() {
	// TODO: should really wrap errors in here and return them to highest callsite
	// TODO: we should use GetUsersPaginated for scalability
	users, err := a.slackClient.GetUsers()
	if err != nil {
		log.Errorf("failed api call to slack GetUsers: %v", err)
		return
	}
	log.Infof("retrieved %d users from GetUsers API", len(users))

	dbUsers := APIToDBUsers(users)
	err = a.db.CreateUsers(dbUsers)
	if err != nil {
		log.Errorf("failed db CreateUsers call: %v", err)
		return
	}
	// TODO: more detailed log message
	log.Infof("wrote %d users to DB", len(dbUsers))
}
