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

	// CreateEvent creates a new prediction event with sub-events, outcomes, and tags
	CreateEvent(req *models.CreateEventRequest) (*models.CreateEventResponse, error)

	// ListEvents retrieves events with filtering and pagination
	ListEvents(req *models.ListEventsRequest) (*models.ListEventsResponse, error)

	// GetEventDetail retrieves full event details including sub-events and outcomes
	GetEventDetail(guid string) (*models.GetEventDetailResponse, error)

	// GetPredictEvent calls Dify workflow API to convert natural language query into structured event data
	GetPredictEvent(ctx context.Context, userQuery string) (*EventDetail, error)
}

type HandlerSvc struct {
	v                    *validator.Validator
	db                   *database.DB
	emailService         *common.EmailService
	smsService           *common.SMSService
	authenticatorService *common.AuthenticatorService
	verificationManager  *common.VerificationCodeManager
	siweVerifier         *common.SIWEVerifier
	kodoService          *common.KodoService
	s3Service            *common.S3Service
	minioService         *common.StorageService
	jwtSecret            string
}

func New(v *validator.Validator,
	db *database.DB,
	emailService *common.EmailService,
	smsService *common.SMSService,
	authenticatorService *common.AuthenticatorService,
	kodoService *common.KodoService,
	s3Service *common.S3Service,
	minioService *common.StorageService,
	jwtSecret string,
	domain string,
) Service {
	return &HandlerSvc{
		v:                    v,
		db:                   db,
		emailService:         emailService,
		smsService:           smsService,
		authenticatorService: authenticatorService,
		verificationManager:  common.NewVerificationCodeManager(),
		kodoService:          kodoService,
		s3Service:            s3Service,
		minioService:         minioService,
		siweVerifier:         common.NewSIWEVerifier(jwtSecret, domain),
		jwtSecret:            jwtSecret,
	}
}
