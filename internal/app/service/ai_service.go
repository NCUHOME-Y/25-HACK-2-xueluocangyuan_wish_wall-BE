package service

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/NCUHOME-Y/25-HACK-2-xueluocangyuan_wish_wall-BE/internal/pkg/logger"

	"github.com/sashabaranov/go-openai"
)

var (
	systemPrompt = `你是一个非常严格的中文许愿墙内容审核员。
你的唯一任务是判断用户提交的信息是否包含任何形式的：
1. 色情或低俗内容
2. 暴力或血腥
3. 辱骂、人身攻击或仇恨言论
4. 政治敏感或违法信息
5. 自残或自杀意图
6. 任何其他不适合在公共场合展示的不当言论

请仔细阅读 [用户信息]，然后**只回答一个词**:
- 如果内容**安全**，请回答 "true"
- 如果内容**不安全或违规**，请回答 "false"
`
)

// CheckContent(handler调用的函数)
// isViolating= true 不安全，丢弃
// isViolating= false 安全，接受
func CheckContent(content string) (isViolating bool, err error) {
	const maxContentLength = 1000
	trimmedContent := strings.TrimSpace(content)
	if trimmedContent == "" {
		logger.Log.Warnw("内容审核失败:内容为空", "content", content)
		return true, fmt.Errorf("内容不能为空")
	}
	if len([]rune(trimmedContent)) > maxContentLength {
		logger.Log.Warnw("内容审核失败:内容过长", "content", content, "maxLength", maxContentLength)
		return true, fmt.Errorf("内容长度不能超过%d个字符", maxContentLength)
	}
	apiKey := os.Getenv("SILICONFLOW_API_KEY") //从环境变量读取API Key
	if apiKey == "" {
		logger.Log.Errorw("AI内容审核失败:环境变量 SILICONFLOW_API_KEY 未设置", "env_var", "SILICONFLOW_API_KEY")
		return true, fmt.Errorf("AI service not configured")
	}

	config := openai.DefaultConfig(apiKey)
	baseURL := os.Getenv("SILICONFLOW_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.siliconflow.cn/v1"
	}
	config.BaseURL = baseURL

	client := openai.NewClientWithConfig(config)

	ctx := context.Background()

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("[用户愿望]: %s", content),
		},
	}

	//发送给AI
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       "Qwen/Qwen3-VL-8B-Instruct",
		Messages:    messages,
		Temperature: 0.0, //需要确定的true或false，不能有随机性
	},
	)

	if err != nil {
		logger.Log.Errorw("Silicon Flow API请求失败", "error", err)
		return true, err
	}

	//解析AI回答
	if len(resp.Choices) == 0 {
		logger.Log.Warnw("AI返回空内容,无法判断安全性", "content", content)
		return true, fmt.Errorf("AI返回空内容,无法判断安全性")
	}

	respText := resp.Choices[0].Message.Content

	//判断错误
	respTextTrimmed := strings.TrimSpace(strings.ToLower(respText))
	if respTextTrimmed == "false" {
		logger.Log.Infow("AI内容审核:不安全愿望被丢弃", "content", content)
		return true, nil
	}
	//正确
	if respTextTrimmed == "true" {
		logger.Log.Infow("AI内容审核:安全愿望被接受", "content", content)
		return false, nil
	}
	//无法判断
	logger.Log.Warnw("AI内容审核:无法判断愿望安全性,默认丢弃", "content", content, "AI回复", respText)
	return true, fmt.Errorf("AI无法判断内容安全性")
}
