package storage

import (
	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/bluenviron/mediamtx/internal/storage/psql"
)

type Storage struct {
	Use                  bool
	Req                  psql.Requests
	DbDrives             bool
	DbUseCodeMP_Contract bool
	DbUseContract        bool
	UseDbPathStream      bool
	UseUpdaterStatus     bool
	UseSrise             bool
	UseProxy             bool
	FileSQLErr           string
	Sql                  conf.Sql
}
