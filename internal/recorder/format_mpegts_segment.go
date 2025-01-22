package recorder

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	errorsql "github.com/bluenviron/mediamtx/internal/errorSQL"
	"github.com/bluenviron/mediamtx/internal/logger"
	"github.com/bluenviron/mediamtx/internal/recordstore"
)

type formatMPEGTSSegment struct {
	f        *formatMPEGTS
	startDTS time.Duration
	startNTP time.Time

	path      string
	fi        *os.File
	lastFlush time.Duration
	lastDTS   time.Duration
}

func (s *formatMPEGTSSegment) initialize() {
	s.lastFlush = s.startDTS
	s.lastDTS = s.startDTS
	s.f.dw.setTarget(s)
}

func (s *formatMPEGTSSegment) close() error {
	s.f.ai.endTime = time.Now().Format("2006-01-02 15:04:05")
	err := s.f.bw.Flush()

	if s.fi != nil {
		s.f.ai.Log(logger.Debug, "closing segment %s", s.path)
		err2 := s.fi.Close()
		if err == nil {
			err = err2
		}
		if err2 == nil {

			duration := s.lastDTS - s.startDTS
			s.f.ai.agent.OnSegmentComplete(s.path, duration)

			if s.f.ai.agent.ClientGRPC.Use {
				stat, err3 := os.Stat(s.path)
				if err3 == nil {
					paths := strings.Split(s.path, "/")
					pathRec := strings.Join(paths[:len(paths)-1], "/")
					var query string
					var attribute string
					if s.f.ai.switches.UsePathStream && s.f.ai.agent.PathStream != "0"{
						attribute = "pathStream"
						query = fmt.Sprintf("(%s,'%s','%s','%s','%s','%s','%s')",
							s.f.ai.agent.PathStream,
							pathRec+"/",
							paths[len(paths)-1],
							s.f.ai.timeStart,
							fmt.Sprint(stat.Size()),
							s.f.ai.endTime,
							s.f.ai.idDsk,
						)
					} else {
						if s.f.ai.agent.ClientGRPC.UseCodeMPAttribute{
							attribute = "code_mp_cam"
						} else {
							attribute = "stream"
						}
						query = fmt.Sprintf("(%s,'%s','%s','%s','%s','%s','%s')",
							s.f.ai.agent.PathName,
							pathRec+"/",
							paths[len(paths)-1],
							s.f.ai.timeStart,
							fmt.Sprint(stat.Size()),
							s.f.ai.endTime,
							s.f.ai.idDsk,
						)
					}
					s.f.ai.agent.Log(logger.Debug, "Sending an insert request to RMS:Server %s, atribute %s, query %s",s.f.ai.agent.ClientGRPC.Server, attribute, query)
					err4 := s.f.ai.agent.ClientGRPC.Post(attribute, query)
					
					if err4 != nil {
						if s.f.ai.switches.UsePathStream && s.f.ai.agent.PathStream != "0"	{
							query = fmt.Sprintf(
							s.f.ai.stor.Sql.InsertPathStream,
							s.f.ai.agent.PathStream,
							pathRec+"/",
							paths[len(paths)-1],
							s.f.ai.timeStart,
							fmt.Sprint(stat.Size()),
							s.f.ai.endTime,
							s.f.ai.idDsk,
						)
						} else {
							query = fmt.Sprintf(
							s.f.ai.stor.Sql.InsertPath,
							s.f.ai.agent.PathName,
							pathRec+"/",
							paths[len(paths)-1],
							s.f.ai.timeStart,
							fmt.Sprint(stat.Size()),
							s.f.ai.endTime,
							s.f.ai.idDsk,
						)
						}
						s.f.ai.agent.Log(logger.Error, "ERROR: error when sending an insert request to RMS: %s", err4)
						errsql := errorsql.SavingRequest(s.f.ai.switches.FileSQLErr, query,s.f.ai.agent.PathName)
						if errsql != nil {
							s.f.ai.agent.Log(logger.Error, "ERROR: error when saving an incomplete sql query: %v", errsql)
						}
						return err
					} else {
						s.f.ai.agent.Log(logger.Debug, "The request was successfully completed")
					}
					
				}
				
			}

			if s.f.ai.stor.Use  && !s.f.ai.agent.ClientGRPC.Use {
				stat, err3 := os.Stat(s.path)

				if err3 == nil {
					paths := strings.Split(s.path, "/")
					pathRec := strings.Join(paths[:len(paths)-1], "/")
					var query string
					if s.f.ai.switches.UsePathStream && s.f.ai.agent.PathStream != "0" {
						query = fmt.Sprintf(
							s.f.ai.stor.Sql.InsertPathStream,
							s.f.ai.agent.PathStream,
							pathRec+"/",
							paths[len(paths)-1],
							s.f.ai.timeStart,
							fmt.Sprint(stat.Size()),
							s.f.ai.endTime,
							s.f.ai.idDsk,
						)

					} else {
						query = fmt.Sprintf(
							s.f.ai.stor.Sql.InsertPath,
							s.f.ai.agent.PathName,
							pathRec+"/",
							paths[len(paths)-1],
							s.f.ai.timeStart,
							fmt.Sprint(stat.Size()),
							s.f.ai.endTime,
							s.f.ai.idDsk,
						)
					}

					s.f.ai.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", query))
					err4 := s.f.ai.stor.Req.ExecQuery(query)
						
					if err4 != nil {
						if err4.Error() == "context canceled" {
							err4 = s.f.ai.stor.Req.ExecQueryNoCtx(query)
							if err4 != nil {
								s.f.ai.agent.Log(logger.Error, "%v", err4)
								errsql := errorsql.SavingRequest(s.f.ai.switches.FileSQLErr, query,s.f.ai.agent.PathName)
								if errsql != nil {
									s.f.ai.agent.Log(logger.Error, "ERROR: error when saving an incomplete sql query: %v", errsql)
								}
								return err4
							}
							s.f.ai.agent.Log(logger.Debug, "The request was successfully completed")
							return err
						}
						s.f.ai.agent.Log(logger.Error, "%v", err4)
						errsql := errorsql.SavingRequest(s.f.ai.switches.FileSQLErr, query,s.f.ai.agent.PathName)
						if errsql != nil {
							s.f.ai.agent.Log(logger.Error, "ERROR: error when saving an incomplete sql query: %v", errsql)
						}
						return err
					}
					s.f.ai.agent.Log(logger.Debug, "The request was successfully completed")
					return err
				}
				err = err3
			}

		}
	}
	return err
}

func (s *formatMPEGTSSegment) Write(p []byte) (int, error) {
	if !s.f.ai.agent.Pathrecord {
		s.f.ai.agent.ChConfigSet <- []struct {
			Name   string
			Record bool
		}{{Name: s.f.ai.agent.PathName, Record: false}}
		err := fmt.Errorf("status_record = 0")
		return 0, err

	}
	if s.fi == nil {
		var err error
		switch{
			case s.f.ai.agent.ClientGRPC.Use:
				
			if s.f.ai.switches.GetDrives	{	
			s.f.ai.agent.Log(logger.Debug, "sending a request to receive disks")
			r,err := s.f.ai.agent.ClientGRPC.Select(s.f.ai.agent.StreamName,"MountPoint")
			if err != nil {
			s.f.ai.agent.Log(logger.Error, "%v", err)
			s.localCreatePath()
			} else {
				s.f.ai.agent.Log(logger.Debug, "The result of executing the query: %v", r.MapDisks)
				if len (r.MapDisks) == 0 {
					s.f.ai.agent.Log(logger.Error, "ERROR:  No values were received in response to the request")
					s.localCreatePath()
				} else {
					drives:=[]interface{}{}
					for path := range r.MapDisks {
						drives = append(drives, path)
					}
					s.f.ai.free = getMostFreeDisk(drives)
					s.f.ai.idDsk = strconv.Itoa(int(r.MapDisks[s.f.ai.free]))
					s.CreatingPaths()
				}

			}} else {
				s.localCreatePath()
				
			}
			if s.f.ai.switches.UsePathStream && s.f.ai.agent.PathStream == "0" {
			s.f.ai.agent.Log(logger.Debug, "A request has been sent to receive Cod_mp and status_record")
			r, err :=s.f.ai.agent.ClientGRPC.Select(s.f.ai.agent.StreamName, "CodeMP")
			if err != nil {
				s.f.ai.agent.Log(logger.Error, "%s", err)
				s.f.ai.agent.Status_record=1
				s.f.ai.agent.PathStream="0"
			} else {
				s.f.ai.agent.Log(logger.Debug, "response received from GRPS: %s", r)
				s.f.ai.agent.PathStream = r.CodeMP
				s.f.ai.agent.Status_record = int8(r.StatusRecord)
				if s.f.ai.agent.Status_record == 0 {
					s.f.ai.agent.ChConfigSet <- []struct {
						Name   string
						Record bool
					}{{Name: s.f.ai.agent.PathName, Record: false}}
					err := fmt.Errorf("status_record = 0")
					return 0, err
				}
			}
		}

				if s.f.ai.agent.Switches.UseCodeMP_Contract  && s.f.ai.agent.CodeMp == "0" {
			s.f.ai.agent.Log(logger.Debug, "A request has been sent to receive CodeMP_Contract")
			r,err:= s.f.ai.agent.ClientGRPC.Select(s.f.ai.agent.StreamName, "CodeMP_Contract")
			if err != nil {
				s.f.ai.agent.Log(logger.Error, "%s", err)
				s.f.ai.agent.CodeMp="0"
			} else {
				s.f.ai.agent.Log(logger.Debug, "response received from GRPS: %s", r)
			}
		}

			case s.f.ai.stor.Use:
				if s.f.ai.switches.GetDrives	{
			s.f.ai.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", s.f.ai.stor.Sql.GetDrives))
			data, err := s.f.ai.stor.Req.SelectData(s.f.ai.stor.Sql.GetDrives)
			if err != nil {
				//записываем ошибку в лог и пробуем создать путь по локальному пути
				s.f.ai.agent.Log(logger.Error, "%v", err)
				s.localCreatePath()
			} else {
				s.f.ai.agent.Log(logger.Debug, "The result of executing the sql query: %v", data)
				if len(data) == 0 {
					s.f.ai.agent.Log(logger.Error, "ERROR:  No values were received in response to the request")
					s.localCreatePath()
				} else {
					idDisks := make(map[string]int16)
					drives := []interface{}{}
					for _, line := range data {
						idDisks[line[1].(string)] = line[0].(int16)
						drives = append(drives, line[1].(string))
					}
					s.f.ai.free = getMostFreeDisk(drives)
					s.f.ai.idDsk = strconv.Itoa(int(idDisks[s.f.ai.free]))

					s.CreatingPaths()
				}

			}
			} else {
				s.localCreatePath()
			}

			if s.f.ai.agent.PathStream == "0" && s.f.ai.switches.UsePathStream {
			s.f.ai.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", fmt.Sprintf(s.f.ai.agent.Stor.Sql.GetPathStream, s.f.ai.agent.StreamName)))
			s.f.ai.agent.Status_record, s.f.ai.agent.PathStream, err = s.f.ai.agent.Stor.Req.SelectPathStream(fmt.Sprintf(s.f.ai.agent.Stor.Sql.GetPathStream, s.f.ai.agent.StreamName))
			if err != nil {
				s.f.ai.agent.PathStream = "0"
				s.f.ai.agent.Status_record = 1
				s.f.ai.agent.Log(logger.Error, "%s", err)
			} else {
				s.f.ai.agent.Log(logger.Debug, "The result of executing the sql query: %b, %s", s.f.ai.agent.Status_record, s.f.ai.agent.PathStream)
				if s.f.ai.agent.Status_record == 0 {
					s.f.ai.agent.ChConfigSet <- []struct {
						Name   string
						Record bool
					}{{Name: s.f.ai.agent.PathName, Record: false}}
					err := fmt.Errorf("status_record = 0")
					return 0, err
				}
			}

		}
		if s.f.ai.agent.CodeMp == "0" && s.f.ai.switches.UseCodeMP_Contract {
			s.f.ai.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", fmt.Sprintf(s.f.ai.agent.Stor.Sql.GetCodeMP, s.f.ai.agent.StreamName)))
			s.f.ai.agent.CodeMp, err = s.f.ai.agent.Stor.Req.SelectCodeMP_Contract(fmt.Sprintf(s.f.ai.agent.Stor.Sql.GetCodeMP, s.f.ai.agent.StreamName))
			if err != nil {
				s.f.ai.agent.Log(logger.Error, "%s", err)
				s.f.ai.agent.CodeMp = "0"
			} else {
				s.f.ai.agent.Log(logger.Debug, "The result of executing the sql query: %s", s.f.ai.agent.CodeMp)
			}
		}

			default:
				s.localCreatePath()
		}

		

		// s.f.ai.agent.Log(logger.Debug, "creating segment %s", s.path)
		// s.path = recordstore.Path{Start: s.startNTP}.Encode(s.f.ai.pathFormat)
		s.f.ai.Log(logger.Debug, "creating segment %s", s.path)

		s.f.ai.timeStart = s.startNTP.Format("2006-01-02 15:04:05")

		err = os.MkdirAll(filepath.Dir(s.path), 0o755)
		if err != nil {
			return 0, err
		}

		fi, err := os.Create(s.path)
		if err != nil {
			return 0, err
		}


		s.f.ai.agent.OnSegmentCreate(s.path)


		s.fi = fi
	}
	return s.fi.Write(p)
}

func (s *formatMPEGTSSegment) CreatingPaths() {
	if s.f.ai.switches.UseCodeMP_Contract {
		if s.f.ai.agent.CodeMp != "0" {
			s.path = fmt.Sprintf(s.f.ai.free+recordstore.Path{Start: s.startNTP}.Encode(s.f.ai.pathFormat), s.f.ai.agent.CodeMp)
			return
		}
		s.path = fmt.Sprintf(s.f.ai.free+recordstore.Path{Start: s.startNTP}.Encode(s.f.ai.pathFormat), "code_mp_cam")
		return
	}
	if s.f.ai.switches.UsePathStream {
		if s.f.ai.agent.PathStream != "0" {
			s.path = fmt.Sprintf(s.f.ai.free+recordstore.Path{Start: s.startNTP}.Encode(s.f.ai.pathFormat), s.f.ai.agent.PathStream)
			return
		}
		s.path = fmt.Sprintf(s.f.ai.free+recordstore.Path{Start: s.startNTP}.Encode(s.f.ai.pathFormat), "stream")
		return
	}
		s.path = fmt.Sprintf(s.f.ai.free + recordstore.Path{Start: s.startNTP}.Encode(s.f.ai.pathFormat))

}

func (s *formatMPEGTSSegment) localCreatePath() {
	if len(s.f.ai.agent.PathFormats) == 0 {
		s.path = recordstore.Path{Start: s.startNTP}.Encode(s.f.ai.pathFormat)
	} else {
		if s.f.ai.stor.Use || s.f.ai.agent.ClientGRPC.Use{
			s.f.ai.free = getMostFreeDiskGroup(s.f.ai.agent.PathFormats)
			s.f.ai.idDsk = s.f.ai.agent.PathFormats[s.f.ai.free]
			s.CreatingPaths()
		} else {
			s.f.ai.free = getMostFreeDiskGroup(s.f.ai.agent.PathFormats)
			s.path = fmt.Sprintf(s.f.ai.free + recordstore.Path{Start: s.startNTP}.Encode(s.f.ai.pathFormat))
			s.f.ai.idDsk = "0"
		}
	}
}
