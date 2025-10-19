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

	"github.com/stretchr/testify/assert"
)

func TestLinkRequest(t *testing.T) {
	req := &URLLinkRequest{
		Path:  "/test",
		Query: "a=1&b=2",
	}

	c := NewClient("appid", "secret")

	body, err := c.marshalRequest(req)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	t.Logf("request body: %s", string(body))

	assert.Equal(t, string(body), `{"path":"/test","query":"a=1&b=2"}`)
}
