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
	URL         string `json:"source_url"`
	Text        string `json:"text"`
	Note        string `json:"note"`
	Colour      string `json:"color"`
	SourceID    string `json:"source_id"`
	SourceTitle string `json:"source_title"`
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
