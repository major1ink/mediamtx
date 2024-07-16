package errorsql

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func  SavingRequest(dir, query, name string) error {
	data := time.Now().Format("2006-01-02")
	file, err := createFile(data, dir, query, name)
	if err != nil {
		return err
	}
	defer file.Close()
	message := []byte(fmt.Sprintf("%s,\n", strings.Split(query, "VALUES")[1]))
	_, err = file.Write(message)
	if err != nil {
		return err
	}
	return nil
}

func  createFile(data, dir, query, name string) (*os.File, error) {
	dir = fmt.Sprintf("%s/%s", dir, data)
	if strings.Contains(query, "pathStream") {
		dir = fmt.Sprintf("%s/%s/%s.txt", dir, "PathStream",name)
	} else {
		dir = fmt.Sprintf("%s/%s/%s.txt", dir, "Stream", name)
	}
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(filepath.Dir(dir), os.ModePerm)
			if err != nil {
				return nil,err 
			}
			file, err := os.OpenFile(dir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				return nil,err 
			}
			insert := []byte(fmt.Sprintf("%sVALUES\n", strings.Split(query, "VALUES")[0]))
			_, err = file.Write(insert)
			if err != nil {
				return nil,err 
			}
			return file,nil

		} else {
			return nil,err 
		}
	}
	file, err := os.OpenFile(dir, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil,err 
	}
	return file, nil
}
