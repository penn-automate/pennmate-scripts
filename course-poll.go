package main

import (
	"bytes"
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
	ResultData  []ResultData `json:"result_data"`
	ServiceMeta ServiceMeta  `json:"service_meta"`
}
type ResultData struct {
	CourseSection        string `json:"course_section"`
	SectionIDNormalized  string `json:"section_id_normalized"`
	PreviousStatus       string `json:"previous_status"`
	Status               string `json:"status"`
	StatusCodeNormalized string `json:"status_code_normalized"`
	Term                 string `json:"term"`
}

const (
	courseStatusUrl = `https://esb.isc-seo.upenn.edu/8091/open_data/course_status/2019C/all`
	serverUpdateUrl = `https://pennmate.com/notify.php`
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
		log.Println(err)
	} else {
		log.Println("Successfully sent message:", s)
	}
}

func updateToServer(data ResultData) {
	marshal, e := json.Marshal(data)
	if e != nil {
		log.Fatal(e)
	}
	request, e := http.NewRequest("POST", serverUpdateUrl, bytes.NewReader(marshal))
	if e != nil {
		log.Fatal(e)
	}
	request.SetBasicAuth(httpUser, httpPass)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	response, e := http.DefaultClient.Do(request)
	if e == nil {
		log.Println("Updated to server:", response.Status)
	}
}

func init() {
	opt := option.WithCredentialsFile(firebaseJson)
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
		time.Sleep(time.Second * 3)
		response, e := http.DefaultClient.Do(request)
		if e != nil {
			log.Println(e)
			continue
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
				//if data.Status == "O" {
				//	go sendMessage(data.SectionIDNormalized)
				//}
				go updateToServer(data)
			} else if !ok {
				statusMap[data.CourseSection] = data.Status
			}
		}
	}
}