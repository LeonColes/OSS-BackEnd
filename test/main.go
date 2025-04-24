package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Swagger文档结构
type SwaggerDoc struct {
	Swagger     string                       `json:"swagger"`
	Info        map[string]interface{}       `json:"info"`
	Host        string                       `json:"host"`
	BasePath    string                       `json:"basePath"`
	Paths       map[string]map[string]Path   `json:"paths"`
	Definitions map[string]SwaggerDefinition `json:"definitions"`
}

type Path struct {
	Tags        []string                   `json:"tags"`
	Summary     string                     `json:"summary"`
	Description string                     `json:"description"`
	OperationID string                     `json:"operationId"`
	Consumes    []string                   `json:"consumes"`
	Produces    []string                   `json:"produces"`
	Parameters  []Parameter                `json:"parameters"`
	Responses   map[string]SwaggerResponse `json:"responses"`
	Security    []map[string][]string      `json:"security"`
}

type Parameter struct {
	Name        string    `json:"name"`
	In          string    `json:"in"`
	Description string    `json:"description"`
	Required    bool      `json:"required"`
	Type        string    `json:"type"`
	Schema      SchemaRef `json:"schema"`
}

type SchemaRef struct {
	Ref string `json:"$ref"`
}

type SwaggerResponse struct {
	Description string    `json:"description"`
	Schema      SchemaRef `json:"schema"`
}

type SwaggerDefinition struct {
	Type       string                        `json:"type"`
	Properties map[string]PropertyDefinition `json:"properties"`
	Required   []string                      `json:"required"`
}

type PropertyDefinition struct {
	Type        string     `json:"type"`
	Format      string     `json:"format"`
	Description string     `json:"description"`
	Items       *SchemaRef `json:"items"`
	Ref         string     `json:"$ref"`
}

// 测试配置
type TestConfig struct {
	BaseURL   string      `json:"base_url"`
	TestCases []*TestCase `json:"test_cases"`
}

// 测试用例
type TestCase struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Skip        bool                   `json:"skip"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	API         API                    `json:"api"`
	Request     map[string]interface{} `json:"request"`
	FileUpload  *FileUpload            `json:"file_upload,omitempty"`
	Expect      ExpectedResponse       `json:"expect"`
	Store       map[string]string      `json:"store,omitempty"`
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
	Contains    []string               `json:"contains,omitempty"`
	NotContains []string               `json:"not_contains,omitempty"`
	JSON        map[string]interface{} `json:"json"`
}

// 存储变量的上下文
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

// 初始化随机数种子
func init() {
	rand.Seed(time.Now().UnixNano())
}

// 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 生成随机字符串(包含特殊字符)
func randomComplexString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 生成随机数字
func randomNumber(length int) string {
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 生成随机用户信息
func generateRandomUserInfo() map[string]string {
	randSuffix := randomString(8)
	return map[string]string{
		"NAME":     "testuser_" + randSuffix,
		"EMAIL":    "test_" + randSuffix + "@example.com",
		"PASSWORD": "Test@123" + randomComplexString(6), // 更复杂的密码
	}
}

// 从引用路径中提取定义名称
func extractDefinitionName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}

// 根据Schema生成请求数据
func generateDataFromSchema(schema SchemaRef, definitions map[string]SwaggerDefinition) map[string]interface{} {
	if schema.Ref == "" {
		return make(map[string]interface{})
	}

	defName := extractDefinitionName(schema.Ref)
	def, ok := definitions[defName]
	if !ok {
		return make(map[string]interface{})
	}

	data := make(map[string]interface{})
	for propName, propDef := range def.Properties {
		switch propDef.Type {
		case "string":
			if strings.Contains(propName, "email") {
				data[propName] = fmt.Sprintf("test_%s@example.com", randomString(5))
			} else if strings.Contains(propName, "password") {
				data[propName] = "Password123!" + randomString(4)
			} else if strings.Contains(propName, "username") || strings.Contains(propName, "name") {
				data[propName] = "test_" + randomString(8)
			} else if strings.Contains(propName, "phone") {
				data[propName] = "138" + randomNumber(8)
			} else {
				data[propName] = "test_" + randomString(5)
			}
		case "integer", "number":
			data[propName] = rand.Intn(100)
		case "boolean":
			data[propName] = rand.Intn(2) == 1
		case "array":
			// 为数组创建空数组或随机项
			data[propName] = []interface{}{}
		case "object":
			// 为对象递归生成数据
			if propDef.Ref != "" {
				subName := extractDefinitionName(propDef.Ref)
				if subDef, ok := definitions[subName]; ok {
					subData := make(map[string]interface{})
					for subPropName, subPropDef := range subDef.Properties {
						// 简单处理下嵌套对象
						if subPropDef.Type == "string" {
							subData[subPropName] = "nested_" + randomString(4)
						} else if subPropDef.Type == "integer" {
							subData[subPropName] = rand.Intn(100)
						}
					}
					data[propName] = subData
				}
			} else {
				data[propName] = map[string]interface{}{}
			}
		}
	}
	return data
}

// 从Swagger文档生成测试用例
func generateTestCases(swaggerDoc *SwaggerDoc) []*TestCase {
	testCases := []*TestCase{}

	// 首先创建用户注册和登录测试
	randomVars := generateRandomUserInfo()

	// 创建用户注册测试
	registerTest := &TestCase{
		Name:        "post_api_oss_user_register",
		Description: "用户注册",
		API: API{
			Endpoint: "/api/oss/user/register",
			Method:   "POST",
			Auth:     false,
		},
		Request: map[string]interface{}{
			"email":    randomVars["EMAIL"],
			"password": randomVars["PASSWORD"],
			"name":     randomVars["NAME"],
		},
		Expect: ExpectedResponse{
			StatusCode: 200,
			JSON: map[string]interface{}{
				"code": 200,
			},
		},
		Store: map[string]string{
			"USER_ID": "data.id",
		},
	}
	testCases = append(testCases, registerTest)

	// 创建用户登录测试，依赖于注册测试
	loginTest := &TestCase{
		Name:        "post_api_oss_user_login",
		Description: "用户登录",
		DependsOn:   []string{"post_api_oss_user_register"},
		API: API{
			Endpoint: "/api/oss/user/login",
			Method:   "POST",
			Auth:     false,
		},
		Request: map[string]interface{}{
			"email":    randomVars["EMAIL"],
			"password": randomVars["PASSWORD"],
		},
		Expect: ExpectedResponse{
			StatusCode: 200,
			JSON: map[string]interface{}{
				"code": 200,
			},
		},
		Store: map[string]string{
			"TOKEN": "data.token",
		},
	}
	testCases = append(testCases, loginTest)

	// 遍历Swagger文档中的所有路径
	for path, pathItem := range swaggerDoc.Paths {
		// 跳过注册和登录的API，因为我们已经手动创建了
		if strings.Contains(path, "/user/register") || strings.Contains(path, "/user/login") {
			continue
		}

		for method, operation := range pathItem {
			// 结构体不能与nil比较，检查是否为零值结构体
			if operation.Summary == "" && len(operation.Tags) == 0 {
				continue
			}

			// 跳过已经手动创建的注册和登录测试
			testName := strings.ToLower(method + path)
			if testName == "post_api_oss_user_register" || testName == "post_api_oss_user_login" {
				continue
			}

			testCase := &TestCase{
				Name:        testName,
				Description: operation.Summary,
				API: API{
					Endpoint: path,
					Method:   strings.ToUpper(method),
					Auth:     false,
				},
				Request: make(map[string]interface{}),
				Expect: ExpectedResponse{
					StatusCode: 200,
					JSON: map[string]interface{}{
						"code": 200,
					},
				},
				Store: make(map[string]string),
			}

			// 如果路径包含需要认证的关键词，设置依赖于登录测试和Auth标记
			if strings.Contains(path, "/user/") ||
				strings.Contains(path, "/group/") ||
				strings.Contains(path, "/project/") ||
				strings.Contains(path, "/file/") ||
				strings.Contains(path, "/share/") ||
				!strings.Contains(path, "/register") &&
					!strings.Contains(path, "/login") {
				testCase.DependsOn = []string{"post_api_oss_user_login"}
				testCase.API.Auth = true
			}

			// 创建请求体
			request := make(map[string]interface{})
			for _, param := range operation.Parameters {
				if param.In == "body" && param.Schema.Ref != "" {
					// 处理请求体
					bodyData := generateDataFromSchema(param.Schema, swaggerDoc.Definitions)
					for k, v := range bodyData {
						request[k] = v
					}
				} else if param.In == "query" && param.Required {
					// 处理查询参数
					switch param.Type {
					case "string":
						request[param.Name] = "query_" + randomString(5)
					case "integer", "number":
						request[param.Name] = rand.Intn(100)
					case "boolean":
						request[param.Name] = rand.Intn(2) == 1
					}
				} else if param.In == "path" && param.Name == "id" {
					// 处理路径参数
					request[param.Name] = "${USER_ID}"
				}
			}

			// 设置请求体
			testCase.Request = request

			// 处理创建操作的存储变量
			if strings.Contains(testName, "create") && method == "post" {
				testCase.Store[strings.Replace(testName, "create", "id", 1)] = "data.id"
			}

			testCases = append(testCases, testCase)
		}
	}

	return testCases
}

// 从Swagger文档生成测试配置
func generateTestConfig(swaggerDoc *SwaggerDoc, baseURL string) *TestConfig {
	testCases := generateTestCases(swaggerDoc)

	return &TestConfig{
		BaseURL:   baseURL,
		TestCases: testCases,
	}
}

// 执行HTTP请求
func executeRequest(testCase *TestCase, baseURL string, ctx *TestContext) (*http.Response, []byte, error) {
	// 替换变量
	endpoint := testCase.API.Endpoint
	for k, v := range ctx.Variables {
		if strVal, ok := v.(string); ok {
			endpoint = strings.ReplaceAll(endpoint, "${"+k+"}", strVal)
		}
	}

	// 构建URL
	url := fmt.Sprintf("%s%s", baseURL, endpoint)

	// 替换请求体中的变量
	requestBody := make(map[string]interface{})
	for k, v := range testCase.Request {
		if strVal, ok := v.(string); ok {
			if strings.HasPrefix(strVal, "${") && strings.HasSuffix(strVal, "}") {
				varName := strVal[2 : len(strVal)-1]
				if val, ok := ctx.Variables[varName]; ok {
					requestBody[k] = val
				}
			} else {
				requestBody[k] = strVal
			}
		} else {
			requestBody[k] = v
		}
	}

	// 创建请求
	var req *http.Request
	var err error

	if testCase.API.Method == "GET" {
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, nil, err
		}

		// 添加查询参数
		q := req.URL.Query()
		for k, v := range requestBody {
			q.Add(k, fmt.Sprintf("%v", v))
		}
		req.URL.RawQuery = q.Encode()
	} else {
		// 序列化请求体
		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return nil, nil, err
		}

		req, err = http.NewRequest(testCase.API.Method, url, strings.NewReader(string(jsonBody)))
		if err != nil {
			return nil, nil, err
		}
		req.Header.Set("Content-Type", "application/json")
	}

	// 设置授权头
	if testCase.API.Auth && ctx.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.AuthToken)
	}

	// 执行请求
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, respBody, nil
}

// 从响应中提取值
func extractValue(jsonData map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var current interface{} = jsonData

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			if val, exists := m[part]; exists {
				current = val
				continue
			}
		}
		return nil, false
	}

	return current, true
}

// 验证响应结果
func validateResponse(testCase *TestCase, resp *http.Response, body []byte, ctx *TestContext) (bool, error) {
	// 检查状态码
	if resp.StatusCode != testCase.Expect.StatusCode {
		return false, fmt.Errorf("状态码不匹配: 预期 %d, 实际 %d", testCase.Expect.StatusCode, resp.StatusCode)
	}

	// 检查响应体
	var respJSON map[string]interface{}
	if err := json.Unmarshal(body, &respJSON); err != nil {
		return false, fmt.Errorf("解析JSON响应失败: %v", err)
	}

	// 检查JSON期望值
	for path, expectedVal := range testCase.Expect.JSON {
		actualVal, found := extractValue(respJSON, path)
		if !found {
			return false, fmt.Errorf("未找到响应中的路径: %s", path)
		}

		// 检查值是否匹配
		if fmt.Sprintf("%v", expectedVal) != fmt.Sprintf("%v", actualVal) {
			return false, fmt.Errorf("值不匹配: 路径 %s, 预期 %v, 实际 %v", path, expectedVal, actualVal)
		}
	}

	// 检查包含的字符串
	respString := string(body)
	for _, str := range testCase.Expect.Contains {
		if !strings.Contains(respString, str) {
			return false, fmt.Errorf("响应中未包含预期的字符串: %s", str)
		}
	}

	// 检查不包含的字符串
	for _, str := range testCase.Expect.NotContains {
		if strings.Contains(respString, str) {
			return false, fmt.Errorf("响应中包含不应出现的字符串: %s", str)
		}
	}

	// 存储指定的变量
	for varName, path := range testCase.Store {
		if strings.HasPrefix(path, "request.") {
			// 从请求中提取值
			requestField := path[8:]
			if val, ok := testCase.Request[requestField]; ok {
				ctx.Variables[varName] = val
				continue
			}
		} else {
			// 从响应中提取值
			val, found := extractValue(respJSON, path)
			if found {
				ctx.Variables[varName] = val
				// 特殊处理token
				if varName == "TOKEN" {
					ctx.AuthToken = fmt.Sprintf("%v", val)
				}
			}
		}
	}

	return true, nil
}

// 检查服务可用性
func checkServiceAvailability(baseURL string) bool {
	// 简单发送GET请求到根路径检查服务是否可用
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(baseURL)
	if err != nil {
		log.Printf("后端服务不可用，错误: %v", err)
		return false
	}
	defer resp.Body.Close()

	log.Printf("后端服务状态检查: %d", resp.StatusCode)
	// 即使返回404或其他状态码，只要服务响应就认为是可用的
	return true
}

// 运行测试用例
func runTest(testCase *TestCase, baseURL string, ctx *TestContext) bool {
	log.Printf("执行测试: %s - %s", testCase.Name, testCase.Description)

	// 检查依赖
	for _, dep := range testCase.DependsOn {
		if passed, ok := ctx.TestResults[dep]; !ok || !passed {
			log.Printf("跳过测试 %s，因为依赖的测试 %s 未通过", testCase.Name, dep)
			return false
		}

		// 特殊处理登录测试依赖，确保获取到认证Token
		if dep == "post_api_oss_user_login" && testCase.API.Auth {
			if token, ok := ctx.Variables["TOKEN"]; ok {
				ctx.AuthToken = fmt.Sprintf("%v", token)
			} else {
				log.Printf("无法获取认证Token，可能导致测试失败")
			}
		}
	}

	// 执行请求
	var resp *http.Response
	var body []byte
	var err error

	// 判断是否为文件上传测试
	if testCase.FileUpload != nil {
		resp, body, err = executeFileUploadRequest(testCase, baseURL, ctx)
	} else {
		resp, body, err = executeRequest(testCase, baseURL, ctx)
	}

	if err != nil {
		log.Printf("执行请求失败: %v", err)
		return false
	}

	// 验证响应
	success, err := validateResponse(testCase, resp, body, ctx)
	if err != nil || !success {
		if err != nil {
			log.Printf("测试失败: %v", err)
		} else {
			log.Printf("测试失败: 响应验证不通过")
		}
		ctx.TestResults[testCase.Name] = false
		return false
	}

	log.Printf("测试通过: %s", testCase.Name)
	ctx.TestResults[testCase.Name] = true
	return true
}

// 运行所有测试
func runTests(config *TestConfig) {
	ctx := NewTestContext()
	successCount := 0
	failCount := 0

	// 检查后端服务可用性
	if !checkServiceAvailability(config.BaseURL) {
		log.Println("警告: 后端服务可能不可用，测试可能会失败")
	}

	// 统计测试用例总数
	totalTests := len(config.TestCases)
	log.Printf("开始执行测试: 共 %d 个测试用例", totalTests)

	// 执行测试
	for _, testCase := range config.TestCases {
		if testCase.Skip {
			log.Printf("跳过测试: %s", testCase.Name)
			continue
		}

		if runTest(testCase, config.BaseURL, ctx) {
			successCount++
		} else {
			failCount++
		}
	}

	// 输出测试统计
	log.Printf("测试完成: 共 %d 个测试用例, 通过 %d, 失败 %d", totalTests, successCount, failCount)
}

func main() {
	// 命令行参数
	swaggerFile := flag.String("swagger", "docs/swagger/swagger.json", "Swagger文档路径")
	baseURL := flag.String("base", "http://localhost:8080", "API基础URL")
	outputFile := flag.String("output", "test_config.json", "测试配置输出文件")
	runFlag := flag.Bool("run", false, "是否执行测试")
	flag.Parse()

	// 读取Swagger文档
	log.Printf("从文件读取Swagger文档: %s", *swaggerFile)
	data, err := ioutil.ReadFile(*swaggerFile)
	if err != nil {
		log.Fatalf("读取Swagger文件失败: %v", err)
	}

	// 解析Swagger文档
	var swaggerDoc SwaggerDoc
	err = json.Unmarshal(data, &swaggerDoc)
	if err != nil {
		log.Fatalf("解析Swagger文档失败: %v", err)
	}

	// 生成测试配置
	testConfig := generateTestConfig(&swaggerDoc, *baseURL)

	// 写入配置文件
	if *outputFile != "" {
		// 确保目录存在
		err = os.MkdirAll(filepath.Dir(*outputFile), 0755)
		if err != nil {
			log.Fatalf("创建输出目录失败: %v", err)
		}

		output, err := json.MarshalIndent(testConfig, "", "  ")
		if err != nil {
			log.Fatalf("序列化测试配置失败: %v", err)
		}

		err = ioutil.WriteFile(*outputFile, output, 0644)
		if err != nil {
			log.Fatalf("写入测试配置文件失败: %v", err)
		}

		log.Printf("成功生成测试配置文件: %s", *outputFile)
	}

	// 执行测试
	if *runFlag {
		runTests(testConfig)
	}
}
