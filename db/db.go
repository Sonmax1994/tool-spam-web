package db

import (
	"context"
	"database/sql"
	"dd/config"
	"fmt"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
	//"github.com/uptrace/bun/extra/bundebug"
)

func buildDSN(cfg config.Database) string {
	params := fmt.Sprintf("timeout=%s&readTimeout=%s&writeTimeout=%s",
		time.Duration(30)*time.Second,
		time.Duration(30)*time.Second,
		time.Duration(30)*time.Second,
	)
	params += "&tls=skip-verify"
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.DB, params)
}

func MakeMysqlConnect(cfg config.Database) (*bun.DB, error) {
	dsn := buildDSN(cfg)
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := bun.NewDB(sqlDB, mysqldialect.New())
	//db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
