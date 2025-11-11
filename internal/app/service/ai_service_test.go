package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

func TestCheckContent(t *testing.T) {
	logger.InitLogger()

	//设置一个模拟silicon flow api服务器
	//mockapiresponse用来控制这个假服务器返回什么内容
	var mockAPIResponse openai.ChatCompletionResponse

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		//发送模拟Json响应
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockAPIResponse)
	}))
	defer mockServer.Close()

	// 覆盖环境变量，使用模拟服务器的URL
	os.Setenv("SILICONFLOW_API_KEY", "test-api-key")
	// 将 SILICONFLOW_BASE_URL 指向模拟服务器的 v1 路径
	os.Setenv("SILICONFLOW_BASE_URL", mockServer.URL+"/v1")

	testCases := []struct {
		name              string
		inputContent      string
		mockResponse      openai.ChatCompletionResponse
		expectedViolating bool
		expectedErrorMsg  string
	}{
		{
			name:              "内容为空",
			inputContent:      " ",
			mockResponse:      openai.ChatCompletionResponse{},
			expectedViolating: true,
			expectedErrorMsg:  "内容不能为空",
		},
		{
			name:              "内容过长",
			inputContent:      strings.Repeat("a", 1001),
			mockResponse:      openai.ChatCompletionResponse{},
			expectedViolating: true,
			expectedErrorMsg:  "内容长度不能超过1000个字符",
		},
		{
			name:         "内容安全",
			inputContent: "今天天气真好，我想去公园散步。",
			mockResponse: openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "true",
						},
					},
				},
			},
			expectedViolating: false,
			expectedErrorMsg:  "",
		},
		{
			name:         "内容不安全",
			inputContent: "我操我要跳楼。",
			mockResponse: openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: openai.ChatCompletionMessage{
							Content: "false",
						},
					},
				},
			},
			expectedViolating: true,
			expectedErrorMsg:  "",
		},
		{
			name:         "API返回无法判断",
			inputContent: "火星文",
			mockResponse: openai.ChatCompletionResponse{
				Choices: []openai.ChatCompletionChoice{}},
			expectedViolating: true,
			expectedErrorMsg:  "AI返回空内容,无法判断安全性",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 设置模拟API响应
			mockAPIResponse = tc.mockResponse

			// 调用要测试的函数
			isViolating, err := CheckContent(tc.inputContent)

			// 验证结果
			assert.Equal(t, tc.expectedViolating, isViolating)

			if tc.expectedErrorMsg != "" {
				// 期望返回错误
				if assert.Error(t, err, "应返回错误") {
					assert.Contains(t, err.Error(), tc.expectedErrorMsg, "错误信息应包含期望的文本")
				}
			} else {
				// 期望不返回错误
				assert.NoError(t, err, "不应返回错误")
			}
		})
	}

	// 清理：移除测试时设置的环境变量
	os.Unsetenv("SILICONFLOW_BASE_URL")
	os.Unsetenv("SILICONFLOW_API_KEY")
}
