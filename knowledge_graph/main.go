package main

import (
    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "strings"

     "github.com/tidwall/gjson"
)

func handler(client http.ResponseWriter, request *http.Request) {

	//Query that returns the topics (for now only 1) and errors
    args := request.URL.Query()

    if len(args["topic"]) == 0 {
        log.Println("Error, missing required parameter 'topic'")
        fmt.Fprintf(client, "Error, missing required parameter 'topic'")
        return
    }

    if len(args["lang"]) == 0 {
        log.Println("Error, missing required parameter 'lang'")
        fmt.Fprintf(client, "Error, missing required parameter 'lang'")
        return
    }

    //topic is the name of the searched page 
    topic := args["topic"][0]
    lang := args["lang"][0]

    //Base Url for API
    var baseURL = "https://" + lang + ".wikipedia.org/w/api.php?action=query"

    //print the topic's name
    fmt.Fprintf(client, "%s\n\n", topic)

    //Query that returns the topics (for now only 1) and errors for reading the title and the text
    resp, err := http.Get(baseURL + "&prop=extracts&format=json&explaintext=true&titles=" + topic)

    if err != nil {
        log.Fatal(err.Error())
        fmt.Fprintf(client, "Cannot connect to Wikipedia\n%s", err.Error())
        return
    }

    json, _ := ioutil.ReadAll(resp.Body)
    fmt.Println(string(json))
    
    titles := gjson.Get(string(json), "query.pages.*.title")
    
    for _, t := range titles.Array() {
        fmt.Fprintf(client, "https://%s.wikipedia.org/wiki/%s\n\n", lang, t.String())
    }

    //Query that returns the topics (for now only 1) and errors for reading the link list
    resp, err = http.Get(baseURL + "&prop=links&format=json&pllimit=max&titles=" + topic)

    if err != nil {
        log.Fatal(err.Error())
        fmt.Fprintf(client, "Cannot connect to Wikipedia\n%s", err.Error())
        return
    }

    json, _ = ioutil.ReadAll(resp.Body)
    fmt.Println(string(json))
    
    links := gjson.Get(string(json), "query.pages.*.links.#.title")
    
    for _, l := range links.Array() {
        if !(strings.Contains(l.String(),":")) {
            fmt.Fprintf(client, "%s\n", l.String())
        }
    }

}

func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":80", nil))
}
