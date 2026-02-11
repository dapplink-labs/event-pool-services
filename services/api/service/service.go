package service

import (
	"context"

	"github.com/multimarket-labs/event-pod-services/database"
	"github.com/multimarket-labs/event-pod-services/services/api/models"
	"github.com/multimarket-labs/event-pod-services/services/api/validator"
	"github.com/multimarket-labs/event-pod-services/services/common"
)

// Service interface defines the business logic methods
// Add your custom service methods here
type Service interface {
	// Example: Add your business methods here
	// GetUserInfo(userID string) (*UserInfo, error)

	// GetPredictEvent calls Dify workflow API to convert natural language query into structured event data
	GetPredictEvent(ctx context.Context, userQuery string) (*EventDetail, error)

	// CreateEvent 创建新的预测事件
	CreateEvent(req *models.CreateEventRequest) (*models.CreateEventResponse, error)

	// ListEvents 查询事件列表（支持多语言）
	ListEvents(req *models.ListEventsRequest) (*models.ListEventsResponse, error)
}

type HandlerSvc struct {
	v                    *validator.Validator
	db                   *database.DB
	authenticatorService *common.AuthenticatorService
	verificationManager  *common.VerificationCodeManager
}

func New(v *validator.Validator,
	db *database.DB,
	authenticatorService *common.AuthenticatorService,
) Service {
	return &HandlerSvc{
		v:                    v,
		db:                   db,
		authenticatorService: authenticatorService,
		verificationManager:  common.NewVerificationCodeManager(),
	}
}
