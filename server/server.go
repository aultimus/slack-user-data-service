package server

import (
	"net/http"
	"time"

	log "github.com/cocoonlife/timber"
	"github.com/gorilla/mux"
)

const (
	ContentType    = "Content-Type"
	DefaultPortNum = "8080"
	MimeTypeJSON   = "application/json"
)

func NewApp() *App {
	return &App{}
}

type App struct {
	server *http.Server
}

func (a *App) Init(portNum string) error {
	log.Infof("init")
	router := mux.NewRouter()

	a.server = &http.Server{
		Addr:           ":" + portNum,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return nil
}

func (a *App) Run() error {
	log.Infof("running server on %s", a.server.Addr)
	return a.server.ListenAndServe()
}
