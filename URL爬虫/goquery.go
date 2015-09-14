package main

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	var File chan string = make(chan string, 10)
	var Num chan int32 = make(chan int32, 1)
	var i int32 = 0
	NFile, err := os.Create("url.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer NFile.Close()
	list := []string{"new_submit", "new_unclaimed", "new_public", "new_alarm", "new_unclaimed"}
	for _, v := range list {
		atomic.AddInt32(&i, 1)
		go geturl(v, File, &i, Num)
	}
	for {
		select {
		case s := <-File:
			NFile.WriteString(fmt.Sprintf("http://www.wooyun.org%s\n", s))
		case i := <-Num:
			if i == 0 {
				return
			}
		}
	}
}

func geturl(url string, File chan string, number *int32, Num chan int32) {
	var i int = 1
	for {
		doc, err := goquery.NewDocument(fmt.Sprintf("http://www.wooyun.org/bugs/%s/page/%d", url, i))
		if err != nil {
			fmt.Println(err)
			break
		}
		page := doc.Find(".listTable").Find("tbody").Find("td")
		if page.Length() == 0 {
			break
		}
		for i := 0; i < page.Length(); i++ {
			x := page.Eq(i).Find("a")
			s, b := x.Attr("href")
			s = fmt.Sprintf("%s\t%s", s, x.Text())
			if b {
				File <- s
			}
		}
		i++
	}
	x := atomic.AddInt32(number, -1)
	Num <- x
}
