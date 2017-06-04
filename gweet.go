package main

import (
	"flag"
	"io/ioutil"
	"log"
    // "strconv"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

const MaxQueueLength = 3
const ItemLifetime = 15 * time.Minute

func main() {
	var debug_enabled = flag.Bool("debug", false, "Enable debug logging")
	// var intf = flag.String("interface", "0.0.0.0", "The interface to listen on")
	// var port = flag.Int("port", 9835, "The port to listen on")
	flag.Parse()
	if *debug_enabled {
		InitLogging(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else {
		InitLogging(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)
	}

	r := mux.NewRouter()
	r.HandleFunc("/relay/", StreamsStreamingGetHandler).Methods("GET").Queries("streaming", "1")
	r.HandleFunc("/relay/", StreamsPostHandler).Methods("POST")

	go Cacher()
	INFO.Println("Listening...")
	log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), Log(r)))
    // log.Fatal(http.ListenAndServe(*intf+":"+strconv.Itoa(*port), Log(r)))
}
