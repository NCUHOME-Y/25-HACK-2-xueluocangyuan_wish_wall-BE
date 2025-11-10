package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert" // 断言库
)

func TestGenerateAndParseToken(t *testing.T) {
	// 设置测试用的 JWT_SECRET 环境变量
	os.Setenv("JWT_SECRET", "my_test_secret_key")

	var testUserID uint = 12345

	// 测试 GenerateToken 函数
	token, err := GenerateToken(testUserID)
	assert.NoError(t, err, "生成 token 时不应发生错误")
	assert.NotEmpty(t, token, "生成的 token 不应为空")

	// 测试 ParseToken 函数
	claims, err := ParseToken(token)
	assert.NoError(t, err, "解析 token 时不应发生错误")
	assert.NotNil(t, claims, "解析后的 claims 不应为 nil")
	assert.Equal(t, testUserID, claims.UserID, "解析出的 UserID 应该与原始 UserID 相同")
	assert.Equal(t, "wish_wall_app", claims.Issuer, "Issuer 应该是 'wish_wall_app'")

	// 测试一个无效 token
	invalidToken := "asdkhwjkjdn9jknjkzHDK3jdnkajuwJNDK,ACUkjandkawoqe1034;'52jq1ndajs"
	claims2, err := ParseToken(invalidToken)
	// 断言解析无效 token 时发生错误
	assert.Error(t, err, "解析无效 token 时应发生错误")
	assert.Nil(t, claims2, "解析无效 token 时 claims 应为 nil")
}

func TestGetJWTSecret(t *testing.T) {
	// 测试当环境变量未设置时，返回默认密钥
	os.Unsetenv("JWT_SECRET")
	secret := getJWTSecret()
	assert.Equal(t, []byte("default_secret_key"), secret)

	// 测试从环境变量读取
	os.Setenv("JWT_SECRET", "env_secret")
	secret = getJWTSecret()
	assert.Equal(t, []byte("env_secret"), secret)
}
