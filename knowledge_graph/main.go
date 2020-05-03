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

//   http://localhost:8080/knowledge_graph?topic=dog&lang=en

type Similarity struct {
    title string
    point float64
}

//Creates topic and language var and looks for errors
func getQuery(request *http.Request) (string, string, bool, string) {
    args := request.URL.Query()

    //If there is not the topic(s) in the query
    if len(args["topic"]) == 0 {
        log.Println("Error, missing required parameter 'topic'")
        return "", "", false, "Error, missing required parameter 'topic'"
    }

    //If there is not the language in the query
    if len(args["lang"]) == 0 {
        log.Println("Error, missing required parameter 'lang'")
        return "", "", false, "Error, missing required parameter 'lang'"
    }

    topic := args["topic"][0]
    lang := args["lang"][0]

    return topic, lang, true, ""

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
	topic, lang, ok, err := getQuery(request)

	if (!ok) { 
		fmt.Fprintf(client, err)
		return 
	}

	//Base Url for API
    baseURL := "https://" + lang + ".wikipedia.org/w/api.php?action=query"

    //Returns the exact title and the text
    main_title, main_text, ok, err := getPage(topic, baseURL)

    if (!ok) { 
        fmt.Fprintf(client, err)
        return 
    }
    
    //Prints the exact topic and the link
    fmt.Fprintf(client, "Topic's name:   %s\n\n", main_title)
    fmt.Fprintf(client, "Topic's link:   https://%s.wikipedia.org/wiki/%s\n\n\n", lang, main_title)

	//Prints the text
    //fmt.Fprintln(client, main_text)
    
    //Returns links' array (not as a string)
    links, ok, err := getLinks(main_title, baseURL)

    if (!ok) { 
        fmt.Fprintf(client, err)
        return 
    }
	

	//Let's start to select best links by tfidf (just with 2 pages each time)
    f := tfidf.New()
	f.AddDocs(main_text)


    //Array for saving links and points for similarity
    var link_points []Similarity
	
    //Map for saving the texts, if they exixt
    m_link_text := make(map[string]string)

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
		fmt.Fprintf(client, "%s\n", link_title)
		
		//Saves every text, if it exixts
		m_link_text[link_title] = link_text

		//Prints the link's text
		//fmt.Fprintln(client, link_text)
		
		//Add the texts for tfidf
		f.AddDocs(link_text)

    }
    fmt.Fprintf(client, "\n\n\n")

    
    //Computes topic's weight 
    w := f.Cal(main_text)
    
    var sim float64
    
    //Prints the weight of the main page
    //fmt.Fprintf(client, "Weight of %s is %+v .\n", main_title, w)
    
    //For every link
    for _, l := range links{
		
		text, ok := m_link_text[l.String()]

    	if (!ok){continue}

    	//Computes weight 
    	w_link := f.Cal(text)

    	//Prints weight
    	//fmt.Fprintf(client, "Weight of %s is %+v .\n", l.String(), w_link)

    	//How much they are similar?
    	sim = (similarity.Cosine(w, w_link))
		//fmt.Fprintf(client, "Similarity with %s is %f .\n", l.String(), sim)

		//Saves the similarity
        link_points = append(link_points, Similarity{title: l.String(), point: sim})
		
		//This is not sorted
        fmt.Fprintf(client, "Similarity with %s is %f .\n", l.String(), sim)
    }

    //Sorts by points
    sort.Slice(link_points, func(i, j int) bool {
        return link_points[i].point > link_points[j].point
    })

    fmt.Fprintf(client, "Similarity:\n")

    //Prints the names and points after the sort
    for _, link := range link_points{
        //fmt.Fprintf(client, "With %s   =   %f .\n", link.title, link.point)
        fmt.Println(link)
    }
}

func main() {
    http.HandleFunc("/", handler)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
