package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ethereum/go-ethereum/log"

	"github.com/multimarket-labs/event-pod-services/services/api/models"
)

// CreateEventHandler 处理 POST /api/v1/events
// 接口 A：生成预测事件
func (rs *Routes) CreateEventHandler(w http.ResponseWriter, r *http.Request) {
	// 记录请求开始
	log.Info("=== CreateEvent Request Started ===",
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.Header.Get("User-Agent"),
	)

	// 解析请求体
	var req models.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error("failed to decode request body", "err", err)
		jsonResponse(w, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "Failed to parse request body: " + err.Error(),
		}, http.StatusBadRequest)
		return
	}

	// 将请求体转换为 JSON 字符串用于日志
	reqJSON, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		log.Warn("failed to marshal request to JSON", "err", err)
		reqJSON = []byte("{}")
	}

	// 打印完整的 curl 命令（可以直接复制使用）
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = "localhost:8080"
	}
	curlCmd := fmt.Sprintf("curl -X POST '%s://%s%s' -H 'Content-Type: application/json' -d '%s'",
		scheme, host, r.URL.Path, string(reqJSON))

	log.Info("📋 Curl command to reproduce this request:")
	log.Info(curlCmd)

	// 打印请求详情摘要
	log.Info("CreateEvent request summary",
		"category_guid", req.CategoryGUID,
		"ecosystem_guid", req.EcosystemGUID,
		"language_guid", req.LanguageGUID,
		"title", req.Title,
		"is_sports", req.IsSports,
		"sub_events_count", len(req.SubEvents),
	)

	// 调用 service 层
	log.Info("Calling service layer to create event")
	response, err := rs.svc.CreateEvent(&req)
	if err != nil {
		log.Error("failed to create event", "err", err)
		jsonResponse(w, models.ErrorResponse{
			Error:   "create_failed",
			Message: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	// 记录成功响应
	log.Info("CreateEvent succeeded",
		"event_guid", response.GUID,
		"title", response.Title,
		"sub_events_count", len(response.SubEvents),
	)

	// 返回成功响应
	jsonResponse(w, response, http.StatusCreated)
	log.Info("=== CreateEvent Request Completed ===")
}

// ListEventsHandler 处理 GET /api/v1/events
// 接口 B：查询事件列表
func (rs *Routes) ListEventsHandler(w http.ResponseWriter, r *http.Request) {
	// 记录请求开始
	log.Info("=== ListEvents Request Started ===",
		"method", r.Method,
		"path", r.URL.Path,
		"query", r.URL.RawQuery,
		"remote_addr", r.RemoteAddr,
	)

	// 解析查询参数
	req := models.ListEventsRequest{
		Page:  1,
		Limit: 20,
	}

	// 解析 page
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			req.Page = page
		}
	}

	// 解析 limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	// 解析 language_guid（必需）
	req.LanguageGUID = r.URL.Query().Get("language_guid")
	if req.LanguageGUID == "" {
		// 尝试从 Header 中获取 Accept-Language
		acceptLang := r.Header.Get("Accept-Language")
		if acceptLang != "" {
			// 这里可以根据 Accept-Language 映射到 language_guid
			// 简化处理：直接使用 Accept-Language 作为 language_guid
			req.LanguageGUID = acceptLang
		}
	}

	if req.LanguageGUID == "" {
		log.Error("language_guid is required")
		jsonResponse(w, models.ErrorResponse{
			Error:   "invalid_request",
			Message: "language_guid is required (via query parameter or Accept-Language header)",
		}, http.StatusBadRequest)
		return
	}

	// 解析 category_guid（可选）
	req.CategoryGUID = r.URL.Query().Get("category_guid")

	// 解析 is_live（可选）
	if isLiveStr := r.URL.Query().Get("is_live"); isLiveStr != "" {
		if isLive, err := strconv.ParseInt(isLiveStr, 10, 16); err == nil {
			isLiveInt16 := int16(isLive)
			req.IsLive = &isLiveInt16
		}
	}

	// 打印请求参数
	log.Info("ListEvents request parameters",
		"page", req.Page,
		"limit", req.Limit,
		"language_guid", req.LanguageGUID,
		"category_guid", req.CategoryGUID,
		"is_live", req.IsLive,
	)

	// 构建完整的 curl 命令（可以直接复制使用）
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	if host == "" {
		host = "localhost:8080"
	}
	fullURL := fmt.Sprintf("%s://%s%s", scheme, host, r.URL.Path)
	if r.URL.RawQuery != "" {
		fullURL += "?" + r.URL.RawQuery
	}
	curlCmd := fmt.Sprintf("curl -X GET '%s'", fullURL)

	log.Info("📋 Curl command to reproduce this request:")
	log.Info(curlCmd)

	// 调用 service 层
	log.Info("Calling service layer to list events")
	response, err := rs.svc.ListEvents(&req)
	if err != nil {
		log.Error("failed to list events", "err", err)
		jsonResponse(w, models.ErrorResponse{
			Error:   "list_failed",
			Message: err.Error(),
		}, http.StatusInternalServerError)
		return
	}

	// 记录成功响应
	log.Info("ListEvents succeeded",
		"events_count", len(response.Events),
		"total", response.Pagination.Total,
		"page", response.Pagination.Page,
		"total_pages", response.Pagination.TotalPages,
	)

	// 返回成功响应
	jsonResponse(w, response, http.StatusOK)
	log.Info("=== ListEvents Request Completed ===")
}
