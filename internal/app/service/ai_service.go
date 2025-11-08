package service

import (
	"fmt"
	//"errors"//引入 "errors"，用于模拟未来的错误返回
)

// AIService 提供与 AI 相关的服务
type AIService struct {
	// 这里可以添加一些字段，比如配置、依赖等
}

// NewAIService 创建一个新的 AIService 实例
// 其他服务，如be2的wishservice，可以通过调用这个函数来获取AIService实例
func NewAIService() *AIService {
	return &AIService{}
}

// IsContentSafe 检查给定内容是否安全
// mock实现，未来会调用实际的AI内容审核服务
func (s *AIService) IsContentSafe(content string) (bool, error) {
	// 模拟内容审核逻辑,未来真正调用AI SDK
	fmt.Printf("ai审核mock:正在审核内容:%s\n", content)
	//返回true表示内容安全，nil表示没有错误
	return true, nil
}
