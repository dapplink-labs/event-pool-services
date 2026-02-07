package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/multimarket-labs/event-pod-services/services/api/validator"

	"github.com/ethereum/go-ethereum/log"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/multimarket-labs/event-pod-services/common/httputil"
	"github.com/multimarket-labs/event-pod-services/config"
	"github.com/multimarket-labs/event-pod-services/database"
	"github.com/multimarket-labs/event-pod-services/services/api/routes"
	"github.com/multimarket-labs/event-pod-services/services/api/service"
	common2 "github.com/multimarket-labs/event-pod-services/services/common"
)

const (
	HealthPath = "/healthz"

	AdminLoginV1Path  = "/api/v1/admin/login"
	AdminLogoutV1Path = "/api/v1/admin/logout"

	AdminAuthListV1Path   = "/api/v1/admin/authList"
	AdminAuthCreateV1Path = "/api/v1/admin/auth/create"
	AdminAuthUpdateV1Path = "/api/v1/admin/auth/update"
	AdminAuthDeleteV1Path = "/api/v1/admin/auth/delete"
)

type APIConfig struct {
	HTTPServer    config.ServerConfig
	MetricsServer config.ServerConfig
}

type API struct {
	router    *chi.Mux
	apiServer *httputil.HTTPServer
	db        *database.DB
	stopped   atomic.Bool
}

func NewApi(ctx context.Context, cfg *config.Config) (*API, error) {
	out := &API{}
	if err := out.initFromConfig(ctx, cfg); err != nil {
		return nil, errors.Join(err, out.Stop(ctx))
	}
	return out, nil
}

func (a *API) initFromConfig(ctx context.Context, cfg *config.Config) error {
	if err := a.initDB(ctx, cfg); err != nil {
		return fmt.Errorf("failed to init DB: %w", err)
	}
	a.initRouter(ctx, cfg)
	if err := a.startServer(cfg.HttpServer); err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}
	return nil
}

func (a *API) initRouter(ctx context.Context, cfg *config.Config) {
	allowedOrigins := []string{"http://localhost:8080", "http://127.0.0.1:8080"}
	allowAllOrigins := false
	if cfg.CORSAllowedOrigins != "" {
		if cfg.CORSAllowedOrigins == "*" {
			allowAllOrigins = true
		} else {
			allowedOrigins = parseCORSOrigins(cfg.CORSAllowedOrigins)
		}
	}

	v := new(validator.Validator)

	emailService, err := common2.NewEmailService(&cfg.EmailConfig)
	if err != nil {
		log.Error("failed to create email service", "err", err)
	}

	authenticatorService := common2.NewAuthenticatorService("PHOENIX")

	var kodoService *common2.KodoService
	if cfg.KodoConfig.AccessKey != "" && cfg.KodoConfig.SecretKey != "" {
		kodoService, err = common2.NewKodoService(&cfg.KodoConfig)
		if err != nil {
			log.Error("failed to create kodo service", "err", err)
		} else {
			log.Info("kodo service initialized successfully")
		}
	}

	var s3Service *common2.S3Service
	if cfg.S3Config.AccessKey != "" && cfg.S3Config.SecretKey != "" {
		s3Service, err = common2.NewS3Service(&cfg.S3Config)
		if err != nil {
			log.Error("failed to create s3 service", "err", err)
		} else {
			log.Info("s3 service initialized successfully")
		}
	}

	var minioService *common2.StorageService
	if cfg.MinioConfig.Endpoint != "" {
		minioService = common2.NewStorageService(cfg.MinioConfig)
		log.Info("minio service initialized successfully")
	}

	svc := service.New(v, a.db, emailService, authenticatorService, kodoService, s3Service, minioService, cfg.JWTSecret, cfg.Domain)
	apiRouter := chi.NewRouter()

	// Add all middlewares BEFORE registering routes
	apiRouter.Use(middleware.Timeout(time.Second * 200)) // Increased for AI prediction (Dify API can take 30-60s)
	apiRouter.Use(middleware.Recoverer)

	// Add CORS middleware
	corsOptions := cors.Options{
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}

	if allowAllOrigins {
		corsOptions.AllowOriginFunc = func(r *http.Request, origin string) bool {
			return true
		}
	} else {
		corsOptions.AllowedOrigins = allowedOrigins
	}

	apiRouter.Use(cors.Handler(corsOptions))

	apiRouter.Use(middleware.Heartbeat(HealthPath))

	// Register routes AFTER all middlewares are defined
	_ = routes.NewRoutes(apiRouter, svc)

	apiRouter.NotFound(func(w http.ResponseWriter, r *http.Request) {
		log.Warn("NotFoundHandler hit", "path", r.URL.Path, "method", r.Method)
		http.Error(w, "route not found", http.StatusNotFound)
	})

	apiRouter.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		log.Warn("MethodNotAllowedHandler hit", "path", r.URL.Path, "method", r.Method)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	/*
	 * ============== backend ===============
	 */
	// TODO: Add your API routes here
	// Example:
	// apiRouter.Get("/api/v1/users", h.GetUsersHandler)
	// apiRouter.Post("/api/v1/users", h.CreateUserHandler)

	/*
	 * Admin routes have been removed. Add your custom routes above.
	 */
	// apiRouter.Post(fmt.Sprintf(AdminLoginV1Path), h.AdminLoginHandler)
	// apiRouter.Post(fmt.Sprintf(AdminLogoutV1Path), h.AdminLogoutHandler)
	// apiRouter.Post(fmt.Sprintf(AdminAuthListV1Path), h.AdminAuthListHandler)
	// apiRouter.Post(fmt.Sprintf(AdminAuthCreateV1Path), h.AdminAuthCreateHandler)
	// apiRouter.Post(fmt.Sprintf(AdminAuthUpdateV1Path), h.AdminAuthUpdateHandler)
	// apiRouter.Post(fmt.Sprintf(AdminAuthDeleteV1Path), h.AdminAuthDeleteHandler)
	//
	// // 角色管理路由
	// apiRouter.Post("/api/v1/role/list", h.GetRoleListHandler)
	// apiRouter.Post("/api/v1/role/create", h.CreateRoleHandler)
	// apiRouter.Post("/api/v1/role/update", h.UpdateRoleHandler)
	// apiRouter.Post("/api/v1/role/delete", h.DeleteRoleHandler)
	// apiRouter.Post("/api/v1/role/info", h.GetRoleInfoHandler)

	/*
	 * ============== frontend ===============
	 */
	a.router = apiRouter
}

func (a *API) initDB(ctx context.Context, cfg *config.Config) error {
	var initDb *database.DB
	var err error
	if !cfg.SlaveDbEnable {
		initDb, err = database.NewDB(ctx, cfg.MasterDB)
		if err != nil {
			log.Error("failed to connect to master database", "err", err)
			return err
		}
	} else {
		initDb, err = database.NewDB(ctx, cfg.SlaveDB)
		if err != nil {
			log.Error("failed to connect to slave database", "err", err)
			return err
		}
	}
	a.db = initDb
	return nil
}

func (a *API) Start(ctx context.Context) error {
	return nil
}

func (a *API) Stop(ctx context.Context) error {
	var result error
	if a.apiServer != nil {
		if err := a.apiServer.Stop(ctx); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to stop API server: %w", err))
		}
	}
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			result = errors.Join(result, fmt.Errorf("failed to close DB: %w", err))
		}
	}
	a.stopped.Store(true)
	log.Info("API service shutdown complete")
	return result
}

func (a *API) startServer(serverConfig config.ServerConfig) error {
	log.Debug("API server listening...", "port", serverConfig.Port)
	addr := net.JoinHostPort(serverConfig.Host, strconv.Itoa(serverConfig.Port))
	srv, err := httputil.StartHTTPServer(addr, a.router)
	if err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}
	log.Info("API server started", "addr", srv.Addr().String())
	a.apiServer = srv
	return nil
}

func (a *API) Stopped() bool {
	return a.stopped.Load()
}

// parseCORSOrigins parses comma-separated CORS origins string
func parseCORSOrigins(origins string) []string {
	var result []string
	for _, origin := range strings.Split(origins, ",") {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
