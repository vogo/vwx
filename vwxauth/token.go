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

package vwxauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vogo/vogo/vlog"
)

const (
	accessTokenURL = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
)

func (c *Service) cacheKeyAccessToken() string {
	return c.client.CacheKeyPrefix + "vwxa:access_token:" + c.client.AppID
}

// GetAccessToken retrieves access token from WeChat API with caching support.
func (c *Service) GetAccessToken() (string, error) {
	if c.client.CacheProvider != nil {
		cachedToken := c.client.CacheProvider.Get(context.Background(), c.cacheKeyAccessToken())
		if cachedToken != "" {
			return cachedToken, nil
		}
	}

	url := fmt.Sprintf(accessTokenURL, c.client.AppID, c.client.AppSecret)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vlog.Errorf("failed to close response body | err: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.ErrCode != 0 {
		return "", errors.New(result.ErrMsg)
	}

	// cache access token
	if c.client.CacheProvider != nil {
		expireTime := time.Duration(result.ExpiresIn-300) * time.Second
		if err := c.client.CacheProvider.Set(context.Background(),
			c.cacheKeyAccessToken(), result.AccessToken, expireTime); err != nil {
			vlog.Errorf("failed to set access token to cache | err: %v", err)
		}
	}

	return result.AccessToken, nil
}
