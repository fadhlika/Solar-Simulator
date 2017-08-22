package main

import (
	"log"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func periodScrap() {
	for {
		scrapAws()

		time.Sleep(60 * time.Second)
	}
}

func scrapAws() {
	shouldUpdate := true
	doc, err := goquery.NewDocument("https://www.elka.fi.itb.ac.id/aws")
	if err != nil {
		log.Println("Error goquery")
		return
	}

	units := make(map[string]float64)
	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		if i > 3 && i < 21 {
			name, _ := s.Attr("name")
			value, _ := s.Attr("value")
			if name == "inTemp" && value == "0" {
				shouldUpdate = false
				log.Panicln("Scrap data error")
			}
			units[name], _ = strconv.ParseFloat(value, 64)
		}
	})
	if shouldUpdate {
		data := Awsdata{
			0, time.Now(),
			units["inTemp"], units["inHumi"], units["AbsPress"], units["RelPress"],
			units["outTemp"], units["outHumi"], units["windir"], units["avgwind"],
			units["gustspeed"], units["solarrad"], units["uv"], units["uvi"],
			units["rainofhourly"], units["rainofdaily"], units["rainofweekly"],
			units["rainofmonthly"], units["rainofyearly"],
			false,
		}
		if err := data.save(); err != nil {
			log.Printf("error save aws data %v\n", err.Error())
			return
		}
	}
}