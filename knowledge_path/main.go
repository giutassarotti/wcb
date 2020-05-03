package main

import (
    "fmt"
    "log"
    "net/http"
    "io/ioutil"
    "strings"
    "sort"

    //Json
    "github.com/tidwall/gjson"

    //TF-IDF
    "github.com/wilcosheh/tfidf"
	"github.com/wilcosheh/tfidf/similarity"
)

//   http://localhost:8080/knowledge_path?start=Matera&end=Palombaro_lungo&lang=it

type Similarity struct {
    title string
    point float64
}

//Creates topic and language var and looks for errors
func getQuery(request *http.Request) (string, string, string, bool, string) {
    args := request.URL.Query()

    //If there are not the topics in the query
    if len(args["start"]) == 0 {
        log.Println("Error, missing required parameter 'start'")
        return "", "", "", false, "Error, missing required parameter 'start'"
    }

    //If there are not the topics in the query
    if len(args["end"]) == 0 {
        log.Println("Error, missing required parameter 'end'")
        return "", "", "", false, "Error, missing required parameter 'end'"
    }

    //If there is not the language in the query
    if len(args["lang"]) == 0 {
        log.Println("Error, missing required parameter 'lang'")
        return "", "", "", false, "Error, missing required parameter 'lang'"
    }

    start := args["start"][0]
    end := args["end"][0]
    lang := args["lang"][0]

    return start, end, lang, true, ""

}

//Returns the title and the text
func getPage(topic string, baseURL string) (string, string, bool, string) {

    //TODO we need them?
    //Deletes the : links (for now, they seems useless)
    if (strings.Contains(topic,":")) {return "", "", false, ""}

    //Query for reading the title and the text
    //Note: Replace spaces with _ (it wasn't working with spaces)
    resp, err := http.Get(baseURL + "&prop=extracts&format=json&explaintext=true&titles=" + strings.ReplaceAll(topic, " ", "_"))

    if err != nil {
        log.Fatal(err.Error())
        return "", "", false, "Cannot connect to Wikipedia\n" + err.Error()
    }

	json, _ := ioutil.ReadAll(resp.Body)

	//Controls if there's a extract (it wouldn't be a usefull page)
	extract := gjson.Get(string(json), "query.pages.*.extract").Array()
	
	//It means it's not a wikipedia page
    if len(extract) == 0 {return "", "", false, ""}
	
	//topic's exact title
	title := gjson.Get(string(json), "query.pages.*.title").Array()[0].String()

	//topic's text
	text := extract[0].String()


	return title, text, true, ""
}

//Returns the links
func getLinks(main_title string, baseURL string) ([]gjson.Result, bool, string) {
	//Query for errors and the link list
    resp, err := http.Get(baseURL + "&prop=links&format=json&pllimit=max&titles=" + strings.ReplaceAll(main_title, " ", "_"))

    if err != nil {
        log.Fatal(err.Error())
        return nil, false, "Cannot connect to Wikipedia\n" + err.Error()
    }

    json, _ := ioutil.ReadAll(resp.Body)
    
    //Returns links' array (not as a string)
	links := gjson.Get(string(json), "query.pages.*.links.#.title").Array()

	
    
    return links, true, ""
}

//main
func handler(client http.ResponseWriter, request *http.Request) {

	//Returns the topic and the language
	start, end, lang, ok, err := getQuery(request)

	if (!ok) { 
		fmt.Fprintf(client, err)
		return 
	}

    log.Println("\n2\n")

	//Base Url for API
    baseURL := "https://" + lang + ".wikipedia.org/w/api.php?action=query"

    //Returns the exact title and the text of the start 
    start_title, start_text, ok, err := getPage(start, baseURL)

    if (!ok) { 
        fmt.Fprintf(client, err)
        return 
    }

    log.Println(client, "\n3\n")

    //Returns the exact title and the text of the end 
    end_title, end_text, ok, err := getPage(end, baseURL)

    if (!ok) { 
        fmt.Fprintf(client, err)
        return 
    }

    log.Println(client, "\n4\n")

    //Map for saving the texts
    m_texts := make(map[string]string)

    //Adds start text and end text
    m_texts[start_title] = start_text
    m_texts[end_title] = end_text
    
    //Prints the exact topic and the link of the start
    fmt.Fprintf(client, "Start's name:   %s\n\n", start_title)
    fmt.Fprintf(client, "Start's link:   https://%s.wikipedia.org/wiki/%s\n\n\n", lang, strings.ReplaceAll(start_title, " ", "_"))

    //Prints the exact topic and the link of the end
    fmt.Fprintf(client, "End's name:   %s\n\n", end_title)
    fmt.Fprintf(client, "End's link:   https://%s.wikipedia.org/wiki/%s\n\n\n", lang, strings.ReplaceAll(end_title, " ", "_"))

    //Let's start the tfidf
    f := tfidf.New()
    f.AddDocs(start_text)
    f.AddDocs(end_text)

    log.Println(client, "\n5\n")

    //Returns links' array (not as a string)
    links, ok, err := getLinks(start_title, baseURL)

    if (!ok) { 
        fmt.Fprintf(client, err)
        return 
    }

    //Array for saving links and points for similarity
    var link_points []Similarity

    log.Println("6")

    //For every link
    for _, l := range links {

    	//Returns the link's title and text, and a bool for "this is a good link or not"
    	link_title, link_text, ok, err := getPage(l.String(), baseURL)

        //If this is not a good link
        if (!ok) { 
            fmt.Fprintf(client, err) 
            continue
        }

		//Prints links' name
		//fmt.Fprintf(client, "%s\n", link_title)
		
		//Saves every text, if it exixts
		m_texts[link_title] = link_text
		
		//Add the texts for tfidf
		f.AddDocs(link_text)
    }

    log.Println("\n7\n")

    fmt.Fprintf(client, "\n\n")

    //Computes topic's weight 
    w_end := f.Cal(end_text)
    
    //Prints the weight of the end page
    //fmt.Fprintf(client, "Weight of %s is %+v .\n", end_title, w)

    //For every link
    for _, l := range links{
		
		text, ok := m_texts[l.String()]

    	if (!ok){continue}

    	//Computes weight 
    	w_link := f.Cal(text)

    	//Prints weight
    	//fmt.Fprintf(client, "Weight of %s is %+v .\n", l.String(), w_link)

    	//How much they are similar?
    	sim := (similarity.Cosine(w_end, w_link))

		//Saves the similarity
        link_points = append(link_points, Similarity{title: l.String(), point: sim})
		
		//This print is not sorted
        //fmt.Fprintf(client, "Similarity with %s is %f .\n", l.String(), sim)
    }

    log.Println("\n8\n")

    //Sorts by points
    sort.Slice(link_points, func(i, j int) bool {
        return link_points[i].point > link_points[j].point
    })

    fmt.Fprintf(client, "Similarity:\n")

    //Prints the names and points after the sort
    for _, link := range link_points{
        fmt.Fprintf(client, "With %s   =   %f .\n", link.title, link.point)
    }

    log.Println("\n9\n")
}

func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":80", nil))
}
