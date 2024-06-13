package errorsql

import (
	"os"
	"time"
)

type Filesqlerror struct {
	File *os.File
	Data string
}

func CreateFilesqlerror() *Filesqlerror {
	return &Filesqlerror{
		Data: time.Now().Format("2006-01-02"),
	}
}

func (f *Filesqlerror) CloseFile() error {
	return f.File.Close()
}