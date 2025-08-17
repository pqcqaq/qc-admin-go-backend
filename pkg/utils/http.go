package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	pkgurl "net/url"
	"strings"
	"time"
)

// HTTPClient 封装的HTTP客户端
type HTTPClient struct {
	client  *http.Client
	headers map[string]string
}

// NewHTTPClient 创建新的HTTP客户端
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		headers: make(map[string]string),
	}
}

// SetHeader 设置默认请求头
func (c *HTTPClient) SetHeader(key, value string) *HTTPClient {
	c.headers[key] = value
	return c
}

// SetHeaders 设置多个默认请求头
func (c *HTTPClient) SetHeaders(headers map[string]string) *HTTPClient {
	for k, v := range headers {
		c.headers[k] = v
	}
	return c
}

// SetUserAgent 设置User-Agent
func (c *HTTPClient) SetUserAgent(userAgent string) *HTTPClient {
	return c.SetHeader("User-Agent", userAgent)
}

// SetContentType 设置Content-Type
func (c *HTTPClient) SetContentType(contentType string) *HTTPClient {
	return c.SetHeader("Content-Type", contentType)
}

// SetAuth 设置Basic认证
func (c *HTTPClient) SetAuth(username, password string) *HTTPClient {
	return c.SetHeader("Authorization", "Basic "+BasicAuth(username, password))
}

// SetBearerToken 设置Bearer Token
func (c *HTTPClient) SetBearerToken(token string) *HTTPClient {
	return c.SetHeader("Authorization", "Bearer "+token)
}

// GET 发送GET请求
func (c *HTTPClient) GET(url string, params map[string]string) (*HTTPResponse, error) {
	if len(params) > 0 {
		url = url + "?" + BuildQuery(params)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

// POST 发送POST请求
func (c *HTTPClient) POST(url string, data interface{}) (*HTTPResponse, error) {
	var body *bytes.Buffer

	switch v := data.(type) {
	case string:
		body = bytes.NewBufferString(v)
	case []byte:
		body = bytes.NewBuffer(v)
	case map[string]interface{}:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
		c.SetContentType("application/json")
	case map[string]string:
		values := pkgurl.Values{}
		for k, v := range v {
			values.Set(k, v)
		}
		body = bytes.NewBufferString(values.Encode())
		c.SetContentType("application/x-www-form-urlencoded")
	default:
		jsonData, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
		c.SetContentType("application/json")
	}

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

// PUT 发送PUT请求
func (c *HTTPClient) PUT(url string, data interface{}) (*HTTPResponse, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.SetContentType("application/json")
	return c.doRequest(req)
}

// DELETE 发送DELETE请求
func (c *HTTPClient) DELETE(url string) (*HTTPResponse, error) {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

// doRequest 执行HTTP请求
func (c *HTTPClient) doRequest(req *http.Request) (*HTTPResponse, error) {
	// 设置默认请求头
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
		Text:       string(body),
	}, nil
}

// HTTPResponse HTTP响应结构
type HTTPResponse struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
	Text       string
}

// JSON 将响应体解析为JSON
func (r *HTTPResponse) JSON(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// IsSuccess 检查是否为成功状态码
func (r *HTTPResponse) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsError 检查是否为错误状态码
func (r *HTTPResponse) IsError() bool {
	return r.StatusCode >= 400
}

// 工具函数

// BuildQuery 构建查询参数字符串
func BuildQuery(params map[string]string) string {
	values := pkgurl.Values{}
	for k, v := range params {
		values.Set(k, v)
	}
	return values.Encode()
}

// BasicAuth 生成Basic认证字符串
func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// ParseURL 解析URL
func ParseURL(rawURL string) (*pkgurl.URL, error) {
	return pkgurl.Parse(rawURL)
}

// JoinURL 拼接URL路径
func JoinURL(baseURL, path string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	path = strings.TrimLeft(path, "/")
	return baseURL + "/" + path
}

// GetQueryParam 从URL中获取查询参数
func GetQueryParam(rawURL, key string) string {
	u, err := pkgurl.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Query().Get(key)
}

// 简化的全局HTTP函数

// SimpleGET 简单的GET请求
func SimpleGET(url string) (*HTTPResponse, error) {
	client := NewHTTPClient(30 * time.Second)
	return client.GET(url, nil)
}

// SimplePOST 简单的POST请求
func SimplePOST(url string, data interface{}) (*HTTPResponse, error) {
	client := NewHTTPClient(30 * time.Second)
	return client.POST(url, data)
}

// DownloadFile 下载文件
func DownloadFile(url string) ([]byte, error) {
	resp, err := SimpleGET(url)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return resp.Body, nil
}
