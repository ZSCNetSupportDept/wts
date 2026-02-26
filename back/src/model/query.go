package model

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"zsxyww.com/wts/model/sqlc"
)

// 系统的一切数据库操作用这个包装，从这个函数返回任何error都会回滚事务，参数是sqlc生成的handler，在函数里面调用它的方法即可进行数据库单元查询
type Query func(q *sqlc.Queries) error

type Store struct {
	*sqlc.Queries
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		Queries: sqlc.New(db),
		db:      db,
	}
}

// 为了方便自动设置RLS上下文，设置了这个函数
func (store *Store) DoQuery(ctx context.Context, wx string, fn Query) error {
	tx, err := store.db.Begin(ctx)
	if err != nil {
		return err
	}
	// Go语言的defer会在返回值表达式被计算之后执行，所以成功时不会回滚
	defer tx.Rollback(ctx)

	qtx := store.WithTx(tx)

	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL wts.wx = '%s'", wx))
	if err != nil {
		return fmt.Errorf("failed to set local wx context: %w", err)
	}

	err = fn(qtx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
