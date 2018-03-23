package main

import (
	"common"
	"encoding/json"
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
	var data map[string]*[]common.Highlight

	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Printf("Error: %s", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	// TODO: validate items received

	log.Printf("Got: %+v", data)

	s := getSession()
	defer s.Close()
	collection := s.DB("storage").C("highlights1")

	items := make([]interface{}, len(*data["items"]))

	for _, x := range *data["items"] {
		items = append(items, x)
	}

	err = collection.Insert(items...)
	if err != nil {
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

	result := make([]*common.Highlight, 0)
	err := collection.Find(nil).All(&result)
	if err != nil {
		log.Printf("Read error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	log.Printf("Result %+v", result[0])

	jsonBytes, err := json.Marshal(result)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBytes)
	} else {
		log.Printf("Marshal error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
