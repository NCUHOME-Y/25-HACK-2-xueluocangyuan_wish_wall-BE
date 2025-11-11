package err

const (
	SUCCESS = 200

	// --- 登录/权限 (1-3) ---
	// 401: 登录失败 (学号或密码错误)
	ERROR_LOGIN_FAILED = 1
	// 403: 无权删除此评论
	ERROR_FORBIDDEN_DELETE = 2
	// 401: 登录已过期 (默认)
	ERROR_UNAUTHORIZED = 3

	// --- 参数错误 (4) ---
	// 400/422: 参数验证失败
	ERROR_PARAM_INVALID = 4

	// --- 服务器错误 (10) ---
	// 500: 服务器错误 (默认) / 503: 服务暂时不可用
	ERROR_SERVER_ERROR = 10

	// --- 资源/业务逻辑 (11-14) ---
	// 410: 愿望不存在或已删除
	ERROR_WISH_DELETED = 11
	// 404: 愿望已删除或不存在 (用于点赞、评论、查互动)
	ERROR_WISH_NOT_FOUND = 12
	// 403: 该愿望不允许评论
	ERROR_FORBIDDEN_COMMENT = 13
	// 404: 评论不存在或已被删除 (用于删除评论，根据你的指定)
	ERROR_COMMENT_NOT_FOUND = 14

	// --- 详细业务错误码 (10000+) ---
	// 503: 服务器暂不可用 (发布新愿望)
	ERROR_SERVER_UNAVAILABLE = 10001
	// 500: 评论失败，请稍后再试
	ERROR_COMMENT_FAILED = 10002
	// 500: 获取互动数据失败
	ERROR_GET_INTERACTIONS_FAILED = 10003
	// 500: 操作失败，服务器出错了 (点赞/取消)
	ERROR_LIKE_FAILED = 10004

	// 403: 无权查看此愿望的互动
	ERROR_FORBIDDEN_INTERACTIONS = 13001
)

// MsgFlags是一个code，message的映射
var MsgFlags = map[int]string{
	SUCCESS: "成功",

	// --- 登录/权限 ---
	ERROR_LOGIN_FAILED:     "登录失败，用户名或密码错误", // 对应 code: 1 (使用 msg.go 的描述)
	ERROR_FORBIDDEN_DELETE: "无权删除此评论",       // 对应 code: 2
	ERROR_UNAUTHORIZED:     "登录已过期，请重新登录",   // 对应 code: 3 (使用最常见的 message)

	// --- 参数/服务器 ---
	ERROR_PARAM_INVALID: "参数验证失败", // 对应 code: 4
	ERROR_SERVER_ERROR:  "服务器错误",  // 对应 code: 10 (你的文档 中 message 不统一，使用 msg.go 默认)

	// --- 资源/业务逻辑 ---
	ERROR_WISH_DELETED:      "该心愿不存在或已被删除", // 对应 code: 11
	ERROR_WISH_NOT_FOUND:    "愿望已删除或不存在",   // 对应 code: 12 (统一使用 openapi.json 中的 message)
	ERROR_FORBIDDEN_COMMENT: "该愿望不允许评论",    // 对应 code: 13
	ERROR_COMMENT_NOT_FOUND: "评论不存在或已被删除",  // 对应 code: 14 (根据你的要求)

	// --- 详细业务错误码 ---
	ERROR_SERVER_UNAVAILABLE:      "服务器暂不可用",     // 对应 code: 10001
	ERROR_COMMENT_FAILED:          "评论失败，请稍后再试",  // 对应 code: 10002
	ERROR_GET_INTERACTIONS_FAILED: "获取互动数据失败",    // 对应 code: 10003
	ERROR_LIKE_FAILED:             "操作失败，服务器出错了", // 对应 code: 10004

	ERROR_FORBIDDEN_INTERACTIONS: "无权查看此愿望的互动", // 对应 code: 13001
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
