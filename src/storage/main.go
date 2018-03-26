package main

import (
	"common"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	mgo "gopkg.in/mgo.v2"
)

var dbName = "storage"

var collectionName = "highlights"

// A string to connect to Mongodb
// Example: [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
var mongoConnectionString = "mongo"

// The host to run the web server at
var host = "0.0.0.0:8000"

// A global variable store a reference to the mongo session
var session *mgo.Session

// getSession creates or clones existing Mongodb session
func getSession() *mgo.Session {
	if session == nil {
		var err error
		session, err = mgo.Dial(mongoConnectionString)
		if err != nil {
			log.Fatal("Failed to start the Mongo session")
		}
	}
	return session.Clone()
}

func init() {
	// Create an index to define unique highlights across different sources
	hl_unique_index := mgo.Index{
		Key:        []string{"text", "source_id"},
		Unique:     true, // will fail if the is a duplicate
		DropDups:   false,
		Background: false,
		Sparse:     true,
	}

	// TODO: refactor this somehow
	session := getSession()
	defer session.Close()
	collection := session.DB(dbName).C(collectionName)

	if err := collection.EnsureIndex(hl_unique_index); err != nil {
		log.Fatal("Failed to create index: %s", err)
	} else {
		log.Println("Index is OK")
	}
}

func main() {
	log.Printf("Init server at %s", host)

	// Regester a rout
	// TODO: implement a common solution via regexps
	route := "/api/v1/highlights/"
	http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Support only exact path
		if r.URL.Path != route {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		session := getSession()
		defer session.Close()
		collection := session.DB(dbName).C(collectionName)

		fmt.Printf("Got %s request", r.Method)

		// Support different behaviour depending on the request type
		if r.Method == http.MethodPost {
			CreateHighlights(w, r, collection)
		} else if r.Method == http.MethodGet {
			ListHighlights(w, r, collection)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	log.Fatal(http.ListenAndServe(host, nil))
}

// CreateHighlights creates highlights from the data received
// TODO: validate items received
func CreateHighlights(w http.ResponseWriter, r *http.Request, c *mgo.Collection) {
	defer r.Body.Close()

	var data map[string]*[]common.Highlight
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("Error: %s", err)
		w.WriteHeader(http.StatusBadRequest)
	}

	result := map[string]int{"updated": 0, "inserted": 0, "errors": 0}

	for _, x := range *data["items"] {
		// Should match the  ``hl_unique_index``
		selector := map[string]string{"text": x.Text, "source_id": x.SourceID}

		info, err := c.Upsert(selector, x)

		if err != nil {
			log.Printf("[WARN] Cant't insert a document: %s", err)
			result["error"] += 1
		} else {
			log.Printf("Success: %+v", *info)
			result["updated"] += info.Updated
			if info.UpsertedId != nil {
				result["inserted"] += 1
			}
		}
	}

	if jsonBytes, err := json.Marshal(result); err != nil {
		log.Printf("[ERR] Something went frong: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBytes)
	}
}

// ListHighlights lists all highlitghts from the database
// TODO: pagination!
func ListHighlights(w http.ResponseWriter, r *http.Request, c *mgo.Collection) {
	result := make([]*common.Highlight, 0)
	if err := c.Find(nil).All(&result); err != nil {
		log.Printf("Read error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	if jsonBytes, err := json.Marshal(result); err != nil {
		log.Printf("Marshal error: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write(jsonBytes)
	}
}
