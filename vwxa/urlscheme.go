/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package vwxa

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/vogo/vogo/vlog"
)

const (
	generateURLSchemeURL = "https://api.weixin.qq.com/wxa/generatescheme?access_token="
)

// URLSchemeRequest represents the request parameters for generating URL Scheme.
type URLSchemeRequest struct {
	JumpWxa        *JumpWxa `json:"jump_wxa,omitempty"`        // 跳转到的目标小程序信息
	IsExpire       *bool    `json:"is_expire,omitempty"`       // 生成的scheme码类型，到期失效：true，永久有效：false
	ExpireType     *int     `json:"expire_type,omitempty"`     // 到期失效的scheme码失效类型，失效时间：0，失效间隔天数：1
	ExpireTime     *int64   `json:"expire_time,omitempty"`     // 到期失效的scheme码的失效时间，为Unix时间戳
	ExpireInterval *int     `json:"expire_interval,omitempty"` // 到期失效的scheme码的失效间隔天数
}

// JumpWxa represents the target Mini Program information for URL Scheme.
type JumpWxa struct {
	Path       string `json:"path,omitempty"`        // 通过scheme码进入的小程序页面路径
	Query      string `json:"query,omitempty"`       // 通过scheme码进入小程序时的query
	EnvVersion string `json:"env_version,omitempty"` // 要打开的小程序版本
}

// URLSchemeResponse represents the response from URL Scheme generation API.
type URLSchemeResponse struct {
	ErrCode  int    `json:"errcode"`
	ErrMsg   string `json:"errmsg"`
	OpenLink string `json:"openlink"`
}

// GenerateURLScheme generates a URL Scheme for WeChat Mini Program.
// 获取小程序scheme码，适用于短信、邮件、外部网页、微信内等拉起小程序的业务场景
func (c *Client) GenerateURLScheme(req *URLSchemeRequest) (*URLSchemeResponse, error) {
	accessToken, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := generateURLSchemeURL + accessToken

	// Set default env_version if not provided
	if req.JumpWxa != nil && req.JumpWxa.EnvVersion == "" {
		req.JumpWxa.EnvVersion = c.envVersion
	}

	jsonData, err := c.marshalURLSchemeRequest(req)
	if err != nil {
		return nil, err
	}

	vlog.Infof("generate url scheme request: %s", string(jsonData))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vlog.Errorf("failed to close response body: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	vlog.Infof("generate url scheme response: %s", string(body))

	var result URLSchemeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}

	return &result, nil
}

func (c *Client) marshalURLSchemeRequest(req *URLSchemeRequest) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if jsonErr := encoder.Encode(req); jsonErr != nil {
		return nil, jsonErr
	}
	// Remove the trailing newline added by Encode
	jsonData := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))

	return jsonData, nil
}

// GenerateSimpleURLScheme generates a simple URL Scheme with path and query.
func (c *Client) GenerateSimpleURLScheme(path, query string) (string, error) {
	isExpire := false
	req := &URLSchemeRequest{
		JumpWxa: &JumpWxa{
			Path:  path,
			Query: query,
		},
		IsExpire: &isExpire, // 永久有效
	}

	resp, err := c.GenerateURLScheme(req)
	if err != nil {
		return "", err
	}

	return resp.OpenLink, nil
}

// GenerateExpirableURLScheme generates a URL Scheme that expires at a specific time.
func (c *Client) GenerateExpirableURLScheme(path, query string, expireTime time.Time) (string, error) {
	isExpire := true
	expireType := 0
	expireUnix := expireTime.Unix()
	req := &URLSchemeRequest{
		JumpWxa: &JumpWxa{
			Path:  path,
			Query: query,
		},
		IsExpire:   &isExpire,
		ExpireType: &expireType, // 失效时间
		ExpireTime: &expireUnix,
	}

	resp, err := c.GenerateURLScheme(req)
	if err != nil {
		return "", err
	}

	return resp.OpenLink, nil
}

// GenerateIntervalURLScheme generates a URL Scheme that expires after a specific number of days.
func (c *Client) GenerateIntervalURLScheme(path, query string, expireIntervalDays int) (string, error) {
	isExpire := true
	expireType := 1
	req := &URLSchemeRequest{
		JumpWxa: &JumpWxa{
			Path:  path,
			Query: query,
		},
		IsExpire:       &isExpire,
		ExpireType:     &expireType, // 失效间隔天数
		ExpireInterval: &expireIntervalDays,
	}

	resp, err := c.GenerateURLScheme(req)
	if err != nil {
		return "", err
	}

	return resp.OpenLink, nil
}
