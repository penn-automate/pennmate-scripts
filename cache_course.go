package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/penn-automate/penn-opendata-api"
)

const (
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
	api := opendata.NewOpenDataAPI(authBearer, authToken).GetRegistrar()
	stmt, e := db.Prepare("REPLACE INTO course_list VALUES(?,?,?,?,?,?,?,?)")
	if e != nil {
		log.Fatal(e)
	}

	count := 0
	iterator := api.SearchCourseSection(nil)
outer:
	for retry := 1; retry <= maxRetry; retry++ {
		if retry != 1 {
			time.Sleep(waitRetry)
		} else {
			time.Sleep(time.Second)
		}
		if !iterator.NextPage() {
			break
		}

		for i := 0; i < iterator.GetPageSize(); i++ {
			data := new(opendata.CourseSearchData)
			err := iterator.GetResult(data, i)
			if err != nil {
				log.Print(err)
				continue outer
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
				([]byte)(iterator.GetRawData(i)),
			)
			if e != nil {
				log.Println(e)
			} else {
				count++
			}
		}
		retry = 0
		log.Printf("Acquired in total %d listings.", count)
	}
	if err := iterator.GetError(); err != nil {
		log.Print(err)
	}
	if e := stmt.Close(); e != nil {
		log.Fatal(e)
	}
}
