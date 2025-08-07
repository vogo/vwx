# vwx - WeChat Go SDK

[![License](https://img.shields.io/badge/license-Apache%202.0-green.svg)](LICENSE)

vwx is a comprehensive WeChat Go SDK that provides API encapsulation for common WeChat development tasks, including Mini Programs and message push functionality.

## Features

- 🔐 **Content Security**
  - Text content security check
  - Multimedia content security check (images/audio)
  - Asynchronous media detection with callback support
- 📱 **User Data Processing**
  - Phone number decryption
  - Session management and authorization
  - User authentication via authorization codes
- 📨 **Message Push**
  - Subscribe message sending
  - Message push receiver with encryption/decryption
  - Support for both plain text and secure modes
- 🔗 **Utility Functions**
  - QR code generation for Mini Programs
  - Access token management with caching
  - Configurable environment support (release/trial/develop)

## Installation

```bash
go get github.com/vogo/vwx
```

## Quick Start

### 1. Initialize Client

```go
package main

import (
    "github.com/vogo/vwx/vwxa"
)

func main() {
    // Basic initialization
    client := vwxa.NewClient("your-app-id", "your-app-secret")
    
    // Initialization with configuration options
    client := vwxa.NewClient(
        "your-app-id", 
        "your-app-secret",
        vwxa.WithEnvVersion("release"), // Environment: release, trial, develop
        vwxa.WithCacheKeyPrefix("myapp:"),
        vwxa.WithCacheProvider(yourCacheProvider), // Optional cache provider
    )
}
```

### 2. Content Security Check

#### Text Content Check

```go
// Check if text content is safe
isSafe, err := client.IsMsgContentSafe("Hello, world!")
if err != nil {
    log.Fatal(err)
}

if isSafe {
    fmt.Println("Content is safe")
} else {
    fmt.Println("Content may have risks")
}

// Detailed security check
response, err := client.MsgSecCheck("Your text content")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Check result: %+v\n", response)
```

#### Multimedia Content Check

```go
// Asynchronous image check
response, err := client.CheckImageAsync(
    "https://example.com/image.jpg",
    vwxa.SceneProfile, // Scene: Profile, Comment, Forum, Social
    "user-openid",
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Trace ID: %s\n", response.TraceID)

// Asynchronous audio check
response, err = client.CheckAudioAsync(
    "https://example.com/audio.mp3",
    vwxa.SceneComment,
    "user-openid",
)

// Parse callback result
callbackResult, err := client.ParseMediaCheckCallback(callbackData)
if err != nil {
    log.Fatal(err)
}

// Check violation
violationInfo := client.CheckMediaViolation(callbackResult)
if violationInfo.IsViolation {
    fmt.Printf("Content violation: %s\n", violationInfo.Reason)
}
```

### 3. User Data Processing

#### Phone Number Decryption

```go
// Parse encrypted phone data
phoneInfo, sessionInfo, err := client.ParsePhoneEncryptedData(encryptedData)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Phone: %s\n", phoneInfo.PhoneNumber)
fmt.Printf("Pure Phone: %s\n", phoneInfo.PurePhoneNumber)
fmt.Printf("Country Code: %s\n", phoneInfo.CountryCode)
```

#### Session Management

```go
// Get session key using authorization code
sessionResponse, err := client.GetSessionKey("authorization-code")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("OpenID: %s\n", sessionResponse.OpenID)
fmt.Printf("Session Key: %s\n", sessionResponse.SessionKey)
```

### 4. Subscribe Messages

```go
// Simple subscribe message
response, err := client.SendSubscribeMessageSimple(
    "user-openid",
    "template-id",
    "pages/index",
    map[string]string{
        "thing1": "Hello",
        "time2":  "2024-01-01 12:00:00",
    },
)

// Advanced subscribe message
request := &vwxa.SubscribeMessageRequest{
    ToUser:     "user-openid",
    TemplateID: "template-id",
    Page:       "pages/detail",
    Data: map[string]*vwxa.SubscribeMessageDataItem{
        "thing1": {Value: "Hello World"},
        "time2":  {Value: "2024-01-01 12:00:00"},
    },
    MiniProgramState: "formal",
    Lang:             "zh_CN",
}

response, err = client.SendSubscribeMessage(request)
```

### 5. QR Code Generation

```go
// Generate QR code for Mini Program
qrCodeData, err := client.GenerateQRCode("scene-value", "pages/index")
if err != nil {
    log.Fatal(err)
}

// Save QR code to file
err = ioutil.WriteFile("qrcode.jpg", qrCodeData, 0644)
```

### 6. Message Push Receiver

```go
import "github.com/vogo/vwx/vwxpush"

// Initialize push receiver
receiver := &vwxpush.WxPushReceiver{
    Token:          "your-token",
    EncodingAESKey: "your-aes-key",
    SecurityMode:   "secure", // or "plain"
    DataType:       "xml",    // or "json"
}

// Handle push message
response, err := receiver.HandlePushMessage(
    func(name string) string {
        // Return URL parameter value by name
        return getURLParam(name)
    },
    requestBody,
    func(appID string, decryptedContent []byte) ([]byte, error) {
        // Your business logic here
        fmt.Printf("Received message from %s: %s\n", appID, string(decryptedContent))
        return []byte("success"), nil
    },
)
```

## API Reference

### vwxa Package

#### Client
- `NewClient(appID, appSecret string, options ...func(*Client)) *Client`
- `WithEnvVersion(env string) func(*Client)`
- `WithCacheKeyPrefix(prefix string) func(*Client)`
- `WithCacheProvider(provider CacheProvider) func(*Client)`

#### Access Token
- `GetAccessToken() (string, error)`

#### Session Management
- `GetSessionKey(code string) (*SessionResponse, error)`

#### Phone Number
- `ParsePhoneEncryptedData(data []byte) (*PhoneInfo, *SessionResponse, error)`
- `DecryptPhoneNumber(sessionKey, encryptedData, iv string) (*PhoneInfo, error)`

#### Content Security
- `MsgSecCheck(content string) (*MsgSecCheckResponse, error)`
- `IsMsgContentSafe(content string) (bool, error)`
- `MediaCheckAsync(mediaURL string, mediaType, scene int, openID string) (*MediaCheckAsyncResponse, error)`
- `CheckImageAsync(imageURL string, scene int, openID string) (*MediaCheckAsyncResponse, error)`
- `CheckAudioAsync(audioURL string, scene int, openID string) (*MediaCheckAsyncResponse, error)`
- `ParseMediaCheckCallback(callbackData []byte) (*MediaCheckCallbackResult, error)`
- `CheckMediaViolation(result *MediaCheckCallbackResult) *ViolationInfo`

#### Subscribe Messages
- `SendSubscribeMessage(request *SubscribeMessageRequest) (*SubscribeMessageResponse, error)`
- `SendSubscribeMessageSimple(openID, templateID, page string, data map[string]string) (*SubscribeMessageResponse, error)`

#### QR Code
- `GenerateQRCode(scene, page string) ([]byte, error)`

### vwxpush Package

#### Message Push Receiver
- `HandlePushMessage(parameterFetcher func(string) string, body []byte, handler func(string, []byte) ([]byte, error)) ([]byte, error)`

## Constants

### Media Types
- `MediaTypeAudio = 1` // Audio
- `MediaTypeImage = 2` // Image

### Scenes
- `SceneProfile = 1` // Profile
- `SceneComment = 2` // Comment
- `SceneForum = 3`   // Forum
- `SceneSocial = 4`  // Social Log

## Cache Provider Interface

```go
type CacheProvider interface {
    Get(ctx context.Context, key string) string
    Set(ctx context.Context, key string, value string, expire time.Duration) error
}
```

Implement this interface to provide custom caching for access tokens.

## Error Handling

All API methods return appropriate error types. WeChat API errors are wrapped with descriptive messages.

```go
response, err := client.MsgSecCheck("content")
if err != nil {
    // Handle error
    log.Printf("Security check failed: %v", err)
    return
}

if response.ErrCode != 0 {
    log.Printf("WeChat API error: %d %s", response.ErrCode, response.ErrMsg)
}
## License

Apache License 2.0 - see [LICENSE](LICENSE) file for details.

## Contributing

Pull requests and issues are welcome!

## Support

If you find this project useful, please give it a ⭐️!