package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 处理文件上传的测试用例
func executeFileUploadRequest(testCase *TestCase, baseURL string, ctx *TestContext) (*http.Response, []byte, error) {
	if testCase.FileUpload == nil {
		return nil, nil, fmt.Errorf("未指定文件上传配置")
	}

	// 替换变量
	endpoint := testCase.API.Endpoint
	for k, v := range ctx.Variables {
		if strVal, ok := v.(string); ok {
			endpoint = strings.ReplaceAll(endpoint, "${"+k+"}", strVal)
		}
	}

	// 构建URL
	url := fmt.Sprintf("%s%s", baseURL, endpoint)

	// 创建一个缓冲区用于存储文件内容
	var requestBody bytes.Buffer
	// 创建一个multipart writer
	writer := multipart.NewWriter(&requestBody)

	// 替换文件内容中的动态变量
	filePath := testCase.FileUpload.FilePath
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("读取文件失败: %v", err)
	}

	// 替换文件内容中的变量
	fileContentStr := string(fileContent)
	// 替换日期
	fileContentStr = strings.Replace(fileContentStr, "{{DATE}}", time.Now().Format("2006-01-02 15:04:05"), -1)
	// 替换随机ID
	fileContentStr = strings.Replace(fileContentStr, "{{RANDOM_ID}}", randomString(10), -1)

	// 创建表单字段
	for k, v := range testCase.Request {
		if strVal, ok := v.(string); ok {
			if strings.HasPrefix(strVal, "${") && strings.HasSuffix(strVal, "}") {
				varName := strVal[2 : len(strVal)-1]
				if val, ok := ctx.Variables[varName]; ok {
					err = writer.WriteField(k, fmt.Sprintf("%v", val))
				} else {
					err = writer.WriteField(k, strVal)
				}
			} else {
				err = writer.WriteField(k, strVal)
			}
		} else {
			err = writer.WriteField(k, fmt.Sprintf("%v", v))
		}
		if err != nil {
			return nil, nil, fmt.Errorf("写入表单字段失败: %v", err)
		}
	}

	// 创建文件表单字段
	fileFieldName := testCase.FileUpload.FieldName
	if fileFieldName == "" {
		fileFieldName = "file"
	}

	fileName := testCase.FileUpload.FileName
	if fileName == "" {
		fileName = filepath.Base(filePath)
	}

	// 创建文件表单字段
	part, err := writer.CreateFormFile(fileFieldName, fileName)
	if err != nil {
		return nil, nil, fmt.Errorf("创建文件表单字段失败: %v", err)
	}

	// 写入文件内容
	_, err = part.Write([]byte(fileContentStr))
	if err != nil {
		return nil, nil, fmt.Errorf("写入文件内容失败: %v", err)
	}

	// 关闭writer
	err = writer.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("关闭writer失败: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest(testCase.API.Method, url, &requestBody)
	if err != nil {
		return nil, nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置Content-Type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 设置授权头
	if testCase.API.Auth && ctx.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+ctx.AuthToken)
	}

	// 执行请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("执行请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("读取响应失败: %v", err)
	}

	return resp, respBody, nil
}
