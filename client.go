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
	"context"
	"time"
)

type Client struct {
	AppID     string
	AppSecret string

	AppEnv string // release, trial, develop

	cacheKeyPrefix string
	cacheProvider  CacheProvider
}

type CacheProvider interface {
	Get(ctx context.Context, key string) string
	Set(ctx context.Context, key string, value string, expire time.Duration) error
}

func NewClient(appID, appSecret string, options ...func(*Client)) *Client {
	c := &Client{
		AppID:     appID,
		AppSecret: appSecret,
		AppEnv:    "release",
	}

	for _, option := range options {
		option(c)
	}

	return c
}

func WithAppEnv(env string) func(*Client) {
	return func(c *Client) {
		c.AppEnv = env
	}
}

func WithCacheKeyPrefix(prefix string) func(*Client) {
	return func(c *Client) {
		c.cacheKeyPrefix = prefix
	}
}

func WithCacheProvider(provider CacheProvider) func(*Client) {
	return func(c *Client) {
		c.cacheProvider = provider
	}
}
