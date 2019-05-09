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

func updateToServer(data opendata.CourseStatusData) {
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
	api := opendata.NewOpenDataAPI(authBearer, authToken).GetRegistrar()
	for {
		time.Sleep(time.Second * 3)
		data, err := api.GetAllCourseStatus(term)
		if err != nil {
			log.Println(err)
			continue
		}
		for _, data := range data {
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
