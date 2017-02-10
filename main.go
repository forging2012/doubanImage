package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func GetUrl(url string) []byte {
	ret, err := http.Get(url)
	if err != nil {
		log.Println(url)
	}
	body := ret.Body
	data, _ := ioutil.ReadAll(body)
	return data
}

func getImage(image_url string) {
	data := GetUrl(image_url)
	body := string(data)
	part := regexp.MustCompile("https://(.*).doubanio.com/view/group_topic/large/public/(.*).jpg")
	match := part.FindAllString(body, -1)
	for _, value := range match {
		submit_url := "http://btlet.com/api/1/upload/?key=4127f561fb9a3dbc4204d298034382f4&source=" + url.QueryEscape(value)
		return_json := GetUrl(submit_url)
		res := make(map[string]interface{})
		json.Unmarshal(return_json, &res)
		log.Printf("%s -> %v \n", value, res["status_code"])
	}
}

func getGroupList(target_url string) {
	fmt.Printf("Begin Url : %s\n", target_url)
	doc, err := goquery.NewDocument(target_url)
	if err != nil {
		panic(err)
		log.Fatal(err)
	}
	// Find the review items
	doc.Find("td.title a").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		href, IsExist := s.Attr("href")
		if IsExist {
			getImage(href)
		}
	})
	wg.Done()
}

var wg sync.WaitGroup

func main() {
	endStartInt := flag.Int("e", 100, "End Start Int Value")
	defaultUrl := flag.String("u", "https://www.douban.com/group/meituikong/discussion?start=", "Group Url")
	flag.Parse()
	for i := 0; i < *endStartInt; i = i + 25 {
		wg.Add(1)
		go getGroupList(*defaultUrl + strconv.Itoa(i))
		time.Sleep(3e9)
	}
	wg.Wait()
}
