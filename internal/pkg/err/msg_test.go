package err

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetMsg(t *testing.T) {
	//测试已知错误码
	msg := GetMsg(ERROR_LOGIN_FAILED)
	assert.Equal(t, "登录失败，用户名或密码错误", msg)

	// 使用已存在的常量名 ERROR_WISH_NOT_FOUND
	msg = GetMsg(ERROR_WISH_NOT_FOUND)
	// 与 MsgFlags 中的值保持一致
	assert.Equal(t, "未找到指定心愿", msg)

	msg = GetMsg(SUCCESS)
	assert.Equal(t, "成功", msg)

	// 测试未知错误码应返回服务器错误默认信息
	msg = GetMsg(9999)
	assert.Equal(t, "服务器错误", msg)
}
