package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Limit struct {
	number  int
	channel chan struct{}
}

func New(number int) *Limit {
	return &Limit{
		number:  number,
		channel: make(chan struct{}, number),
	}
}

func (limit *Limit) Run(f func()) {
	limit.channel <- struct{}{}
	go func() {
		f()
		<-limit.channel
	}()
}

var wg = sync.WaitGroup{}

const (
	concurrency = 5
)

func GetHttpReq(urlPath string, paramMap map[string]string, headerMap map[string]string) ([]byte, int) {
	Url, err := url.Parse(urlPath)
	if err != nil {
		return nil, 0
	}

	params := url.Values{}
	for k, v := range paramMap {
		params.Set(k, v)
	}

	//如果参数中有中文参数,这个方法会进行URLEncode
	Url.RawQuery = params.Encode()

	url2Http := Url.String()
	fmt.Println("Curl url:" + url2Http)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url2Http, nil)

	for k, v := range headerMap {
		req.Header.Set(k, v)
		// go host need set special
		if strings.EqualFold(k, "Host") {
			req.Host = v
		}
	}

	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
		return nil, 0
	}

	return body, 1
}

func main() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	start := time.Now()
	limit := New(concurrency)

	apiUrl := "http://xxxxx"
	max := 100
	params := make(map[string]string)
	params["cmd"] = "104"
	params["hash"] = "F3F574A2E1D7A5367A52C71934BB19D8"
	params["pid"] = "ss"
	params["backupdomain"] = "1"
	params["ssl"] = "0"
	params["jump"] = "0"
	params["appid"] = "1012"

	headers := make(map[string]string)
	headers["Host"] = "trackermv.kugou.com"

	for i := 0; i < max; i++ {
		wg.Add(1)

		value := i

		goFunc := func() {
			fmt.Printf("stat func:%d\n", value)
			data, status := GetHttpReq(apiUrl, params, headers)
			if status != 1 {
				fmt.Println("..................")
			} else {
				fmt.Println(string(data))
			}
			wg.Done()
		}

		limit.Run(goFunc)
	}

	wg.Wait()
	fmt.Printf("耗时: %fs", time.Now().Sub(start).Seconds())
}
