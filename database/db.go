package database

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"path/filepath"

	"github.com/multimarket-labs/event-pod-services/common/retry"
	"github.com/multimarket-labs/event-pod-services/config"
)

type DB struct {
	gorm        *gorm.DB
	Languages   LanguagesDB
	Category    CategoryDB
	Ecosystem   EcosystemDB
	EventPeriod EventPeriodDB
	TeamGroup   TeamGroupDB
	Event       EventDB
	SubEvent    SubEventDB
}

func NewDB(ctx context.Context, dbConfig config.DBConfig) (*DB, error) {
	dsn := fmt.Sprintf("host=%s dbname=%s sslmode=disable", dbConfig.Host, dbConfig.Name)
	if dbConfig.Port != 0 {
		dsn += fmt.Sprintf(" port=%d", dbConfig.Port)
	}
	if dbConfig.User != "" {
		dsn += fmt.Sprintf(" user=%s", dbConfig.User)
	}
	if dbConfig.Password != "" {
		dsn += fmt.Sprintf(" password=%s", dbConfig.Password)
	}

	gormConfig := gorm.Config{
		SkipDefaultTransaction: true,
		CreateBatchSize:        3_000,
	}
	retryStrategy := &retry.ExponentialStrategy{Min: 1000, Max: 20_000, MaxJitter: 250}
	gorms, err := retry.Do[*gorm.DB](context.Background(), 10, retryStrategy, func() (*gorm.DB, error) {
		gorms, err := gorm.Open(postgres.Open(dsn), &gormConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}
		return gorms, nil
	})

	if err != nil {
		return nil, err
	}

	db := &DB{
		gorm:        gorms,
		Languages:   NewLanguagesDB(gorms),
		Category:    NewCategoryDB(gorms),
		Ecosystem:   NewEcosystemDB(gorms),
		EventPeriod: NewEventPeriodDB(gorms),
		TeamGroup:   NewTeamGroupDB(gorms),
		Event:       NewEventDB(gorms),
		SubEvent:    NewSubEventDB(gorms),
	}
	return db, nil
}

func (db *DB) Transaction(fn func(db *DB) error) error {
	return db.gorm.Transaction(func(tx *gorm.DB) error {
		txDB := &DB{
			gorm:        tx,
			Languages:   NewLanguagesDB(tx),
			Category:    NewCategoryDB(tx),
			Ecosystem:   NewEcosystemDB(tx),
			EventPeriod: NewEventPeriodDB(tx),
			TeamGroup:   NewTeamGroupDB(tx),
			Event:       NewEventDB(tx),
			SubEvent:    NewSubEventDB(tx),
		}
		return fn(txDB)
	})
}

func (db *DB) Exec(sql string, values ...interface{}) *gorm.DB {
	return db.gorm.Exec(sql, values...)
}

func (db *DB) Close() error {
	sql, err := db.gorm.DB()
	if err != nil {
		return err
	}
	return sql.Close()
}

func (db *DB) ExecuteSQLMigration(migrationsFolder string) error {
	err := filepath.Walk(migrationsFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("Failed to process migration file: %s", path))
		}
		if info.IsDir() {
			return nil
		}
		fileContent, readErr := os.ReadFile(path)
		if readErr != nil {
			return errors.Wrap(readErr, fmt.Sprintf("Error reading SQL file: %s", path))
		}

		execErr := db.gorm.Exec(string(fileContent)).Error
		if execErr != nil {
			return errors.Wrap(execErr, fmt.Sprintf("Error executing SQL script: %s", path))
		}
		return nil
	})
	return err
}

// GetGorm returns the underlying gorm.DB instance
// Use this when you need direct access to GORM for custom queries
func (db *DB) GetGorm() *gorm.DB {
	return db.gorm
}
