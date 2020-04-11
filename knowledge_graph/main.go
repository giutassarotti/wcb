package main

import (
    "fmt"
    "log"
    "net/http"
    "io/ioutil"

     "github.com/tidwall/gjson"
)

func handler(client http.ResponseWriter, request *http.Request) {
	keys, ok := request.URL.Query()["topic"]
    
    if !ok || len(keys[0]) < 1 {
        log.Println("Url Param 'topic' is missing")
        fmt.Fprintf(client, "Url Param 'topic' is missing")
        return
    }

    topic := keys[0]

    resp, err := http.Get("https://en.wikipedia.org/w/api.php?action=query&prop=links&format=json&pllimit=max&titles="+topic)

    if err != nil {
        log.Fatal(err.Error())
        fmt.Fprintf(client, "Cannot connect to Wikipedia\n%s", err.Error())
        return
    }

    json, _ := ioutil.ReadAll(resp.Body)
    fmt.Println(string(json))
    
    links := gjson.Get(string(json), "query.pages.*.links.#.title")
    fmt.Println(links.Array())
    for _, l := range links.Array() {
        fmt.Fprintf(client, "%s\n", l.String())
    }



    /*
    wiki, err := wikimedia.New(wikimedia.Options{
    	Client:    http.DefaultClient,
    	URL:       "https://en.wikipedia.org/w/api.php",
    	UserAgent: "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 " +
    		"(KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36",
    })
    
    if err != nil {
    	log.Fatal(err.Error())
    	fmt.Fprintf(client, "Cannot connect to Wikipedia\n%s", err.Error())
        return
    }
    
    resp, err := wiki.Query(url.Values{
        "action":      {"query"},
        "prop":        {"extracts"},
        "titles":      {topic},
        "explaintext":    {"true"},
    })

	if err != nil {
    	log.Fatal(err.Error())
    	fmt.Fprintf(client, "Cannot connect to Wikipedia\n%s", err.Error())
        return
    }
    
    
	for _, v := range resp.Query.Pages {
		fmt.Fprintf(client, "https://en.wikipedia.org/wiki/%s\n", v.Title)
    	fmt.Fprintf(client, "%s:\n\n", v.Title)
    }
    
    resp1, err := wiki.Query(url.Values{
		"action":      {"query"},
		"prop":        {"links"},
		"pllimit":	   {"20"},
		"titles":      {topic},
	})
	
	if err != nil {
    	log.Fatal(err.Error())
    	fmt.Fprintf(client, "Cannot connect to Wikipedia\n%s", err.Error())
        return
    }
	
	for k, v := range resp1.Query.Pages {
        fmt.Println(v.Original)
    	fmt.Fprintf(client, "%s:\n%s\n", k, v.Original)
    }
    */
}

func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":80", nil))
}
