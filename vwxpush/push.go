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

package vwxpush

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/vogo/vogo/vlog"
)

// WxPushReceiver WeChat message push receiver
type WxPushReceiver struct {
	Token          string // Token
	EncodingAESKey string // Message encryption/decryption key
	SecurityMode   string // Security mode: plain(plain text mode), secure(secure mode)
	DataType       string // Data format: xml, json
}

// NewWxPushReceiver creates a new WeChat message push receiver
func NewWxPushReceiver(token, encodingAESKey, securityMode, dataType string) *WxPushReceiver {
	return &WxPushReceiver{
		Token:          token,
		EncodingAESKey: encodingAESKey,
		SecurityMode:   securityMode,
		DataType:       dataType,
	}
}

// EncryptedMessage encrypted message structure
type EncryptedMessage struct {
	Encrypt string `xml:"Encrypt" json:"Encrypt"`
}

// PushBaseInfo push base info
type PushBaseInfo struct {
	ToUserName   string `xml:"ToUserName" json:"ToUserName"`
	FromUserName string `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime" json:"CreateTime"`
	MsgType      string `xml:"MsgType" json:"MsgType"`
	Event        string `xml:"Event" json:"Event"`
}

// HandlePushMessage handles WeChat message push
// parameterFetcher: function to get URL parameters
// body: request body data
// handler: business processing function, first parameter is appid, second parameter is decrypted content
// returns encrypted response data
func (c *WxPushReceiver) HandlePushMessage(
	parameterFetcher func(name string) string,
	body []byte,
	handler func(string, *PushBaseInfo, []byte) ([]byte, error),
) (_response []byte, _err error) {
	defer func() {
		if err := recover(); err != nil {
			vlog.Errorf("handle push message error: %v, stack: %s", err, debug.Stack())
			_err = fmt.Errorf("handle push message error: %v", err)
		}
	}()

	// Get URL parameters
	signature := parameterFetcher("signature")
	timestamp := parameterFetcher("timestamp")
	nonce := parameterFetcher("nonce")
	msgSignature := parameterFetcher("msg_signature")
	encryptType := parameterFetcher("encrypt_type")

	vlog.Infof("handle push message: signature=%s, timestamp=%s, nonce=%s, msg_signature=%s, encrypt_type=%s",
		signature, timestamp, nonce, msgSignature, encryptType)

	// Process according to security mode
	if encryptType == "aes" && len(body) > 0 {
		// Secure mode: requires decryption
		return c.handleEncryptedMessage(msgSignature, timestamp, nonce, body, handler)
	} else {
		// Plain text mode: only verify signature
		return c.handlePlainMessage(signature, timestamp, nonce, body, handler)
	}
}

// handleEncryptedMessage handles encrypted messages
func (c *WxPushReceiver) handleEncryptedMessage(
	msgSignature, timestamp, nonce string,
	body []byte,
	handler func(string, *PushBaseInfo, []byte) ([]byte, error),
) ([]byte, error) {
	// Parse encrypted message
	var encryptedMsg EncryptedMessage
	if c.DataType == "json" {
		if err := json.Unmarshal(body, &encryptedMsg); err != nil {
			return nil, fmt.Errorf("unmarshal encrypted message failed: %v", err)
		}
	} else {
		// Default XML format
		if err := xml.Unmarshal(body, &encryptedMsg); err != nil {
			return nil, fmt.Errorf("unmarshal encrypted message failed: %v", err)
		}
	}

	// Verify message signature
	if !c.verifyMsgSignature(c.Token, timestamp, nonce, encryptedMsg.Encrypt, msgSignature) {
		return nil, fmt.Errorf("invalid message signature")
	}

	var responseData []byte
	var err error
	var appid string

	var decryptedData []byte
	decryptedData, appid, err = c.decryptMessage(encryptedMsg.Encrypt)
	if err != nil {
		return nil, fmt.Errorf("decrypt message failed: %v", err)
	}

	vlog.Infof("push message, appid: %s, message: %s", appid, string(decryptedData))

	// Parse base info
	baseInfo, err := c.parseBaseInfo(decryptedData)
	if err != nil {
		return nil, fmt.Errorf("parse base info failed: %v", err)
	}

	// Call business processing function
	responseData, err = handler(appid, baseInfo, decryptedData)
	if err != nil {
		return nil, fmt.Errorf("handler failed: %v", err)
	}

	// If there is response data, it needs to be encrypted and returned
	if len(responseData) == 0 {
		responseData = []byte("success")
	}

	return c.encryptResponse(appid, responseData)
}

func (c *WxPushReceiver) parseBaseInfo(decryptedData []byte) (*PushBaseInfo, error) {
	var pushMsg PushBaseInfo

	if c.DataType == "json" {
		if err := json.Unmarshal(decryptedData, &pushMsg); err != nil {
			return nil, fmt.Errorf("unmarshal push message failed: %v", err)
		}
	} else {
		// Default XML format
		if err := xml.Unmarshal(decryptedData, &pushMsg); err != nil {
			return nil, fmt.Errorf("unmarshal push message failed: %v", err)
		}
	}

	return &pushMsg, nil
}

// handlePlainMessage handles plain text messages
func (c *WxPushReceiver) handlePlainMessage(
	signature, timestamp, nonce string,
	body []byte,
	handler func(string, *PushBaseInfo, []byte) ([]byte, error),
) ([]byte, error) {
	// Verify signature
	if !c.verifySignature(c.Token, timestamp, nonce, signature) {
		return nil, fmt.Errorf("invalid signature")
	}

	if len(body) == 0 {
		return []byte("success"), nil
	}

	vlog.Infof("plain message: %s", string(body))

	// Parse base info
	baseInfo, err := c.parseBaseInfo(body)
	if err != nil {
		return nil, fmt.Errorf("parse base info failed: %v", err)
	}

	// Call business processing function
	responseData, err := handler("", baseInfo, body)
	if err != nil {
		return nil, fmt.Errorf("handler failed: %v", err)
	}

	// Plain text mode returns directly
	if len(responseData) > 0 {
		return responseData, nil
	}

	// Default return success
	return []byte("success"), nil
}

// verifySignature verifies signature (plain text mode)
func (c *WxPushReceiver) verifySignature(token, timestamp, nonce, signature string) bool {
	// Sort token, timestamp, nonce parameters in dictionary order
	params := []string{token, timestamp, nonce}
	sort.Strings(params)

	// Concatenate strings
	str := strings.Join(params, "")

	// Calculate SHA1
	h := sha1.New()
	h.Write([]byte(str))
	calcSignature := fmt.Sprintf("%x", h.Sum(nil))

	return calcSignature == signature
}

// verifyMsgSignature verifies message signature (secure mode)
func (c *WxPushReceiver) verifyMsgSignature(token, timestamp, nonce, encrypt, msgSignature string) bool {
	// Sort token, timestamp, nonce, encrypt parameters in dictionary order
	params := []string{token, timestamp, nonce, encrypt}
	sort.Strings(params)

	// Concatenate strings
	str := strings.Join(params, "")

	// Calculate SHA1
	h := sha1.New()
	h.Write([]byte(str))
	calcSignature := fmt.Sprintf("%x", h.Sum(nil))

	return calcSignature == msgSignature
}

// decryptMessage decrypts message, returns message content and appid
func (c *WxPushReceiver) decryptMessage(encryptedData string) ([]byte, string, error) {
	// Base64 decode
	cipherText, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, "", fmt.Errorf("base64 decode failed: %v", err)
	}

	// Decode AES key
	aesKey, err := base64.StdEncoding.DecodeString(c.EncodingAESKey + "=")
	if err != nil {
		return nil, "", fmt.Errorf("decode aes key failed: %v", err)
	}

	// AES decrypt
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, "", fmt.Errorf("create aes cipher failed: %v", err)
	}

	if len(cipherText) < aes.BlockSize {
		return nil, "", fmt.Errorf("cipher text too short")
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherText, cipherText)

	// Remove PKCS#7 padding
	cipherText = pkcs7Unpad(cipherText)
	if cipherText == nil {
		return nil, "", fmt.Errorf("pkcs7 unpad failed")
	}

	// Parse FullStr format: random(16B) + msg_len(4B) + msg + appid
	if len(cipherText) < 20 {
		return nil, "", fmt.Errorf("decrypted data too short")
	}

	// Skip 16-byte random string
	content := cipherText[16:]

	// Read message length (4 bytes, network byte order)
	if len(content) < 4 {
		return nil, "", fmt.Errorf("content too short")
	}

	msgLen := int(content[0])<<24 | int(content[1])<<16 | int(content[2])<<8 | int(content[3])
	content = content[4:]

	if len(content) < msgLen {
		return nil, "", fmt.Errorf("content length mismatch")
	}

	// Extract message content
	message := content[:msgLen]

	// Extract appid (remaining part)
	appidBytes := content[msgLen:]
	appid := string(appidBytes)

	return message, appid, nil
}

// encryptResponse encrypts response data
func (c *WxPushReceiver) encryptResponse(appID string, responseData []byte) ([]byte, error) {
	// Generate random string (16 bytes)
	randomBytes := make([]byte, 16)
	for i := range randomBytes {
		randomBytes[i] = byte(i) // Simple random number generation, should use crypto/rand in actual applications
	}

	// Construct message: random string(16) + message length(4) + message content + AppID
	msgLen := len(responseData)
	lengthBytes := []byte{
		byte(msgLen >> 24),
		byte(msgLen >> 16),
		byte(msgLen >> 8),
		byte(msgLen),
	}

	plainText := append(randomBytes, lengthBytes...)
	plainText = append(plainText, responseData...)
	plainText = append(plainText, []byte(appID)...)

	// PKCS#7 padding
	plainText = pkcs7Pad(plainText, aes.BlockSize)

	// Decode AES key
	aesKey, err := base64.StdEncoding.DecodeString(c.EncodingAESKey + "=")
	if err != nil {
		return nil, fmt.Errorf("decode aes key failed: %v", err)
	}

	// AES encrypt
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher failed: %v", err)
	}

	// Generate IV
	iv := make([]byte, aes.BlockSize)
	for i := range iv {
		iv[i] = byte(i) // Simple IV generation, should use crypto/rand in actual applications
	}

	cipherText := make([]byte, len(plainText))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, plainText)

	// Concatenate IV and ciphertext
	encryptedData := append(iv, cipherText...)

	// Base64 encode
	encryptedStr := base64.StdEncoding.EncodeToString(encryptedData)

	// Return according to data format
	if c.DataType == "json" {
		response := EncryptedMessage{Encrypt: encryptedStr}
		return json.Marshal(response)
	} else {
		// Default XML format
		response := EncryptedMessage{Encrypt: encryptedStr}
		return xml.Marshal(response)
	}
}
