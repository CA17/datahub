package httpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

type H map[string]string

func Get(url string, header H) (respBytes []byte, err error) {
	return DoRestfulRequest(http.MethodGet, url, nil, header)
}

func Post(url string, body io.Reader, header H) (respBytes []byte, err error) {
	return DoRestfulRequest(http.MethodPost, url, body, header)
}

func DoRestfulRequest(method, url string, body io.Reader, header map[string]string) (respBytes []byte, err error) {
	// 设置超时
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	// 初始化客户端请求对象
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return
	}
	// 添加自定义请求头
	if header != nil {
		for key, value := range header {
			req.Header.Add(key, value)
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	buffer := bytes.Buffer{}
	io.Copy(&buffer, resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("http response status is %d for url  %s", resp.StatusCode, url)
	}
	return buffer.Bytes(), nil
}
