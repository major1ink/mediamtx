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
	TimeStatus           int
	UseUpdaterStatus     bool
	UseSrise             bool
	UseProxy             bool
	Login                string
	Pass                 string
	FileSQLErr           string
	Sql                  conf.Sql
}
