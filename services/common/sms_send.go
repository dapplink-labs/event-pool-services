package common

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/multimarket-labs/event-pod-services/config"
)

type SMSMessage struct {
	PhoneNumbers  []string
	TemplateCode  string
	TemplateParam map[string]string
	SignName      string
}

type SMSService struct {
	config *config.SMSConfig
}

type SMSResponse struct {
	RequestId string `json:"RequestId"`
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	BizId     string `json:"BizId"`
}

func NewSMSService(cfg *config.SMSConfig) (*SMSService, error) {
	if err := validateSMSConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid sms config: %w", err)
	}
	return &SMSService{config: cfg}, nil
}

func (s *SMSService) SendSMS(msg *SMSMessage) error {
	if err := validateSMSMessage(msg); err != nil {
		return fmt.Errorf("invalid sms message: %w", err)
	}

	if msg.SignName == "" {
		msg.SignName = s.config.SignName
	}
	if msg.TemplateCode == "" {
		msg.TemplateCode = s.config.TemplateCode
	}

	params := s.buildParams(msg)

	signature := s.sign(params, "POST")
	params["Signature"] = signature

	return s.sendRequest(params)
}

func (s *SMSService) buildParams(msg *SMSMessage) map[string]string {
	params := make(map[string]string)

	params["SignatureMethod"] = "HMAC-SHA1"
	params["SignatureNonce"] = uuid.New().String()
	params["AccessKeyId"] = s.config.AccessKeyId
	params["SignatureVersion"] = "1.0"
	params["Timestamp"] = time.Now().UTC().Format("2006-01-02T15:04:05Z")
	params["Format"] = "JSON"

	params["Action"] = "SendSms"
	params["Version"] = "2017-05-25"
	params["RegionId"] = "cn-hangzhou"
	params["PhoneNumbers"] = strings.Join(msg.PhoneNumbers, ",")
	params["SignName"] = msg.SignName
	params["TemplateCode"] = msg.TemplateCode

	if len(msg.TemplateParam) > 0 {
		paramBytes, _ := json.Marshal(msg.TemplateParam)
		params["TemplateParam"] = string(paramBytes)
	}

	return params
}

func (s *SMSService) sign(params map[string]string, method string) string {
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sortedParams []string
	for _, k := range keys {
		sortedParams = append(sortedParams, url.QueryEscape(k)+"="+url.QueryEscape(params[k]))
	}
	canonicalizedQueryString := strings.Join(sortedParams, "&")

	stringToSign := method + "&" + url.QueryEscape("/") + "&" + url.QueryEscape(canonicalizedQueryString)

	h := hmac.New(sha1.New, []byte(s.config.AccessKeySecret+"&"))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return signature
}

func (s *SMSService) sendRequest(params map[string]string) error {
	endpoint := s.config.Endpoint
	if endpoint == "" {
		endpoint = "dysmsapi.aliyuncs.com"
	}

	apiUrl := fmt.Sprintf("https://%s/", endpoint)

	data := url.Values{}
	for k, v := range params {
		data.Set(k, v)
	}

	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var smsResp SMSResponse
	if err := json.Unmarshal(body, &smsResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if smsResp.Code != "OK" {
		return fmt.Errorf("sms send failed: code=%s, message=%s", smsResp.Code, smsResp.Message)
	}

	return nil
}

func (s *SMSService) SendVerificationCode(phoneNumber, code string) error {
	msg := &SMSMessage{
		PhoneNumbers: []string{phoneNumber},
		TemplateParam: map[string]string{
			"code": code,
		},
	}
	return s.SendSMS(msg)
}

func validateSMSConfig(cfg *config.SMSConfig) error {
	if cfg.AccessKeyId == "" {
		return fmt.Errorf("access key id is required")
	}
	if cfg.AccessKeySecret == "" {
		return fmt.Errorf("access key secret is required")
	}
	if cfg.SignName == "" {
		return fmt.Errorf("sign name is required")
	}
	if cfg.TemplateCode == "" {
		return fmt.Errorf("template code is required")
	}
	return nil
}

func validateSMSMessage(msg *SMSMessage) error {
	if len(msg.PhoneNumbers) == 0 {
		return fmt.Errorf("at least one phone number is required")
	}
	for _, phone := range msg.PhoneNumbers {
		if phone == "" {
			return fmt.Errorf("phone number cannot be empty")
		}
	}
	return nil
}
