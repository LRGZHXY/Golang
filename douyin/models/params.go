package models

// ParamSignUp 注册请求参数
type ParamSignUp struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

// ParamLogin 登录请求参数
type ParamLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
