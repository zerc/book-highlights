package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"sort"
	"strings"

	"common"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	drive "google.golang.org/api/drive/v3"
)

var colourRe = regexp.MustCompile("background-color:(#.{6})")

var coloursMap = map[string]string{
	"#93e3ed": "blue",
	"#fde096": "yellow",
	"#ffb8a1": "red",
	"#c5e1a5": "green",
}

// The fallback if the colour isn't detected.
var defaultColour = "yellow"

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

	for _, f := range *files {
		log.Printf("Handle file: %s", f.Name)

		doc, err := getFileContent(srv, f.Id)

		if err != nil {
			log.Fatalf("Can't export file %s due to: %s", f.Id, err)
		}

		highlights := getHighlights(f, doc)

		if len(*highlights) > 0 {
			resp, resp_err := common.CreateHighlights(highlights)

			if resp_err != nil {
				log.Fatal(resp_err)
			} else {
				log.Printf("%s", resp)
			}
		} else {
			log.Fatalf("Unable to fetch highlights from %s due to %s", f.Name, err)
		}

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

// getFileContent gets the file's content as HTML Document
func getFileContent(srv *drive.Service, fileId string) (*goquery.Document, error) {
	var doc goquery.Document

	response, err := srv.Files.Export(fileId, "text/html").Download()

	if err != nil {
		return &doc, err
	}

	if response.StatusCode != 200 {
		return &doc, fmt.Errorf("Status code: %d", response.StatusCode)
	}

	return goquery.NewDocumentFromResponse(response)
}

func getHighlights(f *drive.File, doc *goquery.Document) *[]*common.Highlight {
	var Highlights []*common.Highlight

	doc.Find("table table tr td:nth-child(2)").Each(func(i int, s *goquery.Selection) {
		// Avoid "bad" highliths i.e. without text
		if s.Find("p").Size() < 3 {
			html, _ := s.Html()
			log.Printf("Warn: invalid HTML block:\n%s\n", html)
			return
		}

		text := strings.TrimSpace(s.Find("p:nth-child(1)").Text())
		if text == "" {
			log.Printf("Warn: no text found in the block.")
			return
		}

		hl := common.Highlight{Text: text}
		hl.Note = strings.TrimSpace(s.Find("p:nth-child(2)").Text())

		link, _ := s.ParentFiltered("tr").Find("a").Attr("href")
		hl.URL = link
		hl.Colour = getColourFromSelection(s)
		hl.SourceTitle = f.Name
		hl.SourceID = f.Id

		Highlights = append(Highlights, &hl)
	})

	// Reverse the order
	sort.SliceStable(Highlights, func(i, j int) bool {
		return i > j
	})

	return &Highlights
}

func getColourFromSelection(s *goquery.Selection) string {
	style, _ := s.Find("p:nth-child(1) span").Attr("style")
	matches := colourRe.FindStringSubmatch(style)

	var colour string

	if len(matches) >= 2 {
		hex := matches[1]
		colour = coloursMap[hex]
	}

	if colour == "" {
		return defaultColour
	} else {
		return colour
	}
}
