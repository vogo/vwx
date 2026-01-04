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

// Package vwxa provides WeChat Mini Program API client functionality.
package vwx

import (
	"context"
	"time"
)

// Client represents a WeChat Mini Program API client.
type Client struct {
	AppID     string
	AppSecret string

	EnvVersion string // release, trial, develop

	CacheKeyPrefix string
	CacheProvider  CacheProvider
}

// CacheProvider defines the interface for caching access tokens and other data.
type CacheProvider interface {
	Get(ctx context.Context, key string) string
	Set(ctx context.Context, key string, value string, expire time.Duration) error
}

// NewClient creates a new WeChat Mini Program API client with the given app ID and secret.
func NewClient(appID, appSecret string, options ...func(*Client)) *Client {
	c := &Client{
		AppID:      appID,
		AppSecret:  appSecret,
		EnvVersion: "release",
	}

	for _, option := range options {
		option(c)
	}

	return c
}

// WithEnvVersion sets the app environment (release, trial, develop).
func WithEnvVersion(env string) func(*Client) {
	return func(c *Client) {
		c.EnvVersion = env
	}
}

// WithCacheKeyPrefix sets the cache key prefix for the client.
func WithCacheKeyPrefix(prefix string) func(*Client) {
	return func(c *Client) {
		c.CacheKeyPrefix = prefix
	}
}

// WithCacheProvider sets the cache provider for the client.
func WithCacheProvider(provider CacheProvider) func(*Client) {
	return func(c *Client) {
		c.CacheProvider = provider
	}
}
