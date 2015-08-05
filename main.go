package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	path = flag.String("path", "", "path to ServiceReviewCompliance exported json")
)

func main() {
	flag.Parse()

	if *path == "" {
		log.Panic("a path to a service review compliance json dump is required")
	}

	f, err := os.Open(*path)
	if err != nil {
		log.Panic(err)
	}

	decoder := json.NewDecoder(f)

	var compliance []Compliance
	if err := decoder.Decode(&compliance); err != nil {
		log.Panic(err)
	}

	var fraudscores []FraudScoreClm

	for _, c := range compliance {
		fmt.Println(c)
		for _, event := range c.Events {
			score := 0.0
			created_str := ""
			var created_date time.Time
			done := false
			for k, v := range event {

				if k == "Created" {
					createdMap := v.(map[string]interface{})
					created_str = createdMap["$date"].(string)
					created_date, err = time.Parse(time.RFC3339, created_str)
					if err != nil {
						log.Panic(err)
					}
				}
				if k == "NewScore" {
					done = true
					var err error
					score = v.(float64)
					if err != nil {
						log.Panic(err)
					}
					fraudscores = append(fraudscores, FraudScoreClm{
						ReviewId:    c.ReviewId.Id,
						V2Score:     score,
						Created:     created_str,
						CreatedDate: created_date,
					})

					fmt.Println(fraudscores)
					return
				}
			}

			if done {
				break
			}
		}
	}
	// fmt.Println(fraudscores)
}

type FraudScoreClm struct {
	ReviewId    string
	V2Score     float64
	Created     string
	CreatedDate RedshiftDate
}

type RedshiftDate time.Time

func (d *RedshiftDate) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", d.Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}

type Compliance struct {
	ReviewId Id
	Events   []map[string]interface{}
}

type Id struct {
	Id string `json:"$oid"`
}

type Event struct {
	Type     string `json:"_t"`
	NewScore float32
	OldScore float32
	Created  RedshiftDate
}

type Date struct {
	Date time.Time `json:"$date"`
}
