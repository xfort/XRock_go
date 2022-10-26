package net

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type RockHttp struct {
	client *http.Client
}

var HttpClient = NewRockHttpClient()

const TAG = "RockHttp"

func NewRockHttpClient() *RockHttp {
	transport := &http.Transport{
		//Proxy: http.ProxyFromEnvironment,
		DialContext: defaultTransportDialContext(&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}),
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	//transportH3 := &http3.RoundTripper{}

	rockHttp := RockHttp{
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}
	return &rockHttp
}

func defaultTransportDialContext(dialer *net.Dialer) func(context.Context, string, string) (net.Conn, error) {
	return dialer.DialContext
}

func (rockhttp *RockHttp) Do(ctx context.Context, url string, method string, header *http.Header, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if header != nil {
		for key, _ := range *header {
			req.Header.Set(key, header.Get(key))
		}
	}
	return rockhttp.DoRequest(req)
}

// 读取response.Body 的bytes
func (rockhttp *RockHttp) DoResBytes(req *http.Request) ([]byte, error) {
	resp, err := rockhttp.DoRequest(req)
	defer func() {
		if resp != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		return nil, err
	}
	respBytes, err := io.ReadAll(resp.Body)

	log.Println(TAG, req.Method, req.URL.String())

	if req.Header != nil {
		for key, value := range req.Header {
			log.Println(TAG, "request header", key, fmt.Sprintf("%v", value))
		}
	}
	if resp != nil {
		log.Println(TAG, "Response", resp.Status)
		if resp.Header != nil {
			for key, value := range resp.Header {
				log.Println(TAG, "response header", key, fmt.Sprintf("%v", value))
			}
		}
	}
	if respBytes != nil {
		//log.Println("RockHttp", "response ", string(respBytes))
	}
	return respBytes, err
}

// 结果json格式的bytes 转为 指定对象
func (rockhttp *RockHttp) DoResJson(req *http.Request, obj any) error {
	resBytes, err := rockhttp.DoResBytes(req)
	if err != nil {
		return err
	}
	err = json.Unmarshal(resBytes, obj)
	return err
}

// 结果proto格式的bytes 转为 proto-Message
func (rockhttp *RockHttp) DoResProtoMessage(req *http.Request, message *proto.Message) error {
	resBytes, err := rockhttp.DoResBytes(req)
	if err != nil {
		return err
	}
	err = proto.Unmarshal(resBytes, *message)
	return err
}

// json格式的bytes 转为 proto-Message
func (rockhttp *RockHttp) DoResProtoJson(req *http.Request, message proto.Message) error {
	resBytes, err := rockhttp.DoResBytes(req)
	if err != nil {
		return err
	}
	err = protojson.UnmarshalOptions{
		AllowPartial:   true,
		DiscardUnknown: true,
	}.Unmarshal(resBytes, message)
	return err
}

func (rockhttp *RockHttp) DoRequest(req *http.Request) (*http.Response, error) {
	resp, err := rockhttp.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (rockhttp *RockHttp) Get(ctx context.Context, url string, header *http.Header) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	if header != nil {
		req.Header = *header
	}
	return rockhttp.DoResBytes(req)
}

func (rockhttp *RockHttp) PostForm(ctx context.Context, url string, data url.Values, headerArg *http.Header) ([]byte, error) {
	//"application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	body := strings.NewReader(data.Encode())
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return nil, err
	}
	if headerArg != nil {
		for key, _ := range *headerArg {
			req.Header.Set(key, headerArg.Get(key))
		}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return rockhttp.DoResBytes(req)
}
func (rockhttp *RockHttp) GetResponse(ctx context.Context, url string, header *http.Header) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	if header != nil {
		for key, _ := range *header {
			req.Header.Set(key, header.Get(key))
		}
	}
	return rockhttp.DoRequest(req)
}

// 上传文件
func (rockhttp *RockHttp) UploadFile(ctx context.Context, url string, method string, header *http.Header, fileFieldName, itemFilepath, fileName string) (*http.Response, error) {
	req, err := rockhttp.createUploadRequest(ctx, url, method, header, fileFieldName, itemFilepath, fileName)
	if err != nil {
		return nil, err
	}
	return rockhttp.DoRequest(req)
}

// 新建上传文件的 request
func (rockhttp *RockHttp) createUploadRequest(ctx context.Context, url string, method string, header *http.Header, fileFieldName, itemFilepath, fileName string) (*http.Request, error) {
	pReader, pWriter := io.Pipe()
	partWriter := multipart.NewWriter(pWriter)
	go func() {
		defer pWriter.Close()
		defer partWriter.Close()
		formWriter, err := partWriter.CreateFormFile(fileFieldName, fileName)
		if err != nil {
			log.Println("UploadFile() CreateFormFile error " + err.Error())
			return
		}
		file, err := os.Open(itemFilepath)
		if err != nil {
			log.Println("UploadFile() openFile error "+err.Error(), itemFilepath)
			return
		}
		defer file.Close()
		if _, err := io.Copy(formWriter, file); err != nil {
			log.Println("UploadFile() Copy error "+err.Error(), itemFilepath)
			return
		}
	}()

	req, err := http.NewRequestWithContext(ctx, method, url, pReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", partWriter.FormDataContentType())
	if header != nil {
		for key, _ := range *header {
			req.Header.Set(key, header.Get(key))
		}
	}
	return req, nil
}

// 上传文件，结果转为 bytes
func (rockhttp *RockHttp) UploadFileResBytes(ctx context.Context, url string, method string, header *http.Header, fileFieldName, itemFilepath, fileName string) ([]byte, error) {
	req, err := rockhttp.createUploadRequest(ctx, url, method, header, fileFieldName, itemFilepath, fileName)
	if err != nil {
		return nil, err
	}
	return rockhttp.DoResBytes(req)
}

//
//func (*RockHttp) PostJson(r http.Request) (*http.Response, error) {
//	return nil, nil
//}
//
//type RequestOptions interface {
//	option(r *http.Request)
//}
//
//type ResponseOptions interface {
//	option(resp *http.Response)
//}
