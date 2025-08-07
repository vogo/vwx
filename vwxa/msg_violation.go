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
	"fmt"
	"io"
	"net/http"

	"github.com/vogo/vogo/vlog"
)

const (
	msgSecCheckURL = "https://api.weixin.qq.com/wxa/msg_sec_check?access_token=%s"
)

// MsgViolationCheckRequest represents a request for message security check.
type MsgViolationCheckRequest struct {
	Content string `json:"content"` // 要检测的文本内容，长度不超过 500KB
}

// MsgViolationCheckResponse represents the response from message security check.
type MsgViolationCheckResponse struct {
	ErrCode int    `json:"errcode"` // 错误码
	ErrMsg  string `json:"errmsg"`  // 错误信息
}

// MsgViolationCheck detects whether text content contains illegal or non-compliant content.
// Application scenarios:
// - User profile illegal text detection
// - Media news user article and comment content detection
// - Game user uploaded material detection, etc.
// Rate limit: single appId call limit is 4000 times/minute, 2,000,000 times/day
func (c *Client) MsgViolationCheck(content string) (*MsgViolationCheckResponse, error) {
	accessToken, err := c.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get access token error: %v", err)
	}

	url := fmt.Sprintf(msgSecCheckURL, accessToken)

	request := &MsgViolationCheckRequest{
		Content: content,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request error: %v", err)
	}

	vlog.Infof("msg sec check request: %s", string(data))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("send request error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %v", err)
	}

	vlog.Infof("msg sec check response: %s", string(body))

	var response MsgViolationCheckResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %v", err)
	}

	// 根据微信文档，errcode为0表示内容正常，87014表示内容可能潜在风险
	if response.ErrCode != 0 && response.ErrCode != 87014 {
		return &response, errors.New(response.ErrMsg)
	}

	return &response, nil
}

// IsMsgContentSafe is a convenient method to check if content is safe.
// Returns true if content is safe, false if content may have risks.
func (c *Client) IsMsgContentSafe(content string) (bool, error) {
	response, err := c.MsgViolationCheck(content)
	if err != nil {
		return false, err
	}

	// errcode为0表示内容正常
	return response.ErrCode == 0, nil
}
