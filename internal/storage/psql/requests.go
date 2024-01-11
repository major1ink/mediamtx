package psql

func (r *Req) ExecQuery(query string) error {

	_, err := r.pool.Exec(r.ctx, query)
	if err != nil {
		return err
	}

	return nil
}

func (r *Req) SelectData(query string) ([][]interface{}, error) {
	if r.ctx.Err() != nil {
		return nil, r.ctx.Err()
	}

	rows, err := r.pool.Query(r.ctx, query)
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
