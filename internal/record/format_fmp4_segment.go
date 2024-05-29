package record

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/bluenviron/mediacommon/pkg/formats/fmp4"
	"github.com/bluenviron/mediacommon/pkg/formats/fmp4/seekablebuffer"

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
}

func (s *formatFMP4Segment) initialize() {
}

func (s *formatFMP4Segment) close() error {
	var err error
	s.f.a.endTime = time.Now().Format("2006-01-02 15:04:05")
	if s.curPart != nil {
		err = s.curPart.close()
	}

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
								errsql := s.f.a.agent.Filesqlerror.SavingRequest(s.f.a.stor.FileSQLErr, query)
								s.f.a.agent.Log(logger.Error, "ERROR: error when saving an incomplete sql query: %v", errsql)
								return err4
							}
							s.f.a.agent.Log(logger.Debug, "The request was successfully completed")
							return err
						}
						s.f.a.agent.Log(logger.Error, "%v", err4)
						errsql :=s.f.a.agent.Filesqlerror.SavingRequest(s.f.a.stor.FileSQLErr, query)
						s.f.a.agent.Log(logger.Error, "ERROR: error when saving an incomplete sql query: %v", errsql)
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

func (s *formatFMP4Segment) record(track *formatFMP4Track, sample *sample) error {
	if s.curPart == nil {
		s.curPart = &formatFMP4Part{
			s:              s,
			sequenceNumber: s.f.nextSequenceNumber,
			startDTS:       sample.dts,
		}
		s.curPart.initialize()
		s.f.nextSequenceNumber++
	} else if s.curPart.duration() >= s.f.a.agent.PartDuration {
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

	return s.curPart.record(track, sample)
}
