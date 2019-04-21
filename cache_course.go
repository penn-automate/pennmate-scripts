package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type CourseSearch struct {
	ResultData  []json.RawMessage `json:"result_data"`
	ServiceMeta ServiceMeta       `json:"service_meta"`
}
type ResultData struct {
	Activity    string `json:"activity"`
	Credits     string `json:"credits"`
	Instructors []struct {
		Name      string `json:"name"`
		SectionID string `json:"section_id"`
		Term      string `json:"term"`
	} `json:"instructors"`
	MaxEnrollment string `json:"max_enrollment"`
	//MaximumCredit string `json:"maximum_credit"`
	SectionID    string `json:"section_id"`
	SectionTitle string `json:"section_title"`
	Term         string `json:"term"`
}

const (
	courseSearchUrl = `https://esb.isc-seo.upenn.edu/8091/open_data/course_section_search`

	maxRetry  = 5
	waitRetry = time.Second * 3
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", databaseLink)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	request, e := http.NewRequest("GET", courseSearchUrl, nil)
	if e != nil {
		log.Fatal(e)
	}
	request.Header.Add("Authorization-Bearer", authBearer)
	request.Header.Add("Authorization-Token", authToken)
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	stmt, e := db.Prepare("REPLACE INTO course_list VALUES(?,?,?,?,?,?,?,?)")
	if e != nil {
		log.Fatal(e)
	}

	count := 0
	for retry := 1; retry <= maxRetry; retry++ {
		if retry != 1 {
			time.Sleep(waitRetry)
		} else {
			time.Sleep(time.Second)
		}
		response, e := http.DefaultClient.Do(request)
		if e != nil {
			log.Println(e)
			continue
		}

		cs := new(CourseSearch)
		if e := json.NewDecoder(response.Body).Decode(cs); e != nil {
			if e := response.Body.Close(); e != nil {
				log.Println(e)
			}
			log.Println(e)
			continue
		}
		if e := response.Body.Close(); e != nil {
			log.Println(e)
			continue
		}

		for _, rawData := range cs.ResultData {
			data := new(ResultData)
			if e := json.Unmarshal(rawData, data); e != nil {
				log.Println(e)
				continue
			}
			insts := make([]string, 0, len(data.Instructors))
			for _, inst := range data.Instructors {
				insts = append(insts, inst.Name)
			}
			instData, e := json.Marshal(insts)
			if e != nil {
				log.Fatal(e)
			}
			_, e = stmt.Exec(
				data.SectionID,
				data.SectionTitle,
				data.MaxEnrollment,
				instData,
				data.Credits,
				data.Activity,
				data.Term,
				([]byte)(rawData),
			)
			if e != nil {
				log.Println(e)
			} else {
				count++
			}
		}
		if cs.ServiceMeta.CurrentPageNumber == cs.ServiceMeta.NumberOfPages {
			break
		}
		request.URL.RawQuery = fmt.Sprintf("page_number=%d", cs.ServiceMeta.NextPageNumber)
		retry = 0
		log.Printf("Acquired in total %d listings.", count)
	}
	if e := stmt.Close(); e != nil {
		log.Fatal(e)
	}
}
