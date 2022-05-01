package server

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"reflect"
	"text/template"
	"time"

	log "github.com/cocoonlife/timber"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/workos-code-challenge/matthew-ault/db"
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
	CreateUser(user db.User) error
	CreateUsers(user []db.User) error
	UpdateUser(user db.User) error
	GetAllUsers() ([]db.User, error)
}

type Slacker interface {
	GetUsersContext(ctx context.Context) ([]slack.User, error)
}

type App struct {
	server      *http.Server
	db          Storer
	slackClient Slacker
}

// Init initialises the application server, call before Run
func (a *App) Init(portNum string, storer Storer, slackClient Slacker) error {
	log.Infof("init")
	router := mux.NewRouter()

	server := &http.Server{
		Addr:         ":" + portNum,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	router.HandleFunc("/", a.RootHandler)
	router.HandleFunc("/users", a.UsersHandler).Methods(http.MethodGet)
	router.HandleFunc("/webhooks", a.WebhooksHandler).Methods(http.MethodPost)

	a.server = server
	a.db = storer
	a.slackClient = slackClient

	// run asynchronously so we can still serve requests if api is down
	go a.FetchUsersLoop()

	return nil
}

// TODO: do something with long running requests and use context

// Run starts the application server, call Init first
func (a *App) Run() error {
	log.Infof("running server on %s", a.server.Addr)
	return a.server.ListenAndServe()
}

func (a *App) RootHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Hello world today!")
}

// UsersHandler on request renders a html table of users stored by this service
func (a *App) UsersHandler(w http.ResponseWriter, req *http.Request) {
	users, err := a.db.GetAllUsers()
	if err != nil {
		log.Errorf("db GetAllUsers returned error: %v", err)
		return
	}
	usersStruct := struct {
		Users []db.User
	}{Users: users}

	tmpl, _ := template.ParseFiles("./html/users.html")
	tmpl.Execute(w, usersStruct)
}

// WebhooksHandler processes events from the slack events api
// https://api.slack.com/apis/connections/events-api
func (a *App) WebhooksHandler(w http.ResponseWriter, req *http.Request) {
	b, err := io.ReadAll(req.Body)
	if err != nil {
		log.Errorf("failed to parse slack event body: %v", err)
		return
	}
	defer req.Body.Close()

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
		userChangeEvent, ok := event.InnerEvent.Data.(*slack.UserChangeEvent)
		if !ok {
			log.Errorf("user_change event has inner data of type %v ", reflect.TypeOf(event.InnerEvent.Data))
			return
		}
		//fmt.Println(string(b))
		//spew.Dump(event)
		apiUser := userChangeEvent.User
		dbUser := APIToDBUser(apiUser)
		err = a.db.UpdateUser(dbUser)
		if err != nil {
			log.Errorf("error during UpdateUser: %s, user: %s", err.Error(), spew.Sdump(dbUser))
			return
		}

	default: // unregonised event type
		// should we also respond to url_verification events? Seems important when
		// setting up service but maybe uneccesary now webhooks endpoint is setup
		// https://api.slack.com/events/url_verification
		log.Debugf("ignoring event of event type %s", event.InnerEvent.Type)
	}
}

func APIToDBUser(in slack.User) db.User {
	// we could either implement this function via marshalling and unmarshalling
	// or via mapping. marshalling and unmarshalling is more extensible
	// but less can go wrong with a mapping function like this
	return db.User{
		Deleted:            in.Deleted,
		ID:                 in.ID,
		Name:               in.Name,
		ProfileImage512:    in.Profile.Image512,
		ProfileStatusEmoji: in.Profile.StatusEmoji,
		ProfileStatusText:  in.Profile.StatusText,
		RealName:           in.RealName,
		TZ:                 in.TZ,
	}
}

func APIToDBUsers(in []slack.User) []db.User {
	out := make([]db.User, len(in))
	for i := 0; i < len(in); i++ {
		out[i] = APIToDBUser(in[i])
	}
	return out
}

// fetchUsersLoop initialises the database with users fetched from the slack api
// and keeps retrying upon errors, call this in a goroutine so it does not block
func (a *App) FetchUsersLoop() {
	for {
		// TODO: make this timeout configurable
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		err := a.FetchUsers(ctx)
		if err == nil {
			return
		}
		log.Errorf(err.Error())
		// it would be nicer to have more sophisticated backoff strategy e.g. but
		// exponential backoff with randomness but we really only need that if
		// we have lots of clients developing into a thundering herd
		// TODO: make this duration configurable
		time.Sleep(time.Second + time.Duration(rand.Intn(1000))*time.Millisecond)
	}
}

// FetchUsers retrieves the initial set of users used to initialise the database
func (a *App) FetchUsers(ctx context.Context) error {
	// GetUsersContext performs paginated requests
	users, err := a.slackClient.GetUsersContext(ctx)
	if err != nil {
		return fmt.Errorf("failed api call to slack GetUsers: %v", err)
	}
	log.Infof("retrieved %d users from GetUsers API", len(users))

	dbUsers := APIToDBUsers(users)
	err = a.db.CreateUsers(dbUsers)
	if err != nil {
		return fmt.Errorf("failed db CreateUsers call: %v", err)
	}
	// TODO: more detailed log message
	log.Infof("wrote %d users to DB", len(dbUsers))
	return nil
}
