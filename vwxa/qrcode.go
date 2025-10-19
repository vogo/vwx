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
	"io"
	"net/http"

	"github.com/vogo/vogo/vlog"
)

const (
	generateCodeUnlimitURL = "https://api.weixin.qq.com/wxa/getwxacodeunlimit?access_token="
)

// GenerateQRCode generates QR code for WeChat Mini Program with specified scene and page.
func (c *Client) GenerateQRCode(scene, page string) ([]byte, error) {
	accessToken, err := c.GetAccessToken()
	if err != nil {
		return nil, err
	}

	url := generateCodeUnlimitURL + accessToken

	params := map[string]interface{}{
		"scene":       scene,
		"page":        page,
		"check_path":  false,
		"env_version": c.envVersion,
	}

	jsonData, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vlog.Errorf("failed to close response body: %v", closeErr)
		}
	}()

	return io.ReadAll(resp.Body)
}
