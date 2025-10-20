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
	generateURLLinkURL = "https://api.weixin.qq.com/wxa/generate_urllink?access_token="
)

// URLLinkRequest represents the request parameters for generating URL Link.
type URLLinkRequest struct {
	Path           *string    `json:"path,omitempty"`            // 小程序页面路径
	Query          *string    `json:"query,omitempty"`           // 小程序页面query参数
	ExpireType     *int       `json:"expire_type,omitempty"`     // 失效类型：0-到期失效，1-失效间隔天数
	ExpireTime     *int64     `json:"expire_time,omitempty"`     // 到期失效的Unix时间戳
	ExpireInterval *int       `json:"expire_interval,omitempty"` // 失效间隔天数
	CloudBase      *CloudBase `json:"cloud_base,omitempty"`      // 云开发静态网站配置
	EnvVersion     *string    `json:"env_version,omitempty"`     // 小程序版本
}

// CloudBase represents the cloud development static website configuration.
type CloudBase struct {
	Env           string `json:"env"`                      // 云开发环境
	Domain        string `json:"domain,omitempty"`         // 静态网站自定义域名
	Path          string `json:"path,omitempty"`           // H5页面路径
	Query         string `json:"query,omitempty"`          // H5页面query参数
	ResourceAppID string `json:"resource_appid,omitempty"` // 第三方批量代云开发时必填
}

// URLLinkResponse represents the response from URL Link generation API.
type URLLinkResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	URLLink string `json:"url_link"`
}

func (c *Client) marshalRequest(req *URLLinkRequest) ([]byte, error) {
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

// GenerateURLLink generates a URL Link for WeChat Mini Program.
// 获取小程序 URL Link，适用于短信、邮件、网页、微信内等拉起小程序的业务场景
func (c *Client) GenerateURLLink(req *URLLinkRequest) (*URLLinkResponse, error) {
	accessToken, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := generateURLLinkURL + accessToken

	// Set default env_version if not provided
	if req.EnvVersion == nil {
		envVersion := c.envVersion
		req.EnvVersion = &envVersion
	}

	jsonData, err := c.marshalRequest(req)
	if err != nil {
		return nil, err
	}

	vlog.Infof("generate urllink request: %s", string(jsonData))

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

	vlog.Infof("generate urllink response: %s", string(body))

	var result URLLinkResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}

	return &result, nil
}

// GenerateSimpleURLLink generates a simple URL Link with basic parameters.
// 简化版本的URL Link生成，只需要提供基本参数
func (c *Client) GenerateSimpleURLLink(path, query string) (string, error) {
	req := &URLLinkRequest{
		Path:  &path,
		Query: &query,
	}

	resp, err := c.GenerateURLLink(req)
	if err != nil {
		return "", err
	}

	return resp.URLLink, nil
}

// GenerateExpirableURLLink generates a URL Link with expiration time.
// 生成带有过期时间的URL Link
func (c *Client) GenerateExpirableURLLink(path, query string, expireTime time.Time) (string, error) {
	expireType := 0
	expireTimeUnix := expireTime.Unix()
	req := &URLLinkRequest{
		Path:       &path,
		Query:      &query,
		ExpireType: &expireType, // 到期失效
		ExpireTime: &expireTimeUnix,
	}

	resp, err := c.GenerateURLLink(req)
	if err != nil {
		return "", err
	}

	return resp.URLLink, nil
}

// GenerateIntervalURLLink generates a URL Link with expiration interval in days.
// 生成带有失效间隔天数的URL Link
func (c *Client) GenerateIntervalURLLink(path, query string, expireIntervalDays int) (string, error) {
	expireType := 1
	req := &URLLinkRequest{
		Path:           &path,
		Query:          &query,
		ExpireType:     &expireType, // 失效间隔天数
		ExpireInterval: &expireIntervalDays,
	}

	resp, err := c.GenerateURLLink(req)
	if err != nil {
		return "", err
	}

	return resp.URLLink, nil
}
