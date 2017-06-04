package main

import (
    "errors"
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
const ItemLifetime = 5 * time.Minute
var AuthKey string

func getKey() (string, error) {
    b, err := ioutil.ReadFile("key.txt")
    if err != nil {
        return "", errors.New("Error reading key file")
    }

    str := string(b)
    return str, nil
}

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
    
    key, key_err := getKey()
    
    if key_err != nil {
        INFO.Println("ERROR reading authorization key from key file")
        return
    }
    
    if key == "" {
        INFO.Println("ERROR authorization key is empty")
        return
    }
    
    AuthKey = key

	r := mux.NewRouter()
	r.HandleFunc("/relay/", StreamsStreamingGetHandler).Methods("GET").Queries("streaming", "1")
	r.HandleFunc("/relay/", StreamsPostHandler).Methods("POST")

	go Cacher()
	INFO.Println("Listening...")
	log.Fatal(http.ListenAndServe(":" + os.Getenv("PORT"), Log(r)))
    // log.Fatal(http.ListenAndServe(*intf+":"+strconv.Itoa(*port), Log(r)))
}
