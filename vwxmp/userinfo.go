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

package vwxmp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vogo/vogo/vlog"
)

const (
	userInfoURL = "https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s&lang=%s"
)

// UserInfoLang represents the language for user info response.
type UserInfoLang string

const (
	LangZhCN UserInfoLang = "zh_CN" // 简体中文
	LangZhTW UserInfoLang = "zh_TW" // 繁体中文
	LangEN   UserInfoLang = "en"    // 英文
)

// UserInfoResponse represents the response from user info API.
type UserInfoResponse struct {
	OpenID     string   `json:"openid"`     // 用户的唯一标识
	Nickname   string   `json:"nickname"`   // 用户昵称
	Sex        int      `json:"sex"`        // 用户的性别，值为1时是男性，值为2时是女性，值为0时是未知
	Province   string   `json:"province"`   // 用户个人资料填写的省份
	City       string   `json:"city"`       // 普通用户个人资料填写的城市
	Country    string   `json:"country"`    // 国家，如中国为CN
	HeadImgURL string   `json:"headimgurl"` // 用户头像
	Privilege  []string `json:"privilege"`  // 用户特权信息
	UnionID    string   `json:"unionid"`    // 只有在用户将公众号绑定到微信开放平台账号后，才会出现该字段
	ErrCode    int      `json:"errcode"`
	ErrMsg     string   `json:"errmsg"`
}

// GetUserInfo retrieves user profile information.
// accessToken: OAuth access token (obtained from GetOAuthAccessToken)
// openID: user's openid
// lang: language for response (zh_CN, zh_TW, en)
func (s *Service) GetUserInfo(accessToken, openID string, lang UserInfoLang) (*UserInfoResponse, error) {
	vlog.Infof("get user info | openid: %s | lang: %s", openID, lang)

	if lang == "" {
		lang = LangZhCN
	}

	requestURL := fmt.Sprintf(userInfoURL, accessToken, openID, lang)

	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vlog.Errorf("failed to close response body | err: %v", closeErr)
		}
	}()

	var result UserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error: %d %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}
