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
	"net/url"

	"github.com/vogo/vogo/vlog"
)

const (
	authorizeURL         = "https://open.weixin.qq.com/connect/oauth2/authorize"
	oauthAccessTokenURL  = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	oauthRefreshTokenURL = "https://api.weixin.qq.com/sns/oauth2/refresh_token?appid=%s&grant_type=refresh_token&refresh_token=%s"
	oauthCheckTokenURL   = "https://api.weixin.qq.com/sns/auth?access_token=%s&openid=%s"
)

// OAuthScope represents the authorization scope.
type OAuthScope string

const (
	// ScopeBase provides silent authorization, only returns openid.
	ScopeBase OAuthScope = "snsapi_base"
	// ScopeUserInfo requires user confirmation, returns user profile.
	ScopeUserInfo OAuthScope = "snsapi_userinfo"
)

// OAuthAccessTokenResponse represents the response from OAuth access token API.
type OAuthAccessTokenResponse struct {
	AccessToken    string `json:"access_token"`    // 网页授权接口调用凭证
	ExpiresIn      int    `json:"expires_in"`      // access_token接口调用凭证超时时间，单位（秒）
	RefreshToken   string `json:"refresh_token"`   // 用户刷新access_token
	OpenID         string `json:"openid"`          // 用户唯一标识
	Scope          string `json:"scope"`           // 用户授权的作用域
	IsSnapshotUser int    `json:"is_snapshotuser"` // 是否为快照页模式虚拟账号，值为1时是
	UnionID        string `json:"unionid"`         // 用户统一标识（snsapi_userinfo作用域时返回）
	ErrCode        int    `json:"errcode"`
	ErrMsg         string `json:"errmsg"`
}

// BuildAuthorizeURL builds the authorization URL for user to authorize.
// redirectURI: callback URL after authorization
// scope: authorization scope (snsapi_base or snsapi_userinfo)
// state: custom state parameter, will be returned in callback
// forcePopup: force popup for user confirmation even in snsapi_base scope
func (s *Service) BuildAuthorizeURL(redirectURI string, scope OAuthScope, state string, forcePopup bool) string {
	params := url.Values{}
	params.Set("appid", s.client.AppID)
	params.Set("redirect_uri", redirectURI)
	params.Set("response_type", "code")
	params.Set("scope", string(scope))

	if state != "" {
		params.Set("state", state)
	}

	if forcePopup {
		params.Set("forcePopup", "true")
	}

	return fmt.Sprintf("%s?%s#wechat_redirect", authorizeURL, params.Encode())
}

// GetOAuthAccessToken exchanges authorization code for access token.
// code: authorization code obtained from redirect callback
func (s *Service) GetOAuthAccessToken(code string) (*OAuthAccessTokenResponse, error) {
	vlog.Infof("get oauth access token | appid: %s | code: %s", s.client.AppID, code)

	requestURL := fmt.Sprintf(oauthAccessTokenURL, s.client.AppID, s.client.AppSecret, code)

	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vlog.Errorf("failed to close response body | err: %v", closeErr)
		}
	}()

	var result OAuthAccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error: %d %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// RefreshOAuthAccessToken refreshes the access token using refresh token.
// refreshToken: refresh token obtained from GetOAuthAccessToken
func (s *Service) RefreshOAuthAccessToken(refreshToken string) (*OAuthAccessTokenResponse, error) {
	vlog.Infof("refresh oauth access token | appid: %s", s.client.AppID)

	requestURL := fmt.Sprintf(oauthRefreshTokenURL, s.client.AppID, refreshToken)

	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vlog.Errorf("failed to close response body | err: %v", closeErr)
		}
	}()

	var result OAuthAccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("wechat error: %d %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// CheckOAuthAccessToken validates the access token.
// accessToken: OAuth access token to validate
// openID: user's openid
func (s *Service) CheckOAuthAccessToken(accessToken, openID string) error {
	vlog.Infof("check oauth access token | openid: %s", openID)

	requestURL := fmt.Sprintf(oauthCheckTokenURL, accessToken, openID)

	resp, err := http.Get(requestURL)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			vlog.Errorf("failed to close response body | err: %v", closeErr)
		}
	}()

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("wechat error: %d %s", result.ErrCode, result.ErrMsg)
	}

	return nil
}
