package recorder

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/bluenviron/mediacommon/pkg/formats/fmp4"
	"github.com/bluenviron/mediacommon/pkg/formats/fmp4/seekablebuffer"

	errorsql "github.com/bluenviron/mediamtx/internal/errorSQL"
	"github.com/bluenviron/mediamtx/internal/logger"
)

func writeInit(f io.Writer, tracks []*formatFMP4Track) error {
	fmp4Tracks := make([]*fmp4.InitTrack, len(tracks))
	for i, track := range tracks {
		fmp4Tracks[i] = track.initTrack
	}

	init := fmp4.Init{
		Tracks: fmp4Tracks,
	}

	var buf seekablebuffer.Buffer
	err := init.Marshal(&buf)
	if err != nil {
		return err
	}

	_, err = f.Write(buf.Bytes())
	return err
}

type formatFMP4Segment struct {
	f        *formatFMP4
	startDTS time.Duration
	startNTP time.Time

	path    string
	fi      *os.File
	curPart *formatFMP4Part
	lastDTS time.Duration
}

func (s *formatFMP4Segment) initialize() {
	s.lastDTS = s.startDTS
}

func (s *formatFMP4Segment) close() error {
	var err error
	s.f.ai.endTime = time.Now().Format("2006-01-02 15:04:05")
	if s.curPart != nil {
		err = s.curPart.close()
	}
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
						if strings.Contains(s.f.ai.stor.Sql.InsertPath, "code_mp_cam"){
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

func (s *formatFMP4Segment) write(track *formatFMP4Track, sample *sample) error {
	s.lastDTS = sample.dts

	if s.curPart == nil {
		s.curPart = &formatFMP4Part{
			s:              s,
			sequenceNumber: s.f.nextSequenceNumber,
			startDTS:       sample.dts,
		}
		s.curPart.initialize()
		s.f.nextSequenceNumber++
	} else if s.curPart.duration() >= s.f.ai.agent.PartDuration {
		err := s.curPart.close()
		s.curPart = nil

		if err != nil {
			return err
		}

		s.curPart = &formatFMP4Part{
			s:              s,
			sequenceNumber: s.f.nextSequenceNumber,
			startDTS:       sample.dts,
		}
		s.curPart.initialize()
		s.f.nextSequenceNumber++
	}

	return s.curPart.write(track, sample)
}
