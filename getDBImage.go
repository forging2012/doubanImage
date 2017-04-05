package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
)

func GetUrl(url string) ([]byte, error) {
	ret, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	body := ret.Body
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return data, err
}

func getImage(image_url string, k string) {
	doc, err := goquery.NewDocument(image_url)
	if err != nil {
		log.Println(err)
		return
	}

	// Find the Image items
	doc.Find("#link-report img").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		value, IsExist := s.Attr("src")
		if IsExist {

			match := strings.Split(value, "/public/")

			var resNum int

			err = db.QueryRow(`SELECT COUNT(*) FROM chv_images WHERE image_original_filename=?`, match[1]).Scan(&resNum)
			if err != nil {
				log.Fatalf("Error on select database connection: %s", err.Error())
			}

			if resNum == 0 {
				submit_url := "http://788to.com/api/1/upload/?key=" + k + "&source=" + url.QueryEscape(value)
				return_json, err := GetUrl(submit_url)
				if err != nil {
					log.Println(err)
				}
				res := make(map[string]interface{})
				err = json.Unmarshal(return_json, &res)
				if err != nil {
					log.Println(err)
					return
				}
				if res["status_code"].(float64) == 200 {
					log.Printf("%s -> %v \n", value, res["status_code"])
				} else {
					myError := res["error"].(map[string]interface{})
					log.Printf("%s -> %v -> %v \n", value, res["status_code"], myError["message"])
				}
			} else {
				log.Println(value + " -> Skip the same image.")
			}

			time.Sleep(3 * time.Second)
		} else {
			log.Println("Image url error : " + image_url)
			return
		}
	})
}

func getGroupList(target_url string, k string) {
	fmt.Printf("Begin url : %s\n", target_url)
	doc, err := goquery.NewDocument(target_url)
	if err != nil {
		panic(err)
	}

	// Find the items
	doc.Find("td.title a").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		href, IsExist := s.Attr("href")
		if IsExist {
			getImage(href, k)
		} else {
			log.Println("List href null : " + target_url)
		}
	})
	wg.Done()
}

var wg sync.WaitGroup
var db *sql.DB // global variable to share it between main and the HTTP handler
var err error  // global the error

func main() {
	k := flag.String("k", "laoji.org", "Chevereto Key")
	endStartInt := flag.Int("e", 100, "End Start Int Value")
	defaultUrl := flag.String("u", "https://www.douban.com/group/meituikong/discussion?start=", "Group Url")
	db, err = sql.Open("mysql", "img-btlet-select:laoji.org@tcp(v.laoji.org:3306)/img-btlet?charset=utf8")
	if err != nil {
		log.Fatalf("Error on initializing database connection: %s", err.Error())
	}
	defer db.Close()

	flag.Parse()
	for i := 0; i < *endStartInt; i = i + 25 {
		wg.Add(1)
		go getGroupList(*defaultUrl+strconv.Itoa(i), *k)
	}
	wg.Wait()
}
