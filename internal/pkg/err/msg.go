package err

const (
	SUCCESS = 200
	//登陆注册
	ERROR_LOGIN_FAILED = 1

	//鉴权
	ERROR_UNAUTHORIZED = 3

	//参数错误
	ERROR_PARAM_INVALID = 4

	//服务器错误
	ERROR_SERVER_ERROR = 10

	//资源不存在
	ERROR_WISH_DELETED      = 11
	ERROR_WISH_NOT_FOUND    = 12
	ERROR_FORBIDDEN_COMMENT = 13

	ERROR_LIKE_FAILED = 10005
)

// MsgFlags是一个code，message的映射
var MsgFlags = map[int]string{
	SUCCESS:                 "成功",
	ERROR_LOGIN_FAILED:      "登录失败，用户名或密码错误",
	ERROR_UNAUTHORIZED:      "登录已过期，请重新登录",
	ERROR_PARAM_INVALID:     "参数验证失败",
	ERROR_SERVER_ERROR:      "服务器错误",
	ERROR_WISH_DELETED:      "该心愿不存在或已被删除",
	ERROR_WISH_NOT_FOUND:    "未找到指定心愿",
	ERROR_FORBIDDEN_COMMENT: "禁止评论",
	ERROR_LIKE_FAILED:       "点赞失败，请稍后再试",
}

// GetMsg 获取错误码对应的信息
func GetMsg(code int) string {
	msg, ok := MsgFlags[code]
	if ok {
		return msg
	}
	//如果错误码不存在，返回默认信息
	return MsgFlags[ERROR_SERVER_ERROR]
}
