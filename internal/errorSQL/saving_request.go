package errorsql

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func (f *Filesqlerror) SavingRequest(filename string, message []byte) error {
	data := time.Now().Format("2006-01-02")
	if f.File == nil {
		f.Data = data
		f.File.Close()
		filename = fmt.Sprintf("%s_%s.txt", strings.Split(filename, ".txt")[0], data)
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		f.File = file
	}
	if data != f.Data {
		f.Data = data
		f.File.Close()
		filename = fmt.Sprintf("%s_%s.txt", strings.Split(filename, ".txt")[0], data)
		file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		f.File = file
	}
	_, err := f.File.Write(message)
	if err != nil {
		return err
	}
	return nil
}
