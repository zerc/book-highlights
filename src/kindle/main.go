// Script to parse highlights from read.amazon.co.uk/notebook
//
// It uses the Chrome Debugging Protocol to interact with the headless Chrome instance.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/client"
)

// The URL of the page from which the highlights should be parsed.
const PageURL = "https://read.amazon.co.uk/notebook"

// The URL to access to the Chrome instance. See docker-compose.yml.
const ChromeClientURL = "http://chrome:9222/json"

// Enable or disable debug output from the chromedp
var ChromeDebug = os.Getenv("CHROME_DEBUG") == "1"

// The endpoint to use to save parsed data somewhere.
var APIEndpoint = os.Getenv("API_ENTRYPOINT")

// The structure to represent a highlight.
type Highlight struct {
	Text        string `json:"text"`
	Colour      string `json:"color"`
	SourceID    string `json:"source_id"`
	SourceTitle string `json:"source_title"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := CreateClient(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if err = OpenPage(ctx, client); err != nil {
		log.Fatal(err)
	}

	var books *[]*map[string]string
	books, err = GetBooks(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	for _, book := range *books {
		log.Printf("Handling book: %+v\n", *book)

		if err := SelectBook(ctx, client, book); err != nil {
			log.Fatal(err)
		}

		var highlights *[]*Highlight

		if highlights, err = GetHighlights(ctx, client, book); err == nil {
			resp, resp_err := CreateHighlights(highlights)

			if resp_err != nil {
				log.Fatal(resp_err)
			} else {
				log.Printf("%s", resp)
			}

		} else {
			log.Fatal(err)
		}
	}

	err = Finish(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

}

// CreateClient creates the new Client to work with Chrome.
func CreateClient(ctx context.Context) (*chromedp.CDP, error) {
	// if ChromeDebug {
	// 	logOption := chromedp.WithLog(log.Printf)
	// } else {
	// 	logOption := chromedp.WithLog(func(s string) { return nil }) // TODO: make it work
	// }

	client, err := chromedp.New(ctx, chromedp.WithTargets(
		client.New(client.URL(ChromeClientURL)).WatchPageTargets(ctx)),
	// logOption,
	)

	return client, err
}

// OpenPage opens the page with highliths.
func OpenPage(ctx context.Context, client *chromedp.CDP) error {
	if err := client.Run(ctx, chromedp.Navigate(PageURL)); err != nil {
		return err
	}

	var loc string
	if err := client.Run(ctx, chromedp.Location(&loc)); err != nil {
		return err
	}

	if loc != PageURL {
		return fmt.Errorf("Invalid page: %s Can't process.", loc)
	}

	if err := client.Run(ctx, chromedp.WaitNotVisible("#kp-notebook-library-spinner", chromedp.ByQuery)); err != nil {
		return err
	}

	return nil
}

// GetBooks parses the current page and returns a pointer to a list of pointers to maps with books information.
func GetBooks(ctx context.Context, client *chromedp.CDP) (*[]*map[string]string, error) {
	var BookNodes []*cdp.Node
	var Books []*map[string]string

	err := client.Run(
		// NOTE: CSS selector like "#kp-notebook-library div.kp-notebook-library-each-book" finds only the first element.
		ctx, chromedp.Nodes("/html[1]/body[1]/div[1]/div[3]/div[1]/div[1]/div[3]/div[1]/div[1]/div", &BookNodes))

	if err != nil {
		return &Books, err
	}

	log.Printf("Found %d book nodes", len(BookNodes))

	// Reverse the list of books (old once are first)
	sort.SliceStable(BookNodes, func(i, j int) bool {
		return i > j
	})

	for _, node := range BookNodes {
		book := make(map[string]string)
		book["id"] = node.AttributeValue("id")

		var title string

		err = client.Run(ctx, chromedp.Text(fmt.Sprintf("#%s h2", book["id"]), &title))
		if err != nil {
			return &Books, err
		}

		book["title"] = title
		Books = append(Books, &book)
	}

	return &Books, nil
}

// SelectBook selects the book i.e. load its page.
// TODO: make sure that the page contains all highlights preloaded
func SelectBook(ctx context.Context, client *chromedp.CDP, book *map[string]string) error {
	log.Printf("Selecting the book: %s", (*book)["id"])

	var tmp string // we don't care
	err := client.Run(
		// NOTE: chromedp.Click doesn't work at all
		ctx, chromedp.EvaluateAsDevTools(
			fmt.Sprintf("$('#%s span').click() && 'ok'", (*book)["id"]), &tmp))

	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second) // being optimistic TODO use WaitVisible

	return nil
}

// GetHighlights parses the highlights of the book specified.
// TODO: handle the pagination. It does look like initially only 100 highlights are displayed.
func GetHighlights(ctx context.Context, client *chromedp.CDP, book *map[string]string) (*[]*Highlight, error) {
	var Highlights []*Highlight
	var nodes []*cdp.Node
	var count int
	const retryCount = 5 // A number of retries to wait until all hightlights are loaded

	if err := client.Run(
		ctx, chromedp.EvaluateAsDevTools("$('div.kp-notebook-highlight').length", &count)); err != nil {
		return &Highlights, err
	}

	for i := 0; i < retryCount; i++ {
		nodes = nodes[:0]

		// NOTE: CSS selector "div.kp-notebook-highlight > span#highlight" doesn't work (chromedp.ByQuery)
		if err := client.Run(ctx, chromedp.Nodes(
			"/html[1]/body[1]/div[1]/div[3]/div[1]/div[2]/div[1]/div[1]/div[1]/div[1]/div[1]/div[3]/*/div[1]/div[2]/div[1]/div[1]/span[1]", &nodes)); err != nil {
			return &Highlights, err
		}

		if len(nodes) == count {
			break
		} else {
			log.Printf(">> Not all highlights are loaded. Waiting. Attempt: %d\n", i)
			time.Sleep(5 * time.Second)
		}
	}

	log.Printf(">> Found %d highligts (%d expected)\n", len(nodes), count)

	for _, node := range nodes {
		hl := Highlight{}
		hl.SourceID = (*book)["id"]
		hl.SourceTitle = (*book)["title"]
		hl.Colour = GetColourFromClass(node.Parent.AttributeValue("class"))

		if err := client.Run(ctx, chromedp.Text(node.FullXPath(), &hl.Text)); err != nil {
			return &Highlights, err
		}

		Highlights = append(Highlights, &hl)
	}

	// Reverse the list of books (old once are first)
	sort.SliceStable(Highlights, func(i, j int) bool {
		return i > j
	})

	return &Highlights, nil
}

// GetColourFromClass returns the proper colour name from the CSS class.
func GetColourFromClass(class string) string {
	classes := strings.Split(class, " ")
	return strings.TrimPrefix(classes[len(classes)-1], "kp-notebook-highlight-")
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

// Finish does required operations to finish the work with the Chrome.
func Finish(ctx context.Context, client *chromedp.CDP) error {
	// The common workflow is that user manually logs in with his credentials and leaves the session opened
	// so the script could use that to grab the data.
	// But we don't want to the page be opened all the time to increase the "security" of the solution i.e.
	// a random "hacker" who managed to get an access to the debug session will not see Amazon page opened.
	err := client.Run(ctx, chromedp.Navigate("about:blank"))
	if err != nil {
		return err
	}

	err = client.Shutdown(ctx)
	if err != nil {
		return err
	}

	err = client.Wait()
	if err != nil {
		return err
	}

	return nil
}
