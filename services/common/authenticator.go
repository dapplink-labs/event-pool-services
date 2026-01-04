package common

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type AuthenticatorService struct {
	issuer string
}

type TOTPKey struct {
	Secret      string `json:"secret"`        // 密钥
	QRCodeURL   string `json:"qr_code_url"`   // 二维码URL
	QRCodeImage string `json:"qr_code_image"` // Base64编码的二维码图片
}

func NewAuthenticatorService(issuer string) *AuthenticatorService {
	if issuer == "" {
		issuer = "PHOENIX"
	}
	return &AuthenticatorService{
		issuer: issuer,
	}
}

func (s *AuthenticatorService) GenerateSecret(userUuid string) (*TOTPKey, error) {
	if userUuid == "" {
		return nil, fmt.Errorf("account name is required")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: userUuid,
		Period:      30,
		Digits:      otp.DigitsSix,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	img, err := key.Image(256, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code image: %w", err)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode QR code: %w", err)
	}

	base64Image := base64.StdEncoding.EncodeToString(buf.Bytes())

	return &TOTPKey{
		Secret:      key.Secret(),
		QRCodeURL:   key.URL(),
		QRCodeImage: base64Image,
	}, nil
}

func (s *AuthenticatorService) VerifyCode(secret, code string) (bool, error) {
	if secret == "" {
		return false, fmt.Errorf("secret is required")
	}
	if code == "" {
		return false, fmt.Errorf("code is required")
	}

	code = strings.TrimSpace(code)

	if len(code) != 6 {
		return false, fmt.Errorf("code must be 6 digits, got %d digits", len(code))
	}

	for _, c := range code {
		if c < '0' || c > '9' {
			return false, fmt.Errorf("code must contain only digits")
		}
	}

	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period: 30,
		Skew:   1,
		Digits: otp.DigitsSix,
	})

	if err != nil {
		return false, fmt.Errorf("failed to validate code: %w", err)
	}

	return valid, nil
}

func (s *AuthenticatorService) ValidateWithWindow(secret, code string) (bool, error) {
	if secret == "" {
		return false, fmt.Errorf("secret is required")
	}
	if code == "" {
		return false, fmt.Errorf("code is required")
	}

	if len(code) != 6 {
		return false, fmt.Errorf("code must be 6 digits")
	}

	valid, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{
		Period: 30,
		Skew:   1,
		Digits: otp.DigitsSix,
	})

	if err != nil {
		return false, fmt.Errorf("failed to validate code: %w", err)
	}

	return valid, nil
}

func (s *AuthenticatorService) GenerateCurrentCode(secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("secret is required")
	}

	code, err := totp.GenerateCodeCustom(secret, time.Now(), totp.ValidateOpts{
		Period: 30,
		Digits: otp.DigitsSix,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}

	return code, nil
}

func (s *AuthenticatorService) GetRemainingTime() int {
	now := time.Now().Unix()
	period := int64(30)
	remaining := period - (now % period)
	return int(remaining)
}
