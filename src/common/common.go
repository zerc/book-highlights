// Package contains common things between microservices
package common

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
)

// The endpoint to use to save parsed data somewhere.
var APIEndpoint = os.Getenv("API_ENTRYPOINT")

// The structure to represent a highlight.
type Highlight struct {
	URL         string `bson:"source_url" json:"source_url"`
	Text        string `bson:"text" json:"text"`
	Note        string `bson:"note" json:"note"`
	Colour      string `bson:"colour" json:"color"`
	SourceID    string `bson:"source_id" json:"source_id"`
	SourceTitle string `bson:"source_title" json:"source_title"`
}

// CreateHighlights creates highlights in the data store using REST API.
func CreateHighlights(highlights *[]*Highlight) ([]byte, error) {
	payload, _ := json.Marshal(map[string]interface{}{"items": highlights})

	response, err := http.Post(APIEndpoint, "application/json", bytes.NewBuffer(payload))

	if err != nil {
		return make([]byte, 0), err
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body) // TODO: catch an error here?

	return body, nil
}
