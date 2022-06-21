package main

//TODO
// - use CLI or config to permit alternate IP & port binding
import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/ostafen/clover"
)

const (
	BINDPORT string = "80" //change to suit requirement
)

var wait time.Duration

var index *View
var jobs *View

//setup some route endpoints
func newRouter() *mux.Router {
	r := mux.NewRouter()

	//default handler
	r.HandleFunc("/", indexHandler).Methods("GET")

	// setup static content route - strip ./assets/assets/[resource]
	// to keep /assets/[resource] as a route
	staticFileDirectory := http.Dir("./assets/")
	staticFileHandler := http.StripPrefix("/assets/", http.FileServer(staticFileDirectory))
	r.PathPrefix("/assets/").Handler(staticFileHandler).Methods("GET")

	// setup routes to handle job updates and job notes
	r.HandleFunc("/jobs", getJobsHandler).Methods("GET")
	r.HandleFunc("/jobs/{s:[0-9]+}", getJobsHandler).Methods("GET")
	r.HandleFunc("/jobs/{s:[0-9]+}/{f:[s,a,i,t,c]?}", getJobsHandler).Methods("GET")

	r.HandleFunc("/job/{id:[0-9]+}", getJobHandler).Methods("GET")
	r.HandleFunc("/job/{id:[0-9]+}", editJobHandler).Methods("POST")

	r.HandleFunc("/notes/{id:[0-9]+}", getJobNoteHandler).Methods("GET")
	r.HandleFunc("/notes/{id:[0-9]+}", editJobNoteHandler).Methods("POST")
	return r
}

//used to auto detect the active local IP address - not used yet
func GetOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func main() {

	//flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	//id, err := flake.NextID()
	//create the initial collection using some sample data, this only gets created
	//if the collection does not exist
	_, err := os.Stat("./data/jobs/MANIFEST")
	if os.IsNotExist(err) {
		log.Println("Importing demo data")
		loadFromJson("./data/jobs.json")
	}

	jobsdb, err = clover.Open("./data/jobs")
	log.Println("Opening jobs collection")
	if err != nil {
		panic(err.Error())
	}

	// load templates
	log.Println("Loading templates")
	index = NewView("bootstrap", "views/index.gohtml")
	jobs = NewView("bootstrap", "views/jobs.gohtml")

	log.Println("Starting HTTP service on " + BINDPORT)
	r := newRouter()

	// setup HTTP on gorilla mux for a gracefull shutdown
	srv := &http.Server{
		Addr: "0.0.0.0:" + BINDPORT,

		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	// HTTP listener is in a goroutine as its blocking
	go func() {
		if err = srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// setup a ctrl-c trap to ensure a graceful shutdown
	// this would also allow shutting down other pipes/connections. eg DB
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	jobsdb.Close()
	log.Println("shutting down")
	os.Exit(0)
}
