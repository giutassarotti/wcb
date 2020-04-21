package main

import (
    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "strings"

    //Json
    "github.com/tidwall/gjson"

    //TF-IDF
    "github.com/wilcosheh/tfidf"
	"github.com/wilcosheh/tfidf/similarity"
)

//   http://localhost:8080/knowledge_graph?topic=dog&lang=en

func handler(client http.ResponseWriter, request *http.Request) {

	//Query that returns the topics (for now only 1) and errors
    args := request.URL.Query()

    //If there is not the topic(s) in the query
    if len(args["topic"]) == 0 {
        log.Println("Error, missing required parameter 'topic'")
        fmt.Fprintf(client, "Error, missing required parameter 'topic'")
        return
    }

    //Ff there is not the language in the query
    if len(args["lang"]) == 0 {
        log.Println("Error, missing required parameter 'lang'")
        fmt.Fprintf(client, "Error, missing required parameter 'lang'")
        return
    }

    topic := args["topic"][0]
    lang := args["lang"][0]


    
    //Prints the topic's name
    fmt.Fprintf(client, "Topic:   %s\n\n\n", topic)

    
    //Base Url for API
    var baseURL = "https://" + lang + ".wikipedia.org/w/api.php?action=query"

    //Query that returns the topics (for now only 1) and errors for reading the title and the text
    resp, err := http.Get(baseURL + "&prop=extracts&format=json&explaintext=true&titles=" + topic)

    if err != nil {
        log.Fatal(err.Error())
        fmt.Fprintf(client, "Cannot connect to Wikipedia\n%s", err.Error())
        return
    }

	json, _ := ioutil.ReadAll(resp.Body)

	//topic's text
	main_text := gjson.Get(string(json), "query.pages.*.extract").Array()[0].String()
    
    //Prints the text
    //fmt.Fprintln(client, main_text)
    
    titles := gjson.Get(string(json), "query.pages.*.title")
    
    //Prints the title (better using a for, but it's just one)
    for _, t := range titles.Array() {
        fmt.Fprintf(client, "Topic's link:   https://%s.wikipedia.org/wiki/%s\n\n\n", lang, t.String())
    }

    




    //Query that returns the topics (for now only 1) and errors for reading the link list
    resp, err = http.Get(baseURL + "&prop=links&format=json&pllimit=max&titles=" + topic)

    if err != nil {
        log.Fatal(err.Error())
        fmt.Fprintf(client, "Cannot connect to Wikipedia\n%s", err.Error())
        return
    }

    json, _ = ioutil.ReadAll(resp.Body)
    
    //Prints the list of links on logs
    //fmt.Println(string(json))
    
	links := gjson.Get(string(json), "query.pages.*.links.#.title")
    
    //Prints the links
    //fmt.Fprintf(client, "All the Link in the page:\n\n")

    for _, l := range links.Array() {
        if !(strings.Contains(l.String(),":")) {

           //fmt.Fprintf(client, "%s\n", l.String())
        }
    }



    //Let's start to select best links by tfidf (just with 2 pages each time)
    f := tfidf.New()
	f.AddDocs(main_text)


    //Map for saving links and points in tfidf
    m_link_point := make(map[string]float64)

    //Map for saving the texts, if they exixt
    m_link_text := make(map[string]string)

    //fmt.Fprintf(client, "Best Links in the page:\n\n")

    
    //For every link
    for _, l := range links.Array() {
    	
    	//Deletes the : links (for now, they seems useless)
    	if (strings.Contains(l.String(),":")) {
    		continue
    	}

	    resp, err = http.Get(baseURL + "&prop=extracts&format=json&explaintext=true&titles=" + strings.ReplaceAll(l.String(), " ", "_"))

    	if err != nil {
	        log.Fatal(err.Error())
	        fmt.Fprintf(client, "Cannot connect to Wikipedia\n%s", err.Error())
	        return
	    }

		json, _ = ioutil.ReadAll(resp.Body)



		//About link's page text 
		link_extract := gjson.Get(string(json), "query.pages.*.extract").Array()

		//If there isn't the text
		if len(link_extract) == 0{
			continue
		}

		//Prints links' name
		fmt.Fprintf(client, "%s\n", l.String())

		//link's page text 
		link_text := link_extract[0].String()
		
		//Saves every text, if it exixts
		m_link_text[l.String()] = link_text

		//Prints the link's text
		//fmt.Fprintln(client, link_text)
		
		//Add the texts for tfidf
		f.AddDocs(link_text)

    }

    fmt.Fprintf(client, "\n\n\n")

    //Computes topic's weight 
    w := f.Cal(main_text)
    
    //fmt.Fprintf(client, "Weight of %s is %+v .\n", topic, w)

    //For every link
    for _, l := range links.Array(){
    	_, ok := m_link_text[l.String()]

    	if (!ok){
    		continue
    	}

    	//Computes weight 
    	w_link := f.Cal(m_link_text[l.String()])

    	//Prints weight
    	//fmt.Fprintf(client, "Weight of %s is %+v .\n", l.String(), w_link)

    	//How much they are similar?
    	sim := (similarity.Cosine(w, w_link) - 0.5)*2
		fmt.Fprintf(client, "Similarity with %s is %f .\n", l.String(), sim)

		//Saves the similarity
		m_link_point[l.String()] = sim
    }

}

func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":80", nil))
}
