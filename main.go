package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"flag"

	cookiemonster "github.com/MercuryEngineering/CookieMonster"
	"github.com/davecgh/go-spew/spew"
	"github.com/gen2brain/beeep"
	log "github.com/sirupsen/logrus"
)

const (
	Version = "0.1"
)

var currentTickets = make(map[uint]*Ticket, 100)
var debug *bool

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
	log.Infof("Starting version %s", Version)

	cookieFile := flag.String("cookie-file", "cookies.txt", "Path to the cookie file")
	baseUrl := flag.String("base-url", "my.zendesk.com", "Base URL for Zendesk")
	userId := flag.String("userid", "", "You Zendesk user ID")
	debug = flag.Bool("debug", false, "Enable debug output (optional)")

	flag.Parse()

	if len(os.Args) < 4 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	log.SetLevel(log.InfoLevel)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	cookies, err := cookiemonster.ParseFile(*cookieFile)
	if err != nil {
		panic(err)
	}

	detectNewTickets(cookies, *userId, *baseUrl)

	ticker := time.NewTicker(1 * time.Minute)
	quit := make(chan struct{})

	for {
		select {
		case <-ticker.C:
			log.Debugf("Detect tickets fired\n")
			detectNewTickets(cookies, *userId, *baseUrl)
		case <-quit:
			log.Debugf("Stop ticker fired\n")
			ticker.Stop()
			return
		}
	}

}
func detectNewTickets(cookies []*http.Cookie, userId string, baseUrl string) {
	newRes := zendeskCall(cookies, userId, baseUrl)
	for id, t := range newRes {
		if _, ok := currentTickets[id]; ok {
			if t.CommentCount > currentTickets[id].CommentCount {
				url := fmt.Sprintf("https://%s/agent/tickets/%d", baseUrl, t.Id)
				log.Infof("New comment detected:%s\n\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", t.Subject, url, url)
				err := beeep.Alert(fmt.Sprintf("New Comment for %d", t.Id), fmt.Sprintf("New Comment for %s", t.Subject), "assets/warning.png")
				if err != nil {
					panic(err)
				}
				currentTickets[id].CommentCount = t.CommentCount
			}
		} else {
			currentTickets[id] = t
		}
	}
}

type TicketsResponse struct {
	Tickets      []Ticket `json:"tickets"`
	NextPage     string   `json:"next_page"`
	PreviousPage string   `json:"previous_page"`
	Count        uint
}
type Ticket struct {
	Url          string `json:"url"`
	Id           uint   `json:"id"`
	Subject      string `json:"subject"`
	CommentCount uint   `json:"comment_count"`
}

func zendeskCall(cookies []*http.Cookie, userId string, baseUrl string) map[uint]*Ticket {

	url := fmt.Sprintf("https://%s/api/v2/users/%s/tickets/assigned.json?include=comment_count", baseUrl, userId)

	Client := http.Client{
		Timeout: time.Second * 2, // Timeout after 2 seconds
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatalf("New request error: %s", err)
	}

	for _, c := range cookies {
		req.AddCookie(c)
	}
	res, getErr := Client.Do(req)
	if getErr != nil {
		log.Fatalf("Sending request error:%s", getErr)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatalf("Reading response error: %s", readErr)
	}

	log.Debugf("Response %s", body)
	ticketsResponse := TicketsResponse{}
	jsonErr := json.Unmarshal(body, &ticketsResponse)
	if jsonErr != nil {
		log.Fatalf("Converting to json error: %s", jsonErr)
	}
	log.Debugf("Response1=%s\n", spew.Sdump(ticketsResponse))
	parsedTickets := make(map[uint]*Ticket, len(ticketsResponse.Tickets))

	for index, t := range ticketsResponse.Tickets {
		log.Debugf("adding %d->%v\n", t.Id, t)
		parsedTickets[t.Id] = &ticketsResponse.Tickets[index]
		if t.Id == 44783 && *debug {
			if _, ok := currentTickets[44783]; ok {
				parsedTickets[t.Id].CommentCount = currentTickets[t.Id].CommentCount + 1
			}
		}

	}
	log.Debugf("Response2=%s\n", spew.Sdump(parsedTickets))
	return parsedTickets
}
