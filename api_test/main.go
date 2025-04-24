package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 测试配置结构
type TestConfig struct {
	BaseURL   string      `json:"base_url"`
	TestCases []*TestCase `json:"test_cases"`
}

// 测试用例结构
type TestCase struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Skip        bool                   `json:"skip"`
	Depends     []string               `json:"depends_on"`
	API         API                    `json:"api"`
	Request     map[string]interface{} `json:"request"`
	FileUpload  *FileUpload            `json:"file_upload"`
	Expect      ExpectedResponse       `json:"expect"`
	Store       map[string]string      `json:"store"` // 存储响应中的值
}

// API定义
type API struct {
	Endpoint string `json:"endpoint"`
	Method   string `json:"method"`
	Auth     bool   `json:"auth_required"`
}

// 文件上传配置
type FileUpload struct {
	FieldName string `json:"field_name"`
	FileName  string `json:"file_name"`
	FilePath  string `json:"file_path"`
}

// 期望的响应
type ExpectedResponse struct {
	StatusCode  int                    `json:"status_code"`
	Contains    []string               `json:"contains"`
	NotContains []string               `json:"not_contains"`
	JSON        map[string]interface{} `json:"json"`
}

// 测试上下文，保存测试过程中的信息
type TestContext struct {
	Variables   map[string]interface{}
	TestResults map[string]bool
	AuthToken   string
}

func NewTestContext() *TestContext {
	return &TestContext{
		Variables:   make(map[string]interface{}),
		TestResults: make(map[string]bool),
	}
}

func main() {
	configFile := flag.String("config", "test_config.json", "Test configuration file")
	testName := flag.String("test", "", "Run specific test by name")
	flag.Parse()

	// 读取配置文件
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 创建测试上下文
	ctx := NewTestContext()

	// 运行测试
	if *testName != "" {
		// 运行指定测试
		found := false
		for _, tc := range config.TestCases {
			if tc.Name == *testName {
				found = true
				runTest(config, tc, ctx)
				break
			}
		}
		if !found {
			log.Fatalf("Test with name '%s' not found", *testName)
		}
	} else {
		// 运行所有测试
		runAllTests(config, ctx)
	}
}

// 加载测试配置
func loadConfig(configFile string) (*TestConfig, error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config TestConfig
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// 运行所有测试
func runAllTests(config *TestConfig, ctx *TestContext) {
	total := 0
	passed := 0
	skipped := 0
	failed := 0

	for _, tc := range config.TestCases {
		if tc.Skip {
			log.Printf("[SKIP] %s", tc.Name)
			skipped++
			continue
		}

		total++
		result := runTest(config, tc, ctx)
		if result {
			passed++
		} else {
			failed++
		}
	}

	log.Printf("\n=== Test Summary ===")
	log.Printf("Total tests: %d", total)
	log.Printf("Passed: %d", passed)
	log.Printf("Failed: %d", failed)
	log.Printf("Skipped: %d", skipped)
}

// 运行单个测试
func runTest(config *TestConfig, tc *TestCase, ctx *TestContext) bool {
	log.Printf("\n=== Running Test: %s ===", tc.Name)
	log.Printf("Description: %s", tc.Description)

	// 检查依赖的测试是否已通过
	for _, dep := range tc.Depends {
		if !ctx.TestResults[dep] {
			log.Printf("[SKIP] Dependency '%s' failed or was not run", dep)
			return false
		}
	}

	// 准备请求
	url := config.BaseURL + tc.API.Endpoint
	var reqBody []byte
	var err error

	// 替换请求中的变量
	processedRequest := make(map[string]interface{})
	for k, v := range tc.Request {
		if strVal, ok := v.(string); ok && strings.HasPrefix(strVal, "${") && strings.HasSuffix(strVal, "}") {
			varName := strVal[2 : len(strVal)-1]
			if value, exists := ctx.Variables[varName]; exists {
				processedRequest[k] = value
			} else {
				log.Printf("[WARN] Variable '%s' not found in context", varName)
				processedRequest[k] = v
			}
		} else {
			processedRequest[k] = v
		}
	}

	var req *http.Request

	if tc.FileUpload != nil {
		// 处理文件上传
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		// 添加表单字段
		for k, v := range processedRequest {
			field, err := writer.CreateFormField(k)
			if err != nil {
				log.Printf("[ERROR] Failed to create form field: %v", err)
				return false
			}

			// 将值转换为字符串
			var fieldValue string
			switch val := v.(type) {
			case string:
				fieldValue = val
			case float64:
				fieldValue = fmt.Sprintf("%v", val)
			case bool:
				fieldValue = fmt.Sprintf("%v", val)
			default:
				jsonVal, _ := json.Marshal(val)
				fieldValue = string(jsonVal)
			}

			field.Write([]byte(fieldValue))
		}

		// 添加文件
		file, err := os.Open(tc.FileUpload.FilePath)
		if err != nil {
			log.Printf("[ERROR] Failed to open file: %v", err)
			return false
		}
		defer file.Close()

		part, err := writer.CreateFormFile(tc.FileUpload.FieldName, filepath.Base(tc.FileUpload.FileName))
		if err != nil {
			log.Printf("[ERROR] Failed to create form file: %v", err)
			return false
		}
		_, err = io.Copy(part, file)
		if err != nil {
			log.Printf("[ERROR] Failed to copy file content: %v", err)
			return false
		}

		err = writer.Close()
		if err != nil {
			log.Printf("[ERROR] Failed to close multipart writer: %v", err)
			return false
		}

		req, err = http.NewRequest(tc.API.Method, url, body)
		if err != nil {
			log.Printf("[ERROR] Failed to create request: %v", err)
			return false
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())

	} else {
		// 处理JSON请求
		reqBody, err = json.Marshal(processedRequest)
		if err != nil {
			log.Printf("[ERROR] Failed to marshal request body: %v", err)
			return false
		}

		req, err = http.NewRequest(tc.API.Method, url, bytes.NewBuffer(reqBody))
		if err != nil {
			log.Printf("[ERROR] Failed to create request: %v", err)
			return false
		}
		req.Header.Set("Content-Type", "application/json")
	}

	// 添加认证头
	if tc.API.Auth && ctx.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.AuthToken)
	}

	// 发送请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[ERROR] Failed to send request: %v", err)
		return false
	}
	defer resp.Body.Close()

	// 读取响应
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read response body: %v", err)
		return false
	}

	bodyStr := string(bodyBytes)
	var jsonResp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &jsonResp); err != nil {
		log.Printf("[WARN] Response is not valid JSON: %s", bodyStr)
	}

	// 验证响应
	// 检查状态码
	if resp.StatusCode != tc.Expect.StatusCode {
		log.Printf("[FAIL] Expected status code %d but got %d", tc.Expect.StatusCode, resp.StatusCode)
		log.Printf("Response: %s", bodyStr)
		return false
	}

	// 检查响应内容包含指定字符串
	for _, s := range tc.Expect.Contains {
		if !strings.Contains(bodyStr, s) {
			log.Printf("[FAIL] Response does not contain expected string: %s", s)
			log.Printf("Response: %s", bodyStr)
			return false
		}
	}

	// 检查响应内容不包含指定字符串
	for _, s := range tc.Expect.NotContains {
		if strings.Contains(bodyStr, s) {
			log.Printf("[FAIL] Response contains unexpected string: %s", s)
			log.Printf("Response: %s", bodyStr)
			return false
		}
	}

	// 检查JSON响应
	if len(tc.Expect.JSON) > 0 && jsonResp != nil {
		for k, v := range tc.Expect.JSON {
			// 支持深层次访问，如 "data.user.name"
			keys := strings.Split(k, ".")
			var current interface{} = jsonResp

			// 导航到嵌套键
			for i, key := range keys {
				if m, ok := current.(map[string]interface{}); ok {
					if val, exists := m[key]; exists {
						current = val
					} else {
						log.Printf("[FAIL] Key '%s' not found in response JSON", strings.Join(keys[:i+1], "."))
						log.Printf("Response: %s", bodyStr)
						return false
					}
				} else {
					log.Printf("[FAIL] Cannot access key '%s' in response JSON", strings.Join(keys[:i], "."))
					log.Printf("Response: %s", bodyStr)
					return false
				}
			}

			// 转换预期值以便比较
			var expectedValue interface{}
			switch val := v.(type) {
			case string:
				expectedValue = val
			case float64:
				expectedValue = val
			case bool:
				expectedValue = val
			default:
				expectedValue = v
			}

			// 检查值是否匹配
			if fmt.Sprintf("%v", current) != fmt.Sprintf("%v", expectedValue) {
				log.Printf("[FAIL] Expected '%s' to be '%v' but got '%v'", k, expectedValue, current)
				log.Printf("Response: %s", bodyStr)
				return false
			}
		}
	}

	// 存储响应值
	for key, jsonPath := range tc.Store {
		if jsonResp == nil {
			log.Printf("[WARN] Cannot store values, response is not valid JSON")
			continue
		}

		// 支持深层次访问，如 "data.token"
		keys := strings.Split(jsonPath, ".")
		var current interface{} = jsonResp

		// 导航到嵌套键
		for i, k := range keys {
			if m, ok := current.(map[string]interface{}); ok {
				if val, exists := m[k]; exists {
					current = val
				} else {
					log.Printf("[WARN] Key '%s' not found in response JSON", strings.Join(keys[:i+1], "."))
					break
				}
			} else {
				log.Printf("[WARN] Cannot access key '%s' in response JSON", strings.Join(keys[:i], "."))
				break
			}
		}

		// 存储值
		ctx.Variables[key] = current

		// 特别处理token
		if key == "token" {
			ctx.AuthToken = fmt.Sprintf("%v", current)
		}
	}

	log.Printf("[PASS] Test passed")
	ctx.TestResults[tc.Name] = true
	return true
}
