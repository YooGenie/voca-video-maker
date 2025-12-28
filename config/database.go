package config

import (
	"context"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"xorm.io/core"
)

var (
	xormDb     *xorm.Engine
	dbOnce     sync.Once
	dbInstance DatabaseWrapper
)

// ConfigureDatabase - 데이터베이스 연결을 초기화합니다 (한 번만 실행됨)
func ConfigureDatabase() DatabaseWrapper {
	dbOnce.Do(func() {
		dbConnection := Config.Database.ConnectionString

		engine, err := xorm.NewEngine(Config.Database.Driver, dbConnection)
		if err != nil {
			panic(fmt.Errorf("데이터베이스 연결 오류: %w", err))
		}
		fmt.Println("DB connected: ", Config.Database.Connection)

		engine.SetMaxOpenConns(10)
		engine.SetMaxIdleConns(5)
		engine.SetConnMaxLifetime(10 * time.Minute)
		engine.Logger().SetLevel(core.LOG_INFO)

		xormDb = engine
		dbInstance = DatabaseWrapper{engine}
	})

	return dbInstance
}

// GetDatabase - 이미 초기화된 데이터베이스 연결을 반환합니다
func GetDatabase() DatabaseWrapper {
	if xormDb == nil {
		panic("Database not initialized. Call ConfigureDatabase() first.")
	}
	return dbInstance
}

type DatabaseWrapper struct {
	*xorm.Engine
}

func (d DatabaseWrapper) CreateSession(ctx context.Context) (*xorm.Session, context.Context) {
	session := d.NewSession()

	func(session interface{}, ctx context.Context) {
		if s, ok := session.(interface{ SetContext(context.Context) }); ok {
			s.SetContext(ctx)
		}
	}(session, ctx)
	defer session.Close()

	return session, context.WithValue(ctx, ContextDBKey, session)
}

func CleanUp() {
	if xormDb != nil {
		xormDb.Close()
	}
}
