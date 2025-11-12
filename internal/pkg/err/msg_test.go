package err

import (
	"testing"
)

// TestGetMsgTable 采用表驱动测试，覆盖所有已定义的错误码与默认回退逻辑
func TestGetMsgTable(t *testing.T) {
	// 遍历 MsgFlags，检查 GetMsg(code) 与映射值一致
	for code, expected := range MsgFlags {
		got := GetMsg(code)
		if got != expected {
			t.Fatalf("code %d expected '%s' got '%s'", code, expected, got)
		}
	}

	// 检查成功码单独正确（属于 MsgFlags 但显式再测一次以强调语义）
	if GetMsg(SUCCESS) != "成功" {
		t.Fatalf("SUCCESS code expected '成功' got '%s'", GetMsg(SUCCESS))
	}

	//  未知错误码应回退到 ERROR_SERVER_ERROR 对应的消息
	unknownCodes := []int{0, -1, 9999, 123456}
	fallback := MsgFlags[ERROR_SERVER_ERROR]
	for _, c := range unknownCodes {
		if GetMsg(c) != fallback {
			t.Fatalf("unknown code %d expected fallback '%s' got '%s'", c, fallback, GetMsg(c))
		}
	}
}
