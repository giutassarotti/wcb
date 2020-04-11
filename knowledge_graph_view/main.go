package main

import (
    "fmt"
    "log"
    "net/http"
)

func handler(client http.ResponseWriter, request *http.Request) {
	keys, ok := request.URL.Query()["topic"]
    
    if !ok || len(keys[0]) < 1 {
        log.Println("Url Param 'topic' is missing")
        fmt.Fprintf(client, "Url Param 'topic' is missing")
        return
    }

    // Query()["key"] will return an array of items, 
    // we only want the single item.
    topic := keys[0]

    log.Println("Url Param 'topic' is: " + string(topic))
    fmt.Fprintf(client, "Url Param 'topic' is: %s", string(topic))
}

func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":80", nil))
}
