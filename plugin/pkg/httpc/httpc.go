package httpc

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ca17/datahub/plugin/pkg/common"
)

const (
	charsetUTF8                    = "charset=UTF-8"
	MIMEApplicationJSON            = "application/json"
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + charsetUTF8
	MIMETextXML                    = "text/xml"
	MIMETextXMLCharsetUTF8         = MIMETextXML + "; " + charsetUTF8
	MIMETextHTML                   = "text/html"
	MIMETextHTMLCharsetUTF8        = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlain                  = "text/plain"
	MIMETextPlainCharsetUTF8       = MIMETextPlain + "; " + charsetUTF8
	MIMEMultipartForm              = "multipart/form-data"
	MIMEOctetStream                = "application/octet-stream"

	HeaderContentType = "Content-Type"
)

type H map[string]string

func Get(url string, header H) (respBytes []byte, err error) {
	return DoRestfulRequest(http.MethodGet, url, nil, header)
}

func Post(url string, body io.Reader, header H) (respBytes []byte, err error) {
	return DoRestfulRequest(http.MethodPost, url, body, header)
}

func PostJson(url string, data interface{}) (respBytes []byte, err error) {
	body := common.ToJson(data)
	rd := bytes.NewReader([]byte(body))
	return DoRestfulRequest(http.MethodPost, url, rd, map[string]string{
		HeaderContentType: MIMEApplicationJSON,
		"Connection":      "keep-alive",
	})
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
