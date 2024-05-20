package logger

import (
	"fmt"
	"os"
	"strings"
)

func NewLoggerStream(level Level, destinations []Destination, filePath string, nameStream string, logdir string) (*Logger, error) {
	var errf error
	lh := &Logger{
		level: level,
	}

	for _, destType := range destinations {
		switch destType {
		case DestinationStdout:
			lh.destinations = append(lh.destinations, newDestionationStdout())

		case DestinationFile:
			if logdir == "" {
				filePath = fmt.Sprintf("%s_%s.log", strings.Split(filePath, ".log")[0], nameStream)
			} else {
				if _, err := os.Stat(logdir); os.IsNotExist(err) {
					errf = fmt.Errorf("the %s directory does not exist", logdir)
					filePath = fmt.Sprintf("%s_%s.log", strings.Split(filePath, ".log")[0], nameStream)
				} else {
					filePath = fmt.Sprintf("%s/%s_%s.log", logdir, strings.Split(strings.Split(filePath, ".log")[0], "/")[len(strings.Split(filePath, "/"))-1], nameStream)
				}

			}
			dest, err := newDestinationFile(filePath)
			if err != nil {
				lh.Close()
				return nil, err
			}
			lh.destinations = append(lh.destinations, dest)

		case DestinationSyslog:
			dest, err := newDestinationSyslog()
			if err != nil {
				lh.Close()
				return nil, err
			}
			lh.destinations = append(lh.destinations, dest)
		}
	}
	return lh, errf

}
