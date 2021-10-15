/* 
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */


package util

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"k8s.io/klog"
	"net/http"
	"net/url"
	"time"
)

type HttpClient struct {
	Host    string
	Timeout time.Duration
}

// Post
/**
 * @Title:  发送Post请求
 **/
func (request *HttpClient) Post(path string, header map[string]string, body interface{}) (res *http.Response, err error) {
	req, err := request.buildRequest(context.Background(), "POST", path, header, nil, body)
	httpClient := http.Client{Timeout: request.Timeout}
	res, err = httpClient.Do(req)
	if err != nil {
		klog.Errorf("util.Post upload event error: %v", err)
		return
	}
	if res.StatusCode != http.StatusOK {
		klog.Errorf("util.Post upload event status code: %d", res.StatusCode)
	}

	return
}

// HttpsPost
/**
 * @Title:  发送 Https Post 请求
 **/
func (request *HttpClient) HttpsPost(path string, header map[string]string, body interface{}) (res *http.Response, err error) {
	req, err := request.buildRequest(context.Background(), "POST", path, header, nil, body)
	// 跳过签名证书验证
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	httpClient := http.Client{Timeout: request.Timeout, Transport: tr}
	res, err = httpClient.Do(req)
	if err != nil {
		klog.Errorf("util.HttpsPost upload error: %v", err)
		return
	}
	if res.StatusCode != http.StatusOK {
		klog.Errorf("util.HttpsPost upload status code: %d", res.StatusCode)
	}

	return
}

// buildRequest
/**
 * @Title:  buildRequest
 **/
func (request *HttpClient) buildRequest(ctx context.Context, method, path string, header map[string]string, params map[string][]string, body interface{}) (*http.Request, error) {
	u, err := url.Parse(request.Host)
	if err != nil {
		return nil, fmt.Errorf("parse host %s error: %s ", request.Host, err.Error())
	}
	u.Path = path
	if params != nil {
		query := url.Values{}
		for k, values := range params {
			for _, v := range values {
				query.Add(k, v)
			}
		}
		u.RawQuery = query.Encode()
	}

	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode requst body error: %s ", err.Error())
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, fmt.Errorf("build request error: %s", err.Error())
	}

	for k, v := range header {
		req.Header.Add(k, v)
	}
	return req.WithContext(ctx), nil
}
