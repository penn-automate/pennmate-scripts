package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/penn-automate/penn-opendata-api"
)

const serverUpdateURL = `https://pennmate.com/notify.php`

var statusMap = make(map[string]string)

func updateToServer(data opendata.CourseSectionStatus) {
	marshal, e := json.Marshal(data)
	if e != nil {
		log.Fatal(e)
	}
	request, e := http.NewRequest("POST", serverUpdateURL, bytes.NewReader(marshal))
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

func main() {
	log.Println("The app has started running...")
	api := opendata.NewOpenDataAPI(clientId, clientSecret).GetRegistrar()
	for {
		time.Sleep(time.Second * 5)
		data, err := api.GetAllCourseStatus(term)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, d := range data {
			prevStat, ok := statusMap[d.SectionID]
			if ok && d.Status != prevStat {
				log.Printf("Course %s has changed to %s", d.SectionIDNormalized, d.StatusCodeNormalized)
				statusMap[d.SectionID] = d.Status
				//if d.Status == "O" {
				//	go sendMessage(d.SectionIDNormalized)
				//}
				go updateToServer(d)
			} else if !ok {
				statusMap[d.SectionID] = d.Status
			}
		}
	}
}
