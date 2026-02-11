package main

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"

	relayer_node "github.com/multimarket-labs/event-pod-services"
	"github.com/multimarket-labs/event-pod-services/common/cliapp"
	"github.com/multimarket-labs/event-pod-services/common/opio"
	"github.com/multimarket-labs/event-pod-services/config"
	"github.com/multimarket-labs/event-pod-services/database"
	"github.com/multimarket-labs/event-pod-services/elasticsearch"
	"github.com/multimarket-labs/event-pod-services/services/api"
	"github.com/multimarket-labs/event-pod-services/services/common"
	grpc "github.com/multimarket-labs/event-pod-services/services/gprc"
)

var (
	ConfigFlag = &cli.StringFlag{
		Name:    "config",
		Value:   "./event-pool-services-config.local.yaml",
		Aliases: []string{"c"},
		Usage:   "path to config file",
		EnvVars: []string{"PHOENIX_SERVICES_CONFIG"},
	}
	MigrationsFlag = &cli.StringFlag{
		Name:    "migrations-dir",
		Value:   "./migrations",
		Usage:   "path to migrations folder",
		EnvVars: []string{"PHOENIX_SERVICES_MIGRATIONS_DIR"},
	}
)

func runMigrations(ctx *cli.Context) error {
	ctx.Context = opio.CancelOnInterrupt(ctx.Context)
	log.Info("running migrations...")
	cfg, err := config.New(ctx.String(ConfigFlag.Name))
	if err != nil {
		log.Error("failed to load config", "err", err)
		return err
	}
	db, err := database.NewDB(ctx.Context, cfg.MasterDB)
	if err != nil {
		log.Error("failed to connect to database", "err", err)
		return err
	}
	defer func(db *database.DB) {
		err := db.Close()
		if err != nil {
			log.Error("fail to close database", "err", err)
		}
	}(db)
	return db.ExecuteSQLMigration(cfg.Migrations)
}

func runInitDatabase(ctx *cli.Context) error {
	ctx.Context = opio.CancelOnInterrupt(ctx.Context)
	log.Info("初始化数据库基础数据...")
	cfg, err := config.New(ctx.String(ConfigFlag.Name))
	if err != nil {
		log.Error("failed to load config", "err", err)
		return err
	}
	db, err := database.NewDB(ctx.Context, cfg.MasterDB)
	if err != nil {
		log.Error("failed to connect to database", "err", err)
		return err
	}
	defer func(db *database.DB) {
		err := db.Close()
		if err != nil {
			log.Error("fail to close database", "err", err)
		}
	}(db)

	initService := common.NewInitService(db)
	if err := initService.InitDatabase(); err != nil {
		log.Error("初始化数据库失败", "err", err)
		return err
	}

	log.Info("数据库基础数据初始化完成")
	return nil
}

func runEventPool(ctx *cli.Context, shutdown context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	log.Info("running event pool node...")
	cfg, err := config.New(ctx.String(ConfigFlag.Name))
	if err != nil {
		log.Error("failed to load config", "err", err)
		return nil, err
	}
	return relayer_node.NewEventPool(ctx.Context, cfg, shutdown)
}

func runApi(ctx *cli.Context, _ context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	log.Info("running api...")
	cfg, err := config.New(ctx.String(ConfigFlag.Name))
	if err != nil {
		log.Error("failed to load config", "err", err)
		return nil, err
	}
	return api.NewApi(ctx.Context, cfg)
}

func runRpc(ctx *cli.Context, _ context.CancelCauseFunc) (cliapp.Lifecycle, error) {
	fmt.Println("running grpc services...")
	cfg, err := config.New(ctx.String(ConfigFlag.Name))
	if err != nil {
		log.Error("config error", "err", err)
		return nil, err
	}

	grpcServerCfg := &grpc.RpcConfig{
		Host: cfg.RpcServer.Host,
		Port: cfg.RpcServer.Port,
	}

	db, err := database.NewDB(ctx.Context, cfg.MasterDB)
	if err != nil {
		log.Error("new database fail", "err", err)
		return nil, err
	}

	// 初始化Elasticsearch客户端
	esClient, err := elasticsearch.NewESClient(ctx.Context, cfg.ElasticsearchConfig)
	if err != nil {
		log.Error("new elasticsearch client fail", "err", err)
		return nil, err
	}
	if esClient != nil {
		log.Info("elasticsearch client initialized successfully")
	}

	return grpc.NewRpcService(grpcServerCfg, db, esClient)
}

func NewCli() *cli.App {
	flags := []cli.Flag{ConfigFlag}
	migrationFlags := []cli.Flag{MigrationsFlag, ConfigFlag}
	return &cli.App{
		Version:              "0.0.1",
		Description:          "An Services For Phoenix Protocol",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:        "api",
				Flags:       flags,
				Description: "Run event http api service",
				Action:      cliapp.LifecycleCmd(runApi),
			},
			{
				Name:        "index",
				Flags:       flags,
				Description: "Run event node task",
				Action:      cliapp.LifecycleCmd(runEventPool),
			},
			{
				Name:        "rpc",
				Flags:       flags,
				Description: "Run event grpc service",
				Action:      cliapp.LifecycleCmd(runRpc),
			},
			{
				Name:        "migrate",
				Flags:       migrationFlags,
				Description: "Run event database migrations",
				Action:      runMigrations,
			},
			{
				Name:        "init-db",
				Flags:       flags,
				Description: "Initialize database with default data (languages, categories, ecosystems)",
				Action:      runInitDatabase,
			},
			{
				Name:        "version",
				Description: "Show event services project version",
				Action: func(ctx *cli.Context) error {
					cli.ShowVersion(ctx)
					return nil
				},
			},
		},
	}
}
