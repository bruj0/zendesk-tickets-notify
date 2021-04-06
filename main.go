package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	cookiemonster "github.com/MercuryEngineering/CookieMonster"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

const (
	Version = "0.1"
)

var currentTickets = make(map[uint]*Ticket, 100)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)
	log.Infof("Starting version %s", Version)

	cookies, err := cookiemonster.ParseFile("cookies.txt")
	if err != nil {
		panic(err)
	}
	userId := "389137184493"

	detectNewTickets(cookies, userId)

	ticker := time.NewTicker(1 * time.Minute)
	quit := make(chan struct{})

	for {
		select {
		case <-ticker.C:
			log.Debugf("Detect tickets fired\n")
			detectNewTickets(cookies, userId)
		case <-quit:
			log.Debugf("Stop ticker fired\n")
			ticker.Stop()
			return
		}
	}

}
func detectNewTickets(cookies []*http.Cookie, userId string) {
	newRes := zendeskCall(cookies, userId)
	for id, t := range newRes {
		if val, ok := currentTickets[id]; ok {
			if id == 44783 {
				log.Debugf("Checking %d comments %d>%d\n", id, val.CommentCount, currentTickets[id].CommentCount)
			}
			if val.CommentCount > currentTickets[id].CommentCount {
				log.Debugf("New comment detected:%s", t)
			}
		} else {
			currentTickets[id] = t
			if id == 44783 {
				log.Debugf("New ticket adaded %s", spew.Sdump(currentTickets[id]))
			}
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

func zendeskCall(cookies []*http.Cookie, userId string) map[uint]*Ticket {

	url := "https://hashicorp.zendesk.com/api/v2/users/" + userId + "/tickets/assigned.json?include=comment_count"

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

	//log.Debugf("Response %s", body)
	ticketsResponse := TicketsResponse{}
	jsonErr := json.Unmarshal(body, &ticketsResponse)
	if jsonErr != nil {
		log.Fatalf("Converting to json error: %s", jsonErr)
	}
	//log.Debugf("Response1=%s\n", spew.Sdump(ticketsResponse))
	responseTickets := make(map[uint]*Ticket, len(ticketsResponse.Tickets))

	for index, t := range ticketsResponse.Tickets {
		//log.Debugf("adding %d->%v\n", t.Id, t)
		responseTickets[t.Id] = &ticketsResponse.Tickets[index]
		if t.Id == 44783 {
			if _, ok := currentTickets[44783]; ok {
				t.CommentCount = currentTickets[44783].CommentCount + 1
			}
		}
	}
	//log.Debugf("Response2=%s\n", spew.Sdump(responseTickets))
	return responseTickets
}
