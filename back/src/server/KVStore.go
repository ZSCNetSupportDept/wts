package server

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"zsxyww.com/wts/model/sqlc"
)

func KVGet(key string) (string, error) {
	ctx := context.Background()
	var value string
	err := DB.DoQuery(ctx, "system", func(q *sqlc.Queries) error {
		v, e := q.KVGet(ctx, key)
		if e != nil {
			if errors.Is(e, pgx.ErrNoRows) {
				return nil // 未配置视为空，不视为错误
			}
			return e
		}
		value = v
		return nil
	})
	return value, err
}

func KVPut(key string, value string) error {
	ctx := context.Background()
	err := DB.DoQuery(ctx, "system", func(q *sqlc.Queries) error {
		return q.KVPut(ctx, sqlc.KVPutParams{
			Key:   key,
			Value: value,
		})
	})
	return err
}

func KVDelete(key string) error {
	ctx := context.Background()
	err := DB.DoQuery(ctx, "system", func(q *sqlc.Queries) error {
		return q.KVDelete(ctx, key)
	})
	return err
}
