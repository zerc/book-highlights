package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
)

func main() {
	ctx := context.Background()

	b, err := ioutil.ReadFile("/tmp/.credentials/client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/drive-go-quickstart.json
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := drive.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve drive Client %v", err)
	}

	dirId, err := getHighlightsDirectoryID(srv)
	if err != nil {
		log.Fatalf("Can't find the directory with your highlights: %v", err)
	}

	log.Printf("Directory ID: %s", dirId)

	files, err := getFiles(srv, dirId)
	if err != nil {
		log.Fatalf("Can't fetch fils in directory. %v", err)
	}

	for i, f := range *files {

		log.Println(i, f)

		response, err := srv.Files.Export(f.Id, "text/html").Download()

		if err != nil {
			log.Fatalf("Can't export file %s due to %s", f.Id, err)
		}

		defer response.Body.Close()
		body, _ := ioutil.ReadAll(response.Body)

		log.Printf("%s", body)
	}
}

// getHighlightsDirectoryID gets ID of the directory where Google Books highlights are living
func getHighlightsDirectoryID(srv *drive.Service) (string, error) {
	r, err := srv.Files.List().Q("mimeType = 'application/vnd.google-apps.folder' and name = 'Play Books Notes'").Do()

	if err != nil {
		return "", err
	}

	if len(r.Files) > 0 {
		return r.Files[0].Id, nil
	}

	return "", fmt.Errorf("Not found")
}

// getFiles returns a pointer to a list of pointers to files
// TODO: use channels to iterate over fiels and their content
func getFiles(srv *drive.Service, dirId string) (*[]*drive.File, error) {
	r, err := srv.Files.List().Q(fmt.Sprintf("mimeType = 'application/vnd.google-apps.document' and '%s' in parents", dirId)).Do()

	if err != nil {
		var files []*drive.File
		return &files, err
	}

	return &r.Files, nil
}
