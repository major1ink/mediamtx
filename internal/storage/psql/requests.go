package psql

import (
	"context"
	"time"
)

func (r *Req) ExecQuery(query string) error {
	ctx, cancel := context.WithTimeout(r.ctx, time.Duration(r.queryTimeOut)* time.Second)
    defer cancel()
	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (r *Req) ExecQueryNoCtx(query string) error {
	r.ctx = context.Background()
		ctx, cancel := context.WithTimeout(r.ctx, time.Duration(r.queryTimeOut)* time.Second)
    defer cancel()
	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return err
	}
	r.ctx.Done()
	return nil
}

func (r *Req) SelectData(query string) ([][]interface{}, error) {

	if r.ctx.Err() != nil {
		return nil, r.ctx.Err()
	}
	ctx, cancel := context.WithTimeout(r.ctx, time.Duration(r.queryTimeOut)* time.Second)
    defer cancel()
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result [][]interface{}

	for rows.Next() {
		line := make([]interface{}, len(rows.FieldDescriptions()))
		for i := range line {
			var val interface{}
			line[i] = &val
		}

		if err := rows.Scan(line...); err != nil {
			return nil, err
		}

		for i, col := range line {
			line[i] = *col.(*interface{})
		}

		result = append(result, line)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return result, nil
}

func (r *Req) SelectPathStream(query string) (int8, string, error) {
	if r.ctx.Err() != nil {
		return 0, "", r.ctx.Err()
	}
	ctx, cancel := context.WithTimeout(r.ctx, time.Duration(r.queryTimeOut)* time.Second)
    defer cancel()
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return 0, "", err
	}
	defer rows.Close()

	var result string
	var ststus int8
	for rows.Next() {
		err := rows.Scan(&ststus, &result)
		if err != nil {
			return 0, "", err
		}
	}

	if rows.Err() != nil {
		return 0, "", rows.Err()
	}
	return ststus, result, nil
}

func (r *Req) SelectCodeMP_Contract(query string) (string, error) {
	if r.ctx.Err() != nil {
		return "", r.ctx.Err()
	}
	ctx, cancel := context.WithTimeout(r.ctx, time.Duration(r.queryTimeOut)* time.Second)
    defer cancel()
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var result string

	for rows.Next() {
		err := rows.Scan(&result)
		if err != nil {
			return "", err
		}
	}

	if rows.Err() != nil {
		return "", rows.Err()
	}

	return result, nil
}
