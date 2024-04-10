package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// Request 发送请求
func Request(methodUrl string, method string, dataBody string, dataHeader map[string]string) ([]byte, error) {
	jsonStr := []byte(dataBody)
	//1.获取请求
	client := &http.Client{}
	apiUrl := methodUrl
	request, err := http.NewRequest(method, apiUrl, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		return nil, err
	}
	//2.填充header
	for key, value := range dataHeader {
		request.Header.Set(key, value)
	}
	//3.发送请求
	resp, err := client.Do(request) //发送请求
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		return nil, err
	}
	defer resp.Body.Close() //一定要关闭resp.Body
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
		return nil, err
	}
	return content, nil
}
