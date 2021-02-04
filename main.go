package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

var (
	start = "20201201"
	end   = "20210204"
)

func BytesToString(data []byte) string {
	return string(data[:])
}

func WriteAllData(name string, data string) {
	file, err := os.OpenFile(name, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString(data + "\r\n")
	if err != nil {
		panic(err)
	}
}

func GetEuckr(value string) string {

	var bufs bytes.Buffer
	wr := transform.NewWriter(&bufs, korean.EUCKR.NewDecoder())
	wr.Write([]byte(value))
	wr.Close()

	convVal := bufs.String()

	return convVal
}

func GetNextDate(today string) (int, string) {
	if today == end {
		return 0, ""
	}
	var yyyy int
	var mm int
	var dd int

	fmt.Sscanf(today, "%4d%02d%02d", &yyyy, &mm, &dd)

	t := time.Date(yyyy, time.Month(mm), dd+1, 0, 0, 0, 0, time.Local)

	next := fmt.Sprintf("%4d%02d%02d", t.Year(), t.Month(), t.Day())
	return 1, next
}

func main() {
	date := start
	for {
		for i := 1; ; i++ {
			fmt.Println(date)
			url := fmt.Sprintf("https://finance.naver.com/news/news_list.nhn?mode=LSS3D&section_id=101&section_id2=258&section_id3=401&date=%s&page=%d", date, i)
			resp, err := http.Get(url)
			if err != nil {
				panic(err)
			}

			defer resp.Body.Close()

			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				panic(err)
			}

			str := BytesToString(data)

			reg, err := regexp.Compile("\r\n\t|\n|\r\t|\t")
			if err != nil {
				panic(err)
			}

			replace := reg.ReplaceAllString(str, "")

			reg2, err := regexp.Compile("</dd>")
			if err != nil {
				panic(err)
			}

			replace2 := reg2.ReplaceAllString(replace, "</dd>\n")

			// WriteAllData("./all", replace2)

			reg3, err := regexp.Compile("(<dt class=\"articleSubject\").*(</dd>)")
			if err != nil {
				panic(err)
			}

			replace3 := reg3.FindAllString(replace2, 100)

			if len(replace3) == 0 {
				exit, err := regexp.Compile("(<dd class=\"articleSubject\").*(</dd>)")
				if err != nil {
					panic(err)
				}

				replace3 = exit.FindAllString(replace2, 100)
				if len(replace3) == 0 {
					break
				} else {
					continue
				}

			}

			// allname := fmt.Sprintf("./2021-02-04-%03d", i)
			// for i := 0; i < len(replace3); i++ {
			// 	WriteAllData(allname, replace3[i])
			// }

			reg4, err := regexp.Compile("(title=).*(</a>)")
			if err != nil {
				panic(err)
			}

			for i := 0; i < len(replace3); i++ {

				title := reg4.FindString(replace3[i])

				p := strings.Index(title, "title=")
				if p <= -1 {
					continue
				}
				value := title[p+len("title=")+1:]

				p = strings.Index(value, "\"")
				if p <= -1 {
					continue
				}

				value = value[:p]

				convVal := GetEuckr(value)

				fmt.Println(convVal)

				p = strings.Index(convVal, "[마감]프로그램")
				if p > -1 {
					name := "./[마감]프로그램"
					WriteAllData(name, date+"|"+convVal)
				}
				p = strings.Index(convVal, "[마감]코스피 개인")
				if p > -1 {
					name := "./[마감]코스피 개인"
					WriteAllData(name, date+"|"+convVal)
				}
				p = strings.Index(convVal, "[마감]코스피 기관")
				if p > -1 {
					name := "./[마감]코스피 기관"
					WriteAllData(name, date+"|"+convVal)
				}
				p = strings.Index(convVal, "[마감]코스피 외국인")
				if p > -1 {
					name := "./[마감]코스피 외국인"
					WriteAllData(name, date+"|"+convVal)
				}
				p = strings.Index(convVal, "[마감]코스피 하락..")
				if p > -1 {
					name := "./[마감]코스피"
					WriteAllData(name, date+"|"+convVal)
				}
				p = strings.Index(convVal, "[마감]코스피 상승..")
				if p > -1 {
					name := "./[마감]코스피"
					WriteAllData(name, date+"|"+convVal)
				}
			}
		}
		ret, nextdate := GetNextDate(date)
		if ret == 0 {
			break
		}
		date = nextdate
	}

}
