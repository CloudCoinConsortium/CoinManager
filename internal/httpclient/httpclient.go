package httpclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/config"

	"github.com/CloudCoinConsortium/superraidaclientbackend/internal/logger"
)

const ERR_TIMEOUT = 1
const ERR_NET = 2
const ERR_URL = 3
const ERR_COMMON = 4

type HClient struct {
	baseUrl string
	timeout int
	index   int
}

type Error struct {
	Code    int
	Message string
}

type Result struct {
}

func New(url string, index int) *HClient {
	return &HClient{
		baseUrl: url,
		timeout: config.HTTP_TIMEOUT,
		index:   index,
	}
}

func (c *HClient) log(ctx context.Context, message string) {
	prefix := "RAIDA" + strconv.Itoa(c.index)
	logger.L(ctx).Debugf(prefix + " " + message)
}

func (c *HClient) Send(ctx context.Context, nurl string, params map[string]string, doneIssued chan bool, post bool) (string, *Error) {
	sendURL := fmt.Sprintf("%s%s", c.baseUrl, nurl)

	if post {
		c.log(ctx, "POST " + sendURL)
	} else {
		c.log(ctx, "GET " + sendURL)
	}

	var Request string
	URLData := url.Values{}
	for key, element := range params {
		logger.L(ctx).Debug(key + ":" + element)

		if strings.Contains(key, "[]") {
			var ba []string
			logger.L(ctx).Debug("p=" + element)
			err := json.Unmarshal([]byte(element), &ba)
			if err != nil {
				logger.L(ctx).Errorf("Failed to exract bytes from URL parameter: %s", element)
				return "", &Error{
					Code:    ERR_COMMON,
					Message: "Internal Error",
				}
			}
			for _, p := range ba {
				URLData.Add(key, p)
			}
		} else {
			URLData.Set(key, element)
		}
	}

	if post {

	} else {
		u, _ := url.Parse(sendURL)
		u.RawQuery = URLData.Encode()
		Request = fmt.Sprintf("%v", u)
	}

  c.log(ctx, Request)

	body := ""
	var raidahttp = &http.Client{
		Timeout: time.Duration(c.timeout) * time.Second,
	}

	//send request
	beforeSeconds := time.Now()

	logger.L(ctx).Debug(Request)
	if doneIssued != nil {
		logger.L(ctx).Debug("Doing async request")
		go func() {
			raidahttp.Get(Request)
		}()

		doneIssued <- true
		return "", nil
	}

	var response *http.Response
	var err error
	if post {
		response, err = raidahttp.PostForm(sendURL, URLData)
	} else {
		response, err = raidahttp.Get(Request)
	}
	elapsedSeconds := time.Since(beforeSeconds).Nanoseconds() / 1000000

	c.log(ctx, "Total time: " + strconv.Itoa(int(elapsedSeconds)) + " ms")

	if err != nil {
		rerr := &Error{}
		switch err := err.(type) {
		case net.Error:
			if err.Timeout() {
				rerr.Code = ERR_TIMEOUT
				rerr.Message = "Network Timeout"
			} else {
				rerr.Code = ERR_NET
				rerr.Message = "Network Error, " + err.Error()
			}
		case *url.Error:
			fmt.Println("url error")
			if err, ok := err.Err.(net.Error); ok && err.Timeout() {
				rerr.Code = ERR_TIMEOUT
				rerr.Message = "Network URL Timeout"
			} else {
				rerr.Code = ERR_URL
				rerr.Message = "URL Error, " + err.Error()
			}
		default:
			rerr.Code = ERR_COMMON
			rerr.Message = err.Error()
		}

		return "", rerr
	}

	bodybytes, _ := ioutil.ReadAll(response.Body)
	body = string(bodybytes)

	//c.log(body)

	//fmt.Printf("err=%v %v\n",err, err.Error())
	return body, nil

}
