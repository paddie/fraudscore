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
		done := false
		fraudscore := FraudScoreClm{
			ReviewId: c.ReviewId.Id,
		}
		for _, event := range c.Events {
			for k, v := range event {

				if k == "Created" {
					createdMap := v.(map[string]interface{})
					created_str := createdMap["$date"].(string)
					date, err := time.Parse(time.RFC3339, created_str)
					if err != nil {
						log.Panic(err)
					}
					fraudscore.Created = RedshiftDate(date)
				}
				if k == "NewScore" {
					done = true
					fraudscore.V2Score = v.(float64)
				}
			}
		}

		if done {
			fraudscores = append(fraudscores, fraudscore)
		}
	}

	fmt.Println(len(fraudscores))

	file, err := os.Create("out.csv")
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	for _, fraudscore := range fraudscores {
		file.WriteString(fmt.Sprintf("%s,%1.4f,%s,V2\n", fraudscore.ReviewId, fraudscore.V2Score, fraudscore.Created))
	}
}

type FraudScoreClm struct {
	ReviewId string
	V2Score  float64
	Created  RedshiftDate
}

type RedshiftDate time.Time

func (d RedshiftDate) String() string {
	return fmt.Sprintf("%s", time.Time(d).Format("2006-01-02 15:04:05"))
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
