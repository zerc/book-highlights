package main

import (
	"common"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	mgo "gopkg.in/mgo.v2"
)

var host = "0.0.0.0:8000"

var session *mgo.Session

func getSession() *mgo.Session {
	if session == nil {
		var err error
		session, err = mgo.Dial("mongo")
		if err != nil {
			log.Fatal("Failed to start the Mongo session")
		}
	}
	return session.Clone()
}

func main() {
	log.Printf("Init server at %s", host)
	http.HandleFunc("/api/v1/highlights/", HighlightsHandler)
	log.Fatal(http.ListenAndServe(host, nil))
}

func HighlightsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		CreateHighlights(w, r)
	} else if r.Method == http.MethodGet {
		ListHighlights(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func CreateHighlights(w http.ResponseWriter, r *http.Request) {
	// var highlights []*common.Highlight
	var items map[string]*[]common.Highlight

	defer r.Body.Close()

	// body, _ := ioutil.ReadAll(r.Body) // TODO: catch an error here?
	// err := json.Unmarshal(body, &items)
	err := json.NewDecoder(r.Body).Decode(&items)
	if err != nil {
		log.Printf("Error: %s", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	log.Printf("Got: %+v", items)

	s := getSession()
	defer s.Close()
	collection := s.DB("storage").C("highlights1")
	hl := *items["items"]

	bulk := collection.Bulk()
	bulk.Insert(hl...)

	if _, err := bulk.Run(); err != nil {
		log.Printf("DB Error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func ListHighlights(w http.ResponseWriter, r *http.Request) {
	log.Printf("Got GET request")

	s := getSession()
	defer s.Close()
	collection := s.DB("storage").C("highlights1")

	var highlights []common.Highlight
	err := collection.Find(nil).All(&highlights)

	if err != nil {
		log.Printf("Read error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	log.Printf("Result %+v", highlights)

	result, err := json.Marshal(highlights)
	if err != nil {
		fmt.Fprintf(w, "%s", result)
		w.WriteHeader(http.StatusOK)
	} else {
		log.Printf("Marshal error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
