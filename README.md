# vwx - 微信 Go SDK

[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

vwx 是一个微信 Go SDK，提供了微信开发中常用的 API 接口封装, 包括小程序、消息推送等功能。

## 功能特性

- 🔐 内容安全检查
  - **文本内容安全检查** 
  - **多媒体内容安全检查** 
- 📱 用户数据处理
  - **手机号解密**
  - **会话管理**
-  📨 消息推送
  - **订阅消息**
- 🔗 工具功能
  - **小程序码生成**
  - **访问令牌管理**

## 安装

```bash
go get github.com/vogo/vwxa
```

## 快速开始

### 1. 初始化客户端

```go
package main

import (
    "github.com/vogo/vwxa"
)

func main() {
    // 基础初始化
    client := vwxa.NewClient("your-app-id", "your-app-secret")
    
    // 带配置选项的初始化
    client := vwxa.NewClient(
        "your-app-id", 
        "your-app-secret",
        vwxa.WithAppEnv("release"), // 环境：release, trial, develop
        vwxa.WithCacheKeyPrefix("myapp:"),
        vwxa.WithCacheProvider(yourCacheProvider), // 可选的缓存提供者
    )
}
```

### 2. 内容安全检查

#### 文本内容检查

```go
// 单条内容检查
result, err := client.MsgSecCheck("要检查的文本内容")
if err != nil {
    log.Printf("检查失败: %v", err)
    return
}

// 简单的安全性判断
isSafe, err := client.IsContentSafe("要检查的文本内容")
if err != nil {
    log.Printf("检查失败: %v", err)
    return
}

if isSafe {
    log.Println("内容安全")
} else {
    log.Println("内容存在风险")
}
```

#### 多媒体内容检查

```go
// 图片异步检查
result, err := client.CheckImageAsync("https://example.com/image.jpg", 1, "user-openid")
if err != nil {
    log.Printf("图片检查失败: %v", err)
    return
}
log.Printf("检查任务ID: %s", result.TraceID)

// 音频异步检查
result, err := client.CheckAudioAsync("https://example.com/audio.mp3", 1, "user-openid")
if err != nil {
    log.Printf("音频检查失败: %v", err)
    return
}

// 解析异步检查回调结果
callbackData := []byte(`{"trace_id":"xxx","status_code":0,...}`) // 微信回调数据
callbackResult, err := client.ParseMediaCheckCallback(callbackData)
if err != nil {
    log.Printf("解析回调失败: %v", err)
    return
}

// 检查是否违规
violationInfo, isViolation := client.CheckMediaViolation(callbackResult)
if isViolation {
    log.Printf("检测到违规内容: %s", violationInfo.Description)
    log.Printf("违规建议: %s", violationInfo.Suggestion)
}
```

### 3. 手机号解密

```go
// 方式1: 直接解析加密数据
encryptedData := []byte(`{
    "encrypted_data": "...",
    "iv": "...",
    "code": "..."
}`)

phoneInfo, err := client.ParsePhoneEncryptedData(encryptedData)
if err != nil {
    log.Printf("解析失败: %v", err)
    return
}

log.Printf("手机号: %s", phoneInfo.PhoneNumber)
log.Printf("纯手机号: %s", phoneInfo.PurePhoneNumber)
log.Printf("国家代码: %s", phoneInfo.CountryCode)

// 方式2: 直接解密
phoneInfo, err := client.DecryptPhoneNumber(sessionKey, encryptedData, iv)
if err != nil {
    log.Printf("解密失败: %v", err)
    return
}
```

### 4. 订阅消息

```go
// 发送订阅消息
request := &vwxa.SubscribeMessageRequest{
    ToUser:     "user-openid",
    TemplateID: "template-id",
    Page:       "pages/index/index",
    Data: map[string]*vwxa.SubscribeMessageDataItem{
        "thing1": {Value: "消息内容"},
        "time2":  {Value: "2024-01-01 12:00:00"},
    },
}

response, err := client.SendSubscribeMessage(request)
if err != nil {
    log.Printf("发送失败: %v", err)
    return
}

// 简化发送方式
err = client.SendSubscribeMessageSimple(
    "user-openid",
    "template-id",
    "pages/index/index",
    map[string]string{
        "thing1": "消息内容",
        "time2":  "2024-01-01 12:00:00",
    },
)
```

### 5. 小程序码生成

```go
// 生成小程序码
qrCodeData, err := client.GenerateQRCode("scene=123&param=value", "pages/index/index")
if err != nil {
    log.Printf("生成失败: %v", err)
    return
}

// 保存到文件
ioutil.WriteFile("qrcode.jpg", qrCodeData, 0644)
```

## 许可证

本项目采用 [Apache License 2.0](LICENSE) 许可证。

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。