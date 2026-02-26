package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/mattn/go-sqlite3"
	"zsxyww.com/wts/config"
)

func Connect(cfg *config.Config) *pgxpool.Pool {

	if cfg.DB.Type == "PostgreSQL" {

		url := "postgres" + "://" + cfg.DB.User + ":" + cfg.DB.Password + "@" + cfg.DB.Path + ":" + cfg.DB.Port + "/" + cfg.DB.Name
		if !cfg.DB.SSL {
			url += "?sslmode=disable"
		}
		if cfg.Debug.ProgramVerbose {
			println(url)
		}

		dbcfg, err := pgxpool.ParseConfig(url)
		if err != nil {
			panic(err)
		}
		dbcfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			// pgx无法自动识别数据库里自定义的枚举类型及其数组类型，在这里手动注册
			for _, baseTypeName := range []string{"wts.block", "wts.isp", "wts.category", "wts.status"} {

				baseType, err := conn.LoadType(ctx, baseTypeName)
				if err != nil {
					return fmt.Errorf("failed to load base type %s: %w", baseTypeName, err)
				}
				conn.TypeMap().RegisterType(baseType)

				arrayTypeName := baseTypeName + "[]"
				arrayType, err := conn.LoadType(ctx, arrayTypeName)
				if err != nil {
					return fmt.Errorf("failed to load array type %s: %w", arrayTypeName, err)
				}
				conn.TypeMap().RegisterType(arrayType)
			}
			return nil
		}

		db, err := pgxpool.NewWithConfig(context.Background(), dbcfg)
		if err != nil {
			panic(err)
		}

		ct, err := db.Exec(context.Background(), "SELECT version();")
		if err != nil {
			panic(err)
		}
		if cfg.Debug.ProgramVerbose {
			fmt.Println(ct.String())
		}

		return db

	}

	//if cfg.DB.Type == "SQLite" {
	//	db := sqlx.MustConnect("sqlite3", cfg.DB.Path)
	//	return db
	//}

	panic("Unsupported database type: " + cfg.DB.Type)

}
