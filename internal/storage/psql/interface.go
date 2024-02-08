package psql

type Requests interface {
	ExecQuery(query string) error
	SelectData(query string) ([][]interface{}, error)
	SelectPathStream(query string) (string, error)
}
