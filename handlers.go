package main

import (
    "encoding/json"
    "errors"
	"fmt"
    "io/ioutil"
	"net/http"
	"time"

	"github.com/pmylund/go-cache"
)

type Bundle struct {
    Cmd string
}

var c = cache.New(ItemLifetime, ItemLifetime * 2)

func makeMessage(cmd string) interface{} {
	message := make(map[string]interface{})
	message["cmd"] = cmd
	message["created"] = time.Now().Format(time.RFC3339Nano)
	return message
}

func StreamsStreamingGetHandler(w http.ResponseWriter, r *http.Request) {
    key := r.Header.Get("auth_key")
    if AuthKey != key {
        http.Error(w, "An error has occurred", http.StatusInternalServerError)
        return
    }
    
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "An error has occurred", http.StatusInternalServerError)
		return
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, "An error has occurred", http.StatusInternalServerError)
		return
	}

	// Don't forget to close the connection:
	defer conn.Close()
	defer conn.Write(Chunk(""))

	fmt.Fprintf(bufrw, "HTTP/1.1 200 OK\r\n")
	fmt.Fprintf(bufrw, "Transfer-Encoding: chunked\r\n")
	fmt.Fprintf(bufrw, "Content-Type: application/json\r\n\r\n")
	bufrw.Flush()
    
	messageBus := TopicMap.Register(key)
	defer TopicMap.Unregister(key, messageBus)

	// Keepalive ticker
	ticker := time.Tick(30 * time.Second)
	for {
		var err error
		select {
		case message, ok := <-messageBus:
			if !ok {
				return
			}
			assertedMessage := message.(map[string]interface{})
			_, err = conn.Write(Chunk(JSONToString(assertedMessage) + "\n"))
		case _ = <-ticker:
			// Send the keepalive.
			_, err = conn.Write(Chunk("\n"))
		}

		// An error means the connection was closed, return.
		if err != nil {
			return
		}
	}
}

func StreamsPostHandler(w http.ResponseWriter, r *http.Request) {
    key := r.Header.Get("auth_key")
    if AuthKey != key {
        http.Error(w, "An error has occurred", http.StatusInternalServerError)
        return
    }

	decoder := json.NewDecoder(r.Body)
    var b Bundle
	err := decoder.Decode(&b)

	if err != nil {
		http.Error(w, "An error has occurred", http.StatusInternalServerError)
        return
	}
    
	message := makeMessage(b.Cmd)

	// Write the message to the cache.
	CacheBus <- CacheMessage{1, message, key}

	fmt.Fprint(w, "RCVD")
}
