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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestURLSchemeRequest(t *testing.T) {
	isExpire := false
	req := &URLSchemeRequest{
		JumpWxa: &JumpWxa{
			Path:       "/pages/index/index",
			Query:      "a=1&b=2",
			EnvVersion: "release",
		},
		IsExpire: &isExpire,
	}

	c := NewClient("appid", "secret")

	body, err := c.marshalURLSchemeRequest(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	t.Logf("request body: %s", string(body))

	expected := `{"jump_wxa":{"path":"/pages/index/index","query":"a=1&b=2","env_version":"release"},"is_expire":false}`
	assert.Equal(t, expected, string(body))
}

func TestURLSchemeRequestWithExpire(t *testing.T) {
	expireTime := time.Now().Add(24 * time.Hour).Unix()
	isExpire := true
	expireType := 0

	req := &URLSchemeRequest{
		JumpWxa: &JumpWxa{
			Path:  "/pages/test/test",
			Query: "scene=test",
		},
		IsExpire:   &isExpire,
		ExpireType: &expireType,
		ExpireTime: &expireTime,
	}

	c := NewClient("appid", "secret")

	body, err := c.marshalURLSchemeRequest(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	t.Logf("request body: %s", string(body))

	assert.Contains(t, string(body), `"is_expire":true`)
	assert.Contains(t, string(body), `"expire_type":0`)
	assert.Contains(t, string(body), `"expire_time":`)
}

func TestURLSchemeRequestWithInterval(t *testing.T) {
	isExpire := true
	expireType := 1
	expireInterval := 7
	req := &URLSchemeRequest{
		JumpWxa: &JumpWxa{
			Path:  "/pages/test/test",
			Query: "scene=test",
		},
		IsExpire:       &isExpire,
		ExpireType:     &expireType,
		ExpireInterval: &expireInterval,
	}

	c := NewClient("appid", "secret")

	body, err := c.marshalURLSchemeRequest(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	t.Logf("request body: %s", string(body))

	expected := `{"jump_wxa":{"path":"/pages/test/test","query":"scene=test"},"is_expire":true,"expire_type":1,"expire_interval":7}`
	assert.Equal(t, expected, string(body))
}

func TestGenerateExpirableURLSchemeWithTimeType(t *testing.T) {
	c := NewClient("test_appid", "test_secret")

	// Test that the function accepts time.Time parameter
	expireTime := time.Now().Add(24 * time.Hour)

	// This would normally make an HTTP request, but we're just testing the parameter type
	// In a real test environment, you'd mock the HTTP client
	_, err := c.GenerateExpirableURLScheme("/pages/test", "param=value", expireTime)

	// We expect an error because we don't have valid credentials, but the important thing
	// is that the function accepts time.Time parameter without compilation errors
	assert.Error(t, err) // This will fail due to invalid credentials, which is expected
}
