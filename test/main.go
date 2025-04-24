package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	Auth     bool   `json:"auth"`
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

// 初始化随机数生成器
func init() {
	// 设置日志输出到控制台
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 初始化随机数生成器
	rand.Seed(time.Now().UnixNano())

	// 确保当前工作目录是test目录
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	// 切换工作目录到test目录
	if filepath.Base(dir) != "test" {
		testDir := filepath.Join(dir, "test")
		if _, err := os.Stat(testDir); !os.IsNotExist(err) {
			os.Chdir(testDir)
		}
	}
}

// 生成指定长度的随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 生成只包含字母数字的随机字符串
func randomAlphaNum(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 生成随机用户信息
func generateRandomUserInfo() map[string]string {
	randSuffix := randomString(8)
	groupKey := randomAlphaNum(12)
	return map[string]string{
		"USERNAME": "testuser" + randSuffix,
		"EMAIL":    "test" + randSuffix + "@example.com",
		"PHONE":    "138" + randomString(8),
		"GROUPKEY": groupKey,
	}
}

func main() {
	log.Printf("测试程序启动...")
	configFile := flag.String("config", "test_config.json", "测试配置文件路径")
	testName := flag.String("test", "", "指定测试用例名称")
	skipEnvCheck := flag.Bool("skip-env", false, "跳过环境检测")
	flag.Parse()

	// 读取配置文件
	log.Printf("加载配置文件: %s", *configFile)
	config, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}
	log.Printf("成功加载测试用例数: %d", len(config.TestCases))

	// 环境检测
	if !*skipEnvCheck {
		log.Printf("正在进行环境检测...")
		if !RunEnvironmentCheck(config.BaseURL) {
			log.Fatal("环境检测失败，测试终止")
		}
		log.Printf("环境检测通过")
	}

	// 创建测试上下文
	ctx := NewTestContext()
	log.Printf("测试上下文已创建")

	// 运行测试
	if *testName != "" {
		// 运行指定测试
		log.Printf("准备运行指定测试: %s", *testName)
		found := false
		for _, tc := range config.TestCases {
			if tc.Name == *testName {
				found = true
				runTest(config, tc, ctx)
				break
			}
		}
		if !found {
			log.Fatalf("未找到测试用例: %s", *testName)
		}
	} else {
		// 运行所有测试
		log.Printf("准备运行所有测试...")
		runAllTests(config, ctx)
	}

	log.Printf("测试程序执行完毕")
}

// CheckServerReady 检查后端服务是否准备就绪
func CheckServerReady(baseURL string) bool {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// 移除/health路径，只检查base URL是否可以访问
	serverURL := strings.TrimSuffix(baseURL, "/api/oss")
	resp, err := client.Get(serverURL)
	if err != nil {
		log.Printf("服务器连接失败: %v", err)
		return false
	}
	defer resp.Body.Close()

	log.Printf("服务器准备就绪，状态码: %d", resp.StatusCode)
	return true
}

// CheckTestFilesExist 检查测试所需文件是否存在
func CheckTestFilesExist() bool {
	files := []string{
		"test_config.json",
		"test_file.txt",
	}

	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			log.Printf("测试文件不存在: %s", file)
			return false
		}
	}

	log.Printf("所有测试文件存在")
	return true
}

// RunEnvironmentCheck 主环境检测函数
func RunEnvironmentCheck(baseURL string) bool {
	fmt.Println("===== 开始环境检测 =====")

	serverReady := CheckServerReady(baseURL)
	filesExist := CheckTestFilesExist()

	fmt.Println("===== 环境检测完成 =====")

	return serverReady && filesExist
}

// 加载测试配置并替换模板变量
func loadConfig(configFile string) (*TestConfig, error) {
	file, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// 生成随机变量
	randomVars := generateRandomUserInfo()

	// 替换模板变量
	content := string(file)
	for key, value := range randomVars {
		pattern := fmt.Sprintf("{{%s}}", key)
		content = strings.ReplaceAll(content, pattern, value)
	}

	log.Printf("随机用户信息: %v", randomVars)

	// 解析配置
	var config TestConfig
	err = json.Unmarshal([]byte(content), &config)
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
		log.Printf("\n=== Running Test: %s ===", tc.Name)
		log.Printf("Description: %s", tc.Description)

		// 检查依赖的测试是否已通过
		dependsFailed := false
		for _, dep := range tc.Depends {
			if !ctx.TestResults[dep] {
				log.Printf("[SKIP] Dependency '%s' failed or was not run", dep)
				dependsFailed = true
				break
			}
		}

		if dependsFailed {
			failed++
			continue
		}

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
	// 替换URL中的变量
	url := config.BaseURL + replaceVars(tc.API.Endpoint, ctx)

	// 准备请求体
	var reqBody []byte
	var err error

	// 处理请求参数中的变量
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
			fmt.Fprintf(field, "%v", v)
		}

		// 添加文件
		file, err := os.Open(tc.FileUpload.FilePath)
		if err != nil {
			log.Printf("[ERROR] Failed to open file %s: %v", tc.FileUpload.FilePath, err)
			return false
		}
		defer file.Close()

		part, err := writer.CreateFormFile(tc.FileUpload.FieldName, tc.FileUpload.FileName)
		if err != nil {
			log.Printf("[ERROR] Failed to create form file: %v", err)
			return false
		}
		_, err = io.Copy(part, file)
		if err != nil {
			log.Printf("[ERROR] Failed to copy file content: %v", err)
			return false
		}

		writer.Close()
		req, err = http.NewRequest(tc.API.Method, url, body)
		if err != nil {
			log.Printf("[ERROR] Failed to create request: %v", err)
			return false
		}
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else {
		// 普通请求
		reqBody, err = json.Marshal(processedRequest)
		if err != nil {
			log.Printf("[ERROR] Failed to marshal request: %v", err)
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
		log.Printf("[ERROR] Request failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read response: %v", err)
		return false
	}

	// 检查状态码
	if resp.StatusCode != tc.Expect.StatusCode {
		log.Printf("[FAIL] Expected status code %d but got %d", tc.Expect.StatusCode, resp.StatusCode)
		log.Printf("Response: %s", string(respBody))
		return false
	}

	// 检查响应内容包含的字符串
	respStr := string(respBody)
	for _, str := range tc.Expect.Contains {
		if !strings.Contains(respStr, str) {
			log.Printf("[FAIL] Response does not contain '%s'", str)
			log.Printf("Response: %s", respStr)
			return false
		}
	}

	// 检查响应不应包含的字符串
	for _, str := range tc.Expect.NotContains {
		if strings.Contains(respStr, str) {
			log.Printf("[FAIL] Response contains '%s' but should not", str)
			log.Printf("Response: %s", respStr)
			return false
		}
	}

	// 解析JSON响应
	var respJSON map[string]interface{}
	if err := json.Unmarshal(respBody, &respJSON); err != nil {
		log.Printf("[ERROR] Failed to parse JSON response: %v", err)
		log.Printf("Response: %s", respStr)
		return false
	}

	// 记录整个响应JSON
	jsonBytes, _ := json.MarshalIndent(respJSON, "", "  ")
	log.Printf("[DEBUG] Response JSON: %s", jsonBytes)

	// 验证JSON响应
	for path, expectedValue := range tc.Expect.JSON {
		parts := strings.Split(path, ".")
		value := getNestedValue(respJSON, parts)
		if !compareValues(value, expectedValue) {
			log.Printf("[FAIL] Expected '%s' to be '%v' but got '%v'", path, expectedValue, value)
			log.Printf("Response: %s", respStr)
			return false
		}
	}

	// 存储需要的值
	for name, path := range tc.Store {
		parts := strings.Split(path, ".")
		value := getNestedValue(respJSON, parts)
		if value != nil {
			ctx.Variables[name] = value
			log.Printf("[INFO] Stored '%s' = '%v'", name, value)
		} else {
			log.Printf("[WARN] Failed to extract value for path '%s'", path)
			log.Printf("Response JSON: %v", respJSON)
		}
	}

	// 存储认证令牌
	if tc.Name == "login_user" && ctx.Variables["token"] != nil {
		ctx.AuthToken = fmt.Sprintf("%v", ctx.Variables["token"])
		log.Printf("[INFO] Auth token stored")
	}

	// 测试通过
	log.Printf("[PASS] Test passed")
	ctx.TestResults[tc.Name] = true
	return true
}

// 替换字符串中的变量
func replaceVars(str string, ctx *TestContext) string {
	for varName, value := range ctx.Variables {
		str = strings.ReplaceAll(str, "${"+varName+"}", fmt.Sprintf("%v", value))
	}
	return str
}

// 获取嵌套的JSON值
func getNestedValue(data map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return nil
	}

	// 处理当前层级
	current := path[0]

	// 处理数组索引，如 data.0.id 格式
	if index, err := strconv.Atoi(current); err == nil {
		// 如果当前路径是数字，尝试作为数组索引处理
		if arr, ok := data["data"].([]interface{}); ok && index >= 0 && index < len(arr) {
			if len(path) == 1 {
				return arr[index]
			}

			// 如果数组元素是map，继续递归处理
			if nestedMap, ok := arr[index].(map[string]interface{}); ok {
				return getNestedValue(nestedMap, path[1:])
			}
			return nil
		}
	}

	// 常规的对象属性访问
	if len(path) == 1 {
		return data[current]
	}

	// 递归处理嵌套对象
	if nested, ok := data[current].(map[string]interface{}); ok {
		return getNestedValue(nested, path[1:])
	}

	return nil
}

// 比较两个值是否相等
func compareValues(actual, expected interface{}) bool {
	// 将字符串形式的数字转换为数字进行比较
	if strActual, okA := actual.(string); okA {
		if numExpected, okE := expected.(float64); okE {
			if numActual, err := json.Number(strActual).Float64(); err == nil {
				return numActual == numExpected
			}
		}
	}
	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}
