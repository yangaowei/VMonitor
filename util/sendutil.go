package util

import (
	//"bytes"
	//"encoding/json"
	"fmt"
	//"log"
	"net/http"
	"net/url"
	"strconv"
	//"os"
	//"time"
)

func SendData(timestamp int64, data string) {

	url_host := "http://10.151.30.72:20099/api/log"
	tmp := strconv.FormatInt(timestamp, 10)
	resp, err := http.PostForm(url_host,
		url.Values{"timestamp": {tmp}, "data": {data}})

	if err != nil {
		fmt.Println(err)
		//return
	} else {
		fmt.Println(resp)
	}

}
