package nacos

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	ServerURL   string
	Username    string
	Password    string
	Namespace   string
	Token       string
	TokenExpiry int64 // Unix时间戳，表示token过期时间
}

type Config struct {
	DataID  string `json:"dataId"`
	Group   string `json:"group"`
	Content string `json:"content"`
	Type    string `json:"type"`
}

type Namespace struct {
	Namespace         string `json:"namespace"`
	NamespaceShowName string `json:"namespaceShowName"`
	NamespaceDesc     string `json:"namespaceDesc"`
	Quota             int    `json:"quota"`
	ConfigCount       int    `json:"configCount"`
	Type              int    `json:"type"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	TokenTtl    int    `json:"tokenTtl"`
	GlobalAdmin bool   `json:"globalAdmin"`
}

func NewClient(serverURL, username, password, namespace string) *Client {
	return &Client{
		ServerURL:   strings.TrimSuffix(serverURL, "/"),
		Username:    username,
		Password:    password,
		Namespace:   namespace,
		Token:       "",
		TokenExpiry: 0,
	}
}

func (c *Client) Login() (*LoginResponse, error) {
	data := url.Values{}
	data.Set("username", c.Username)
	data.Set("password", c.Password)

	// 使用Nacos v1 API
	loginURL := c.ServerURL + "/nacos/v1/auth/users/login"

	// 创建请求
	req, err := http.NewRequest("POST", loginURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("创建登录请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "nacos-cli")

	// 发送请求
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("登录失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("登录失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return nil, fmt.Errorf("解析登录响应失败: %w, 原始响应: %s", err, string(body))
	}

	if loginResp.AccessToken == "" {
		return nil, fmt.Errorf("登录失败，未获取到有效的访问令牌")
	}

	// 设置token和过期时间
	c.Token = loginResp.AccessToken
	c.TokenExpiry = time.Now().Add(time.Duration(loginResp.TokenTtl) * time.Second).Unix()

	return &loginResp, nil
}

func (c *Client) GetConfig(dataID, group string) (*Config, error) {
	params := url.Values{}
	params.Set("dataId", dataID)
	params.Set("group", group)
	if c.Namespace != "" {
		params.Set("tenant", c.Namespace)
	}

	// 使用Nacos v1 API
	url := c.ServerURL + "/nacos/v1/cs/configs?" + params.Encode()
	fmt.Printf("获取配置URL: %s\n", url)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取配置失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("配置不存在")
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Config{
		DataID:  dataID,
		Group:   group,
		Content: string(content),
	}, nil
}

func (c *Client) PublishConfig(config *Config) error {
	data := url.Values{}
	data.Set("dataId", config.DataID)
	data.Set("group", config.Group)
	data.Set("content", config.Content)
	if config.Type != "" {
		data.Set("type", config.Type)
	}
	if c.Namespace != "" {
		data.Set("tenant", c.Namespace)
	}

	// 使用Nacos v1 API
	url := c.ServerURL + "/nacos/v1/cs/configs"
	fmt.Printf("发布配置URL: %s\n", url)

	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("发布配置失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("发布配置失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) DeleteConfig(dataID, group string) error {
	params := url.Values{}
	params.Set("dataId", dataID)
	params.Set("group", group)
	if c.Namespace != "" {
		params.Set("tenant", c.Namespace)
	}

	// 使用Nacos v1 API
	url := c.ServerURL + "/nacos/v1/cs/configs?" + params.Encode()
	fmt.Printf("删除配置URL: %s\n", url)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("删除配置失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("删除配置失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *Client) ListConfigs(pageNo, pageSize int) ([]Config, error) {
	// 使用Nacos v1 API的配置查询接口，根据curl命令构建请求
	params := url.Values{}
	params.Set("dataId", "") // 空字符串表示查询所有
	params.Set("group", "")  // 空字符串表示查询所有
	params.Set("appName", "")
	params.Set("config_tags", "")
	params.Set("pageNo", fmt.Sprintf("%d", pageNo))
	params.Set("pageSize", fmt.Sprintf("%d", pageSize))
	params.Set("search", "blur") // 使用模糊搜索
	if c.Namespace != "" {
		params.Set("tenant", c.Namespace)
	}
	if c.Token != "" {
		params.Set("accessToken", c.Token)
	}
	if c.Username != "" {
		params.Set("username", c.Username)
	}

	// 使用Nacos v1 API的正确端点
	url := c.ServerURL + "/nacos/v1/cs/configs?" + params.Encode()

	// 使用Nacos v1 API的正确端点

	// 使用GET方法调用Nacos v1配置查询API
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("User-Agent", "nacos-cli")
	req.Header.Set("Accept", "application/json")
	if c.Token != "" {
		req.Header.Set("accessToken", c.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取配置列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 检查HTTP状态码
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("获取配置列表失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应

	// Nacos v1 API响应格式
	var result struct {
		TotalCount     int      `json:"totalCount"`
		PageNumber     int      `json:"pageNumber"`
		PagesAvailable int      `json:"pagesAvailable"`
		PageItems      []Config `json:"pageItems"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析配置列表失败: %w, 原始响应: %s", err, string(body))
	}

	return result.PageItems, nil
}

// ListNamespaces 获取命名空间列表
func (c *Client) ListNamespaces() ([]Namespace, error) {
	// 使用Nacos v1 API
	url := c.ServerURL + "/nacos/v1/console/namespaces"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// 设置请求头
	req.Header.Set("User-Agent", "nacos-cli")
	req.Header.Set("Accept", "application/json")
	if c.Token != "" {
		req.Header.Set("accessToken", c.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取命名空间列表失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 检查HTTP状态码
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("获取命名空间列表失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Code    int         `json:"code"`
		Message string      `json:"message"`
		Data    []Namespace `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析命名空间列表失败: %w, 原始响应: %s", err, string(body))
	}

	if result.Code != 200 {
		return nil, fmt.Errorf("API调用失败: %s", result.Message)
	}

	return result.Data, nil
}

// CreateNamespace 创建命名空间
func (c *Client) CreateNamespace(namespaceId, namespaceName, namespaceDesc string) error {
	data := url.Values{}
	data.Set("customNamespaceId", namespaceId)
	data.Set("namespaceName", namespaceName)
	data.Set("namespaceDesc", namespaceDesc)

	// 使用Nacos v1 API
	url := c.ServerURL + "/nacos/v1/console/namespaces"

	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "nacos-cli")
	if c.Token != "" {
		req.Header.Set("accessToken", c.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("创建命名空间失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("创建命名空间失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteNamespace 删除命名空间
func (c *Client) DeleteNamespace(namespaceId string) error {
	params := url.Values{}
	params.Set("namespaceId", namespaceId)

	// 使用Nacos v1 API
	url := c.ServerURL + "/nacos/v1/console/namespaces?" + params.Encode()

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "nacos-cli")
	if c.Token != "" {
		req.Header.Set("accessToken", c.Token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("删除命名空间失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("删除命名空间失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	return nil
}