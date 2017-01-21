package main

import (
	"fmt"
	"go/httputils"
)

func main() {
	res, err := httputils.HttpGet("http://appv2.lygou.cc/chat/square?user_id=9&group_id=200072131477984441&select_sex=0&page=1&name=54861391&query_user_id=3081386&code=120110&identity=10000002777&password=1CAF8236DFA3634A&tag_id=2", nil)
	fmt.Printf("result: %+v, error: %+v \n", res, err)

	res, err = httputils.HttpPost("http://120.26.48.229:9081/usersvc/GetUserinfo", "{\"id\":97}", nil)
	fmt.Printf("result: %+v, error: %+v \n", res, err)
}
