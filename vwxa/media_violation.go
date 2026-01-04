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
	mediaCheckAsyncURL = "https://api.weixin.qq.com/wxa/media_check_async?access_token=%s"
)

const (
	ViolationMediaTypeAudio = 1 // 音频
	ViolationMediaTypeImage = 2 // 图片

	ViolationSceneProfile = 1 // 资料
	ViolationSceneComment = 2 // 评论
	ViolationSceneForum   = 3 // 论坛
	ViolationSceneSocial  = 4 // 社交日志

	ViolationSuggestRisky  = "risky"  // 风险
	ViolationSuggestPass   = "pass"   // 通过
	ViolationSuggestReview = "review" // 审核
)

// MediaViolationCheckAsyncRequest represents a request for asynchronous media content security check.
type MediaViolationCheckAsyncRequest struct {
	MediaURL  string `json:"media_url"`  // 要检测的图片或音频的url
	MediaType int    `json:"media_type"` // 1:音频;2:图片
	Version   int    `json:"version"`    // 接口版本号，2.0版本为固定值2
	Scene     int    `json:"scene"`      // 场景枚举值（1 资料；2 评论；3 论坛；4 社交日志）
	OpenID    string `json:"openid"`     // 用户的openid（用户需在近两小时访问过小程序）
}

// MediaViolationCheckAsyncResponse represents the response from asynchronous media content security check.
type MediaViolationCheckAsyncResponse struct {
	ErrCode int    `json:"errcode"`  // 错误码
	ErrMsg  string `json:"errmsg"`   // 错误信息
	TraceID string `json:"trace_id"` // 唯一请求标识，标记单次请求，用于匹配异步推送结果
}

// MediaViolationCheckCallbackResult represents the callback result data structure for asynchronous detection.
type MediaViolationCheckCallbackResult struct {
	ToUserName   string                             `json:"ToUserName"`   // 小程序的username
	FromUserName string                             `json:"FromUserName"` // 平台推送服务UserName
	CreateTime   int64                              `json:"CreateTime"`   // 发送时间
	MsgType      string                             `json:"MsgType"`      // 默认为：event
	Event        string                             `json:"Event"`        // 默认为：wxa_media_check
	AppID        string                             `json:"appid"`        // 小程序的appid
	TraceID      string                             `json:"trace_id"`     // 任务id
	Version      int                                `json:"version"`      // 可用于区分接口版本
	ErrCode      int                                `json:"errcode"`      // 错误码，仅当该值为0时，结果有效
	Result       *MediaViolationCheckResult         `json:"result"`       // 综合结果
	Detail       []*MediaViolationCheckDetailResult `json:"detail"`       // 详细检测结果
}

// MediaViolationCheckResult represents the comprehensive detection result.
type MediaViolationCheckResult struct {
	Suggest string `json:"suggest"` // 建议，有risky、pass、review三种值
	Label   int    `json:"label"`   // 命中标签枚举值，100 正常；20001 时政；20002 色情；20006 违法犯罪；21000 其他
}

// MediaViolationCheckDetailResult represents the detailed detection result.
type MediaViolationCheckDetailResult struct {
	Strategy string `json:"strategy"` // 策略类型
	ErrCode  int    `json:"errcode"`  // 错误码，仅当该值为0时，该项结果有效
	Suggest  string `json:"suggest"`  // 建议，有risky、pass、review三种值
	Label    int    `json:"label"`    // 命中标签枚举值，100 正常；20001 时政；20002 色情；20006 违法犯罪；21000 其他
	Prob     int    `json:"prob"`     // 0-100，代表置信度，越高代表越有可能属于当前返回的标签（label）
}

// MediaViolationInfo represents information about content violation.
type MediaViolationInfo struct {
	IsViolation bool   `json:"is_violation"` // 是否违规
	Reason      string `json:"reason"`       // 违规原因
	Label       int    `json:"label"`        // 违规标签
	Suggest     string `json:"suggest"`      // 建议操作
}

// MediaCheckAsync asynchronously detects whether images/audio contain illegal or non-compliant content.
// mediaURL: URL of the image or audio to be detected
// mediaType: 1 for audio, 2 for image
// scene: Scene enumeration value (1 profile, 2 comment, 3 forum, 4 social log)
// openID: User's openid (user must have accessed the mini program within the last two hours)
// Rate limit: single appId call limit is 2000 times/minute, 200,000 times/day; file size limit: single file size not exceeding 10M
func (c *Service) MediaViolationCheckAsync(mediaURL string, mediaType, scene int, openID string) (*MediaViolationCheckAsyncResponse, error) {
	accessToken, err := c.authSvc.GetAccessToken()
	if err != nil {
		return nil, fmt.Errorf("get access token error: %v", err)
	}

	url := fmt.Sprintf(mediaCheckAsyncURL, accessToken)

	request := &MediaViolationCheckAsyncRequest{
		MediaURL:  mediaURL,
		MediaType: mediaType,
		Version:   2, // 2.0版本固定值
		Scene:     scene,
		OpenID:    openID,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal request error: %v", err)
	}

	vlog.Infof("media check async | req: %s", string(data))

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

	vlog.Infof("media check async | resp: %s", string(body))

	var response MediaViolationCheckAsyncResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("unmarshal response error: %v", err)
	}

	if response.ErrCode != 0 {
		return &response, errors.New(response.ErrMsg)
	}

	return &response, nil
}

// ParseMediaCheckCallback parses the asynchronous callback result of multimedia content security detection.
func (c *Service) ParseMediaCheckCallback(callbackData []byte) (*MediaViolationCheckCallbackResult, error) {
	var result MediaViolationCheckCallbackResult
	if err := json.Unmarshal(callbackData, &result); err != nil {
		return nil, fmt.Errorf("unmarshal callback data error: %v", err)
	}

	return &result, nil
}

// CheckMediaViolation determines whether multimedia content violates regulations and returns violation description.
func (c *Service) CheckMediaViolation(result *MediaViolationCheckCallbackResult) *MediaViolationInfo {
	violationInfo := &MediaViolationInfo{
		IsViolation: false,
		Reason:      "内容正常",
		Label:       100,
		Suggest:     "pass",
	}

	// 检查错误码
	if result.ErrCode != 0 {
		violationInfo.IsViolation = true
		violationInfo.Reason = fmt.Sprintf("检测失败，错误码：%d", result.ErrCode)
		return violationInfo
	}

	// 检查综合结果
	if result.Result != nil {
		violationInfo.Label = result.Result.Label
		violationInfo.Suggest = result.Result.Suggest

		switch result.Result.Suggest {
		case ViolationSuggestRisky:
			violationInfo.IsViolation = true
			violationInfo.Reason = c.getLabelDescription(result.Result.Label)
		case ViolationSuggestReview:
			violationInfo.IsViolation = true
			violationInfo.Reason = fmt.Sprintf("内容需要人工审核：%s", c.getLabelDescription(result.Result.Label))
		case ViolationSuggestPass:
			violationInfo.IsViolation = false
			violationInfo.Reason = "内容正常"
		}
	}

	// 检查详细结果，如果有任何一项为risky，则认为违规
	for _, detail := range result.Detail {
		if detail.ErrCode == 0 && detail.Suggest == "risky" {
			violationInfo.IsViolation = true
			if violationInfo.Reason == "内容正常" {
				violationInfo.Reason = fmt.Sprintf("检测到违规内容：%s（策略：%s，置信度：%d%%）",
					c.getLabelDescription(detail.Label), detail.Strategy, detail.Prob)
			}
			break
		}
	}

	return violationInfo
}

// 获取标签描述
// getLabelDescription returns the description for a given label code.
func (c *Service) getLabelDescription(label int) string {
	switch label {
	case 100:
		return "正常内容"
	case 20001:
		return "时政内容"
	case 20002:
		return "色情内容"
	case 20006:
		return "违法犯罪内容"
	case 21000:
		return "其他违规内容"
	default:
		return fmt.Sprintf("未知标签：%d", label)
	}
}

// CheckImageAsync is a convenient method for asynchronous image content security detection.
func (c *Service) CheckImageAsync(imageURL string, scene int, openID string) (*MediaViolationCheckAsyncResponse, error) {
	return c.MediaViolationCheckAsync(imageURL, 2, scene, openID)
}

// CheckAudioAsync is a convenient method for asynchronous audio content security detection.
func (c *Service) CheckAudioAsync(audioURL string, scene int, openID string) (*MediaViolationCheckAsyncResponse, error) {
	return c.MediaViolationCheckAsync(audioURL, 1, scene, openID)
}
