package record

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bluenviron/mediamtx/internal/logger"
)

type formatMPEGTSSegment struct {
	f        *formatMPEGTS
	startDTS time.Duration
	startNTP time.Time

	lastFlush time.Duration
	path      string
	fi        *os.File
}

func (s *formatMPEGTSSegment) initialize() {
	s.lastFlush = s.startDTS
	s.f.dw.setTarget(s)
}

func (s *formatMPEGTSSegment) close() error {
	s.f.a.endTime = time.Now().Format("2006-01-02 15:04:05")
	err := s.f.bw.Flush()

	if s.fi != nil {
		s.f.a.agent.Log(logger.Debug, "closing segment %s", s.path)
		err2 := s.fi.Close()
		if err == nil {
			err = err2
		}

		if err2 == nil {
			s.f.a.agent.OnSegmentComplete(s.path)

			if s.f.a.stor.Use {
				stat, err3 := os.Stat(s.path)

				if err3 == nil {
					paths := strings.Split(s.path, "/")
					pathRec := strings.Join(paths[:len(paths)-1], "/")
					var query string
					if s.f.a.stor.UseDbPathStream {
						query = fmt.Sprintf(
							s.f.a.stor.Sql.InsertPath,
							pathRec+"/",
							paths[len(paths)-1],
							s.f.a.timeStart,
							fmt.Sprint(stat.Size()),
							s.f.a.agent.PathStream,
							s.f.a.endTime,
							s.f.a.free,
						)
					} else {
						query = fmt.Sprintf(
							s.f.a.stor.Sql.InsertPath,
							pathRec+"/",
							paths[len(paths)-1],
							s.f.a.timeStart,
							fmt.Sprint(stat.Size()),
							s.f.a.agent.PathName,
							s.f.a.endTime,
							s.f.a.free,
						)	
					}

					s.f.a.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", query))
					err4 := s.f.a.stor.Req.ExecQuery(query)
				
					if err4 != nil {
						if err4.Error() == "context canceled" {
							err4 = s.f.a.stor.Req.ExecQueryNoCtx(query)
							if err4 != nil {
								s.f.a.agent.Log(logger.Error, "%v", err4)
								errsql:= s.f.a.agent.Filesqlerror.SavingRequest(s.f.a.stor.FileSQLErr, query)
								if errsql != nil {
									s.f.a.agent.Log(logger.Error, "ERROR: error when saving an incomplete sql query: %v", errsql)
								}
								return err4
							}
							s.f.a.agent.Log(logger.Debug, "The request was successfully completed")
							return err
						}
						s.f.a.agent.Log(logger.Error, "%v", err4)
						errsql:= s.f.a.agent.Filesqlerror.SavingRequest(s.f.a.stor.FileSQLErr, query)
						if errsql != nil {
							s.f.a.agent.Log(logger.Error, "ERROR: error when saving an incomplete sql query: %v", errsql)
						}
						return err
					}
					s.f.a.agent.Log(logger.Debug, "The request was successfully completed")
					return err
				}
				err = err3
			}
		}
	}

	return err
}

func (s *formatMPEGTSSegment) Write(p []byte) (int, error) {
	if s.fi == nil {
		var err error
		if s.f.a.stor.DbDrives {
			// проверка на использование бд, если бд не используеться будет записываться по локальным путям
			if s.f.a.stor.Use {
				s.f.a.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", s.f.a.stor.Sql.GetDrives))
				data, err := s.f.a.stor.Req.SelectData(s.f.a.stor.Sql.GetDrives)
				if err != nil {
					//записываем ошибку в лог и пробуем создать путь по локальному пути
					s.f.a.agent.Log(logger.Error, "%v", err)
					s.localCreatePath()
				} else {
					s.f.a.agent.Log(logger.Debug, "The result of executing the sql query: %v", data)
					if len(data) == 0 {
						s.f.a.agent.Log(logger.Error, "ERROR:  No values were received in response to the request")
						s.localCreatePath()
					} else {
						drives := []interface{}{}
						for _, line := range data {
							drives = append(drives, line[0].(string))
						}
						s.f.a.free = getMostFreeDisk(drives)
						s.dbCreatingPaths()
					}

				}

			} else {
				s.localCreatePath()
			}
		} else {
			s.localCreatePath()
		}

		s.f.a.agent.Log(logger.Debug, "creating segment %s", s.path)

		s.f.a.timeStart = s.startNTP.Format("2006-01-02 15:04:05")

		err = os.MkdirAll(filepath.Dir(s.path), 0o755)
		if err != nil {
			return 0, err
		}

		fi, err := os.Create(s.path)
		if err != nil {
			return 0, err
		}

		s.f.a.agent.OnSegmentCreate(s.path)

		s.fi = fi
	}

	return s.fi.Write(p)
}

func (s *formatMPEGTSSegment) dbCreatingPaths() {
	if s.f.a.stor.DbUseCodeMP_Contract && s.f.a.stor.UseDbPathStream {
		s.path = fmt.Sprintf(s.f.a.free+Path{Start: s.startNTP}.Encode(s.f.a.pathFormat), s.f.a.agent.CodeMp, s.f.a.agent.PathStream)
	} else {

		if s.f.a.stor.DbUseCodeMP_Contract {

			s.path = fmt.Sprintf(s.f.a.free+Path{Start: s.startNTP}.Encode(s.f.a.pathFormat), s.f.a.agent.CodeMp)
		}
		if s.f.a.stor.UseDbPathStream {
			s.path = fmt.Sprintf(s.f.a.free+Path{Start: s.startNTP}.Encode(s.f.a.pathFormat), s.f.a.agent.PathStream)
		}
		if !s.f.a.stor.DbUseCodeMP_Contract && !s.f.a.stor.UseDbPathStream {
			s.path = fmt.Sprintf(s.f.a.free + Path{Start: s.startNTP}.Encode(s.f.a.pathFormat))
		}
	}
}

func (s *formatMPEGTSSegment) localCreatePath() {
	if len(s.f.a.agent.PathFormats) == 0 {
		s.path = Path{Start: s.startNTP}.Encode(s.f.a.pathFormat)
	} else {
		if s.f.a.stor.Use {
			s.f.a.free = getMostFreeDiskGroup(s.f.a.agent.PathFormats)
			s.dbCreatingPaths()
		} else {
			s.f.a.free = getMostFreeDiskGroup(s.f.a.agent.PathFormats)
			s.path = fmt.Sprintf(s.f.a.free + Path{Start: s.startNTP}.Encode(s.f.a.pathFormat))
		}
	}
}
