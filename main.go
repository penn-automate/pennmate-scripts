package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

type CourseStatus struct {
	ResultData []struct {
		CourseSection        string `json:"course_section"`
		SectionIDNormalized  string `json:"section_id_normalized"`
		PreviousStatus       string `json:"previous_status"`
		Status               string `json:"status"`
		StatusCodeNormalized string `json:"status_code_normalized"`
		Term                 string `json:"term"`
	} `json:"result_data"`
	ServiceMeta struct {
		CurrentPageNumber  int    `json:"current_page_number"`
		ErrorText          string `json:"error_text"`
		NextPageNumber     int    `json:"next_page_number"`
		NumberOfPages      int    `json:"number_of_pages"`
		PreviousPageNumber int    `json:"previous_page_number"`
		ResultsPerPage     int    `json:"results_per_page"`
	} `json:"service_meta"`
}

const (
	courseStatusUrl = `https://esb.isc-seo.upenn.edu/8091/open_data/course_status/2019C/all`
	authBearer      = `UPENN_OD_emjT_1002233`
	authToken       = `2bbk83v7fasdth34j5asas`
)

var (
	statusMap  = make(map[string]string)
	messageTtl = time.Minute * 10
	sendClient *messaging.Client
)

func sendMessage(course string) {
	message := &messaging.Message{
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Title: course,
				Body:  "The course opens just now.",
				Sound: "default",
			},
			TTL: &messageTtl,
		},
		Topic: strings.ReplaceAll(strings.ReplaceAll(course, "-", ""), " ", "%"),
	}

	s, err := sendClient.Send(context.Background(), message)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully sent message:", s)
}

func init() {
	opt := option.WithCredentialsFile("penn-automate.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}
	sendClient, err = app.Messaging(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	log.Println("The app has started running...")
}

func main() {
	request, e := http.NewRequest("GET", courseStatusUrl, nil)
	if e != nil {
		log.Fatal(e)
	}
	request.Header.Add("Authorization-Bearer", authBearer)
	request.Header.Add("Authorization-Token", authToken)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	for {
		response, e := http.DefaultClient.Do(request)
		if e != nil {
			log.Fatal(e)
		}
		cs := new(CourseStatus)
		if e = json.NewDecoder(response.Body).Decode(cs); e != nil {
			log.Fatal(e)
		}
		if err := response.Body.Close(); err != nil {
			log.Fatal(err)
		}
		for _, data := range cs.ResultData {
			prevStat, ok := statusMap[data.CourseSection]
			if ok && data.Status != prevStat {
				log.Printf("Course %s has changed to %s", data.SectionIDNormalized, data.StatusCodeNormalized)
				statusMap[data.CourseSection] = data.Status
				if data.Status == "O" {
					go sendMessage(data.SectionIDNormalized)
				}
			} else if !ok {
				statusMap[data.CourseSection] = data.Status
			}
		}
		time.Sleep(time.Second * 3)
	}
}
