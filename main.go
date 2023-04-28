package main

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"io"
	"julian-req-stat/config"
	dbjulian "julian-req-stat/db"
	"julian-req-stat/helper"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var webSite = "https://marpla.ru/test.php?art="

type MarplaResponse struct {
	Search    string            `json:"q"`
	Total     string            `json:"total"`
	PopularWB string            `json:"popular_wb"`
	Dates     map[string]string `json:"-"`
}

func main() {
	db, err := sqlx.Connect("postgres", config.DBConnection)
	if err != nil {
		log.Fatalln(err)
	}

	// Create a channel to signal when to stop the service
	stopChan := make(chan bool)

	// Start a goroutine to execute the method every 30 minutes
	go func() {
		for {
			// Execute the method
			process(db)

			// Wait for 30 minutes before executing the method again
			select {
			case <-time.After(30 * time.Minute):
			case <-stopChan:
				return
			}
		}
	}()

	// Wait for a signal to stop the service
	<-stopChan
}

func process(db *sqlx.DB) {
	log.Println("started process")
	articles := dbjulian.GetArticles(db)
	mu := sync.Mutex{}
	timeNow := time.Now()
	wg := sync.WaitGroup{}

	res := make([]dbjulian.RequestStat, 0)
	requestsStat, err := dbjulian.GetCheckRequestStatSlice(db)
	if err != nil {
		log.Fatal(err)
	}

	for _, article := range articles {
		wg.Add(1)
		go func(article dbjulian.Article, mu *sync.Mutex, timeNow time.Time, requestsStat map[string]bool) {
			defer wg.Done()
			url := webSite + article.Nomenclature
			response, err := http.Get(url)
			if err != nil {
				log.Fatal(err)
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}

			html := string(body)

			regex := regexp.MustCompile(`data = \[[\S\s]*\//`)
			found := regex.FindAllString(html, -1)

			newRegex := regexp.MustCompile(`data = \[\n[\S\s]*(data_cats = \[)`)
			parse := newRegex.FindAllString(found[0], -1)

			arr := strings.Split(parse[0], "\n")
			startTag := helper.GetStringBetween(html, "<a", "</a>")

			href := helper.GetStringBetween(startTag, "href=", `"`)
			href = href[strings.Index(href, `"`)+1 : strings.LastIndex(href, `"`)]

			src := helper.GetStringBetween(startTag, "src=", "' >")
			src = src[strings.Index(src, `'`)+1 : strings.LastIndex(src, `'`)]

			name := helper.GetStringBetween(startTag, "<h4 style='margin-top: 0px;'>", "<span")
			name = name[strings.Index(name, ">")+1 : strings.LastIndex(name, "<")]

			normalizedArr := []string{}

			for _, article := range arr {
				if strings.HasPrefix(article, "{q:") {
					normalizedArr = append(normalizedArr, article)
				}
			}

			for i := range normalizedArr {
				var record MarplaResponse

				data := helper.RepairJson(normalizedArr[i])

				if err := json.Unmarshal([]byte(data), &record.Dates); err != nil {
					log.Println("Error:", err)
					return
				}

				if err := json.Unmarshal([]byte(data), &record); err != nil {
					log.Println("Error:", err)
					return
				}

				for date, searchPlaceRaw := range record.Dates {
					if !strings.Contains(date, "dt_") {
						continue
					}

					dateString := strings.TrimLeft(date, "dt_")

					dateString = dateString + "." + fmt.Sprintf("%v", timeNow.Year())
					dateTime, err := helper.ParseDate(dateString)
					if err != nil {
						log.Fatalln(err)
					}

					checkDateTime := dateTime.Format("2006-01-02")
					checkArticleId := strconv.FormatInt(article.ID, 10)

					if ok := requestsStat[checkArticleId+checkDateTime+record.Search]; ok {
						continue
					}

					searchPlace := helper.ExtractBetweenTags(searchPlaceRaw)

					if searchPlace == `<img src="https://v1.iconsearch.ru/uploads/icons/fatcow/32x32/accept.png" data-toggle="tooltip" style="height: 13px;" data-mytooltip="Найдено где-то в индексе, возможно на оч. дальней странице" data-original-title="" title="" height="16px">` {
						searchPlace = "*"
					}

					res = append(res, dbjulian.RequestStat{
						ArticleID:   article.ID,
						Name:        record.Search,
						Results:     record.Total,
						FrequencyWB: record.PopularWB,
						SearchPlace: searchPlace,
						Date:        dateTime,
						CreatedAt:   time.Now(),
						UpdatedAt:   time.Now(),
					})
				}
			}
		}(article, &mu, timeNow, requestsStat)
	}
	wg.Wait()

	portions := len(res) / 6553
	start := 0

	for i := 1; i <= portions; i++ {
		_, err = dbjulian.InsertRequestStat(db, res[start:6552*i])
		if err != nil {
			log.Println(err)
			return
		}

		start = 6552*i + 1
	}

	log.Println("ended process")
}
