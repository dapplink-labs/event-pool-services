package models

// SendVerificationCodeRequest 发送验证码请求（60秒有效期）
type SendVerificationCodeRequest struct {
	Target string `json:"target" binding:"required"` // 目标: 邮箱地址或手机号
	Type   string `json:"type" binding:"required"`   // 验证码类型: email, sms
}

// SendVerificationCodeResponse 发送验证码响应
type SendVerificationCodeResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	Code          string `json:"code,omitempty"` // 验证码（测试环境可返回）
	ExpiresIn     int    `json:"expires_in"`     // 过期时间（秒）
	RemainingTime int    `json:"remaining_time"` // 剩余时间（秒）
}

// VerifyVerificationCodeRequest 验证验证码请求
type VerifyVerificationCodeRequest struct {
	Target string `json:"target" binding:"required"` // 目标: 邮箱地址或手机号
	Code   string `json:"code" binding:"required"`   // 验证码
}

// VerifyVerificationCodeResponse 验证验证码响应
type VerifyVerificationCodeResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type GenerateTOTPRequest struct {
	UserId string `json:"user_id" binding:"required"`
}

type GenerateTOTPResponse struct {
	Secret        string `json:"secret"`
	QRCodeURL     string `json:"qr_code_url"`
	QRCodeImage   string `json:"qr_code_image"`
	CurrentCode   string `json:"current_code"`
	RemainingTime int    `json:"remaining_time"`
	Message       string `json:"message"`
}

type VerifyTOTPRequest struct {
	Secret string `json:"secret" binding:"required"`
	Code   string `json:"code" binding:"required"`
}

type VerifyTOTPResponse struct {
	Valid         bool   `json:"valid"`
	RemainingTime int    `json:"remaining_time"`
	Message       string `json:"message"`
}
