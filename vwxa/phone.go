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
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/vogo/vogo/vlog"
)

// PhoneEncryptedData represents the encrypted phone data from WeChat Mini Program.
type PhoneEncryptedData struct {
	EncryptedData string `json:"encryptedData"`
	IV            string `json:"iv"`
	Code          string `json:"code"`
}

// PhoneInfo represents the decrypted phone information from WeChat.
type PhoneInfo struct {
	PhoneNumber     string `json:"phoneNumber"`
	PurePhoneNumber string `json:"purePhoneNumber"`
	CountryCode     string `json:"countryCode"`
}

// ParsePhoneEncryptedData parses and decrypts phone encrypted data from WeChat Mini Program.
func (c *Client) ParsePhoneEncryptedData(data []byte) (*PhoneInfo, *SessionResponse, error) {
	var encData PhoneEncryptedData
	err := json.Unmarshal(data, &encData)
	if err != nil {
		return nil, nil, err
	}

	if len(encData.Code) == 0 || len(encData.EncryptedData) == 0 || len(encData.IV) == 0 {
		return nil, nil, fmt.Errorf("code or encryptedData or iv is empty")
	}

	sessionInfo, err := c.GetSessionKey(encData.Code)
	if err != nil {
		return nil, nil, err
	}

	phoneInfo, err := c.DecryptPhoneNumber(sessionInfo.SessionKey, encData.EncryptedData, encData.IV)
	if err != nil {
		return nil, nil, err
	}

	return phoneInfo, sessionInfo, nil
}

// DecryptPhoneNumber decrypts phone number using session key, encrypted data and IV.
func (c *Client) DecryptPhoneNumber(sessionKey, encryptedData, iv string) (_info *PhoneInfo, _err error) {
	defer func() {
		if err := recover(); err != nil {
			vlog.Errorf("decrypt phone number error: %v, stack: %s", err, debug.Stack())
			_err = fmt.Errorf("decrypt phone number error: %v", err)
		}
	}()

	vlog.Infof("decrypt phone number: sessionKey=%s, encryptedData=%s, iv=%s",
		sessionKey, encryptedData, iv)

	key, err := base64.StdEncoding.DecodeString(sessionKey)
	if err != nil {
		return nil, err
	}

	cipherText, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	mode := cipher.NewCBCDecrypter(block, ivBytes)
	mode.CryptBlocks(cipherText, cipherText)

	// 处理 PKCS#7 填充
	cipherText = pkcs7Unpad(cipherText)
	if cipherText == nil {
		vlog.Error("decrypt phone number error: unpad failed")
		return nil, fmt.Errorf("unpad failed")
	}

	var phoneInfo PhoneInfo
	if err = json.Unmarshal(cipherText, &phoneInfo); err != nil {
		return nil, err
	}

	return &phoneInfo, nil
}

func pkcs7Unpad(data []byte) []byte {
	length := len(data)
	if length == 0 {
		return nil
	}

	padding := int(data[length-1])
	if padding > length {
		return nil
	}

	for i := length - padding; i < length; i++ {
		if data[i] != byte(padding) {
			return nil
		}
	}

	return data[:length-padding]
}
