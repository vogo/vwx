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
	subscribeMessageSendURL = "https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s"
)

// SubscribeMessageDataItem represents a data item in a subscribe message.
type SubscribeMessageDataItem struct {
	Value string `json:"value"`
}

// SubscribeMessageRequest represents a request to send a subscribe message.
type SubscribeMessageRequest struct {
	ToUser           string                               `json:"touser"`                      // 接收者（用户）的 openid
	TemplateID       string                               `json:"template_id"`                 // 所需下发的订阅消息的模板id
	Page             string                               `json:"page,omitempty"`              // 点击模板卡片后的跳转页面，仅限本小程序内的页面
	Data             map[string]*SubscribeMessageDataItem `json:"data"`                        // 模板内容
	MiniProgramState string                               `json:"miniprogram_state,omitempty"` // 跳转小程序类型：developer为开发版；trial为体验版；formal为正式版；默认为正式版
	Lang             string                               `json:"lang,omitempty"`              // 进入小程序查看的语言类型，支持zh_CN(简体中文)、en_US(英文)、zh_HK(繁体中文)、zh_TW(繁体中文)，默认为zh_CN
}

// SubscribeMessageResponse represents the response from sending a subscribe message.
type SubscribeMessageResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// SendSubscribeMessage sends a subscribe message to the specified user.
func (c *Service) SendSubscribeMessage(request *SubscribeMessageRequest) (*SubscribeMessageResponse, error) {
	accessToken, err := c.authSvc.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get access token error: %v", err)
	}

	url := fmt.Sprintf(subscribeMessageSendURL, accessToken)

	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request error: %v", err)
	}

	vlog.Infof("send subscribe message | req: %s", string(data))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("send request error: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vlog.Errorf("failed to close response body | err: %v", closeErr)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response error: %v", err)
	}

	vlog.Infof("send subscribe message | resp: %s", string(body))

	var response SubscribeMessageResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %v", err)
	}

	if response.ErrCode != 0 {
		return &response, errors.New(response.ErrMsg)
	}

	return &response, nil
}

// SendSubscribeMessageSimple is a convenient method to send a subscribe message with simple data.
func (c *Service) SendSubscribeMessageSimple(openID, templateID, page string, data map[string]string) (*SubscribeMessageResponse, error) {
	// 构建数据项
	dataItems := make(map[string]*SubscribeMessageDataItem)
	for k, v := range data {
		dataItems[k] = &SubscribeMessageDataItem{Value: v}
	}

	// 构建请求
	request := &SubscribeMessageRequest{
		ToUser:     openID,
		TemplateID: templateID,
		Page:       page,
		Data:       dataItems,
	}

	// 发送请求
	return c.SendSubscribeMessage(request)
}
