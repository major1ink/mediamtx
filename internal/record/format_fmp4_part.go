package record

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bluenviron/mediacommon/pkg/formats/fmp4"
	"github.com/bluenviron/mediacommon/pkg/formats/fmp4/seekablebuffer"

	"github.com/bluenviron/mediamtx/internal/logger"
)

func writePart(
	f io.Writer,
	sequenceNumber uint32,
	partTracks map[*formatFMP4Track]*fmp4.PartTrack,
) error {
	fmp4PartTracks := make([]*fmp4.PartTrack, len(partTracks))
	i := 0
	for _, partTrack := range partTracks {
		fmp4PartTracks[i] = partTrack
		i++
	}

	part := &fmp4.Part{
		SequenceNumber: sequenceNumber,
		Tracks:         fmp4PartTracks,
	}

	var buf seekablebuffer.Buffer
	err := part.Marshal(&buf)
	if err != nil {
		return err
	}

	_, err = f.Write(buf.Bytes())
	return err
}

type formatFMP4Part struct {
	s              *formatFMP4Segment
	sequenceNumber uint32
	startDTS       time.Duration

	partTracks map[*formatFMP4Track]*fmp4.PartTrack
	endDTS     time.Duration
}

func (p *formatFMP4Part) initialize() {
	p.partTracks = make(map[*formatFMP4Track]*fmp4.PartTrack)
}

func (p *formatFMP4Part) close() error {
	if p.s.fi == nil {
		if p.s.f.a.stor.DbDrives {
			// проверка на использование бд, если бд не используеться будет записываться по локальным путям
			if p.s.f.a.stor.Use {
				p.s.f.a.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", p.s.f.a.stor.Sql.GetDrives))
				data, err := p.s.f.a.stor.Req.SelectData(p.s.f.a.stor.Sql.GetDrives)
				if err != nil {
					//записываем ошибку в лог и пробуем создать путь по локальному пути
					p.s.f.a.agent.Log(logger.Error, "%v", err)
					p.localCreatePath()
				} else {
					p.s.f.a.agent.Log(logger.Debug, "ERROR: The result of executing the sql query: %v", data)
					if len(data) == 0 {
						p.localCreatePath()
					} else {
						drives := []interface{}{}
						for _, line := range data {
							drives = append(drives, line[0].(string))
						}
						p.s.f.a.free = getMostFreeDisk(drives)
						p.dbCreatingPaths()
					}

				}
			} else {
				p.localCreatePath()
			}
		} else {
			p.localCreatePath()
		}

		p.s.f.a.agent.Log(logger.Debug, "creating segment %s", p.s.path)

		err := os.MkdirAll(filepath.Dir(p.s.path), 0o755)
		if err != nil {
			return err
		}

		fi, err := os.Create(p.s.path)
		if err != nil {
			return err
		}
		p.s.f.a.timeStart = p.s.startNTP.Format("2006-01-02 15:04:05")
		if p.s.f.a.stor.Use {
			paths := strings.Split(p.s.path, "/")
			pathRec := strings.Join(paths[:len(paths)-1], "/")
			if p.s.f.a.stor.UseDbPathStream {
				query := fmt.Sprintf(
					p.s.f.a.stor.Sql.InsertPath,
					pathRec+"/",
					paths[len(paths)-1],
					p.s.f.a.timeStart,
					p.s.f.a.agent.PathStream,
					p.s.f.a.free,
				)
				p.s.f.a.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", query))
				err := p.s.f.a.stor.Req.ExecQuery(query)
				if err != nil {
					p.s.f.a.agent.Log(logger.Error, "%v", err)
					message := []byte(query + "\n")
					p.s.f.a.agent.Filesqlerror.SavingRequest(p.s.f.a.stor.FileSQLErr, message)
				} else {
					p.s.f.a.agent.Log(logger.Debug, "The request was successfully completed")
				}
			} else {
				query := fmt.Sprintf(
					p.s.f.a.stor.Sql.InsertPath,
					pathRec+"/",
					paths[len(paths)-1],
					p.s.f.a.timeStart,
					p.s.f.a.agent.PathName,
					p.s.f.a.free,
				)
				p.s.f.a.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", query))
				err := p.s.f.a.stor.Req.ExecQuery(query)
				if err != nil {
					p.s.f.a.agent.Log(logger.Error, "%v", err)
					message := []byte(query + "\n")
					p.s.f.a.agent.Filesqlerror.SavingRequest(p.s.f.a.stor.FileSQLErr, message)
				} else {
					p.s.f.a.agent.Log(logger.Debug, "The request was successfully completed")
				}
			}

		}

		p.s.f.a.agent.OnSegmentCreate(p.s.path)

		err = writeInit(fi, p.s.f.tracks)
		if err != nil {
			fi.Close()
			return err
		}

		p.s.fi = fi

	}
	return writePart(p.s.fi, p.sequenceNumber, p.partTracks)
}

func (p *formatFMP4Part) write(track *formatFMP4Track, sample *sample) error {
	partTrack, ok := p.partTracks[track]
	if !ok {
		partTrack = &fmp4.PartTrack{
			ID:       track.initTrack.ID,
			BaseTime: durationGoToMp4(sample.dts-p.s.startDTS, track.initTrack.TimeScale),
		}
		p.partTracks[track] = partTrack
	}

	partTrack.Samples = append(partTrack.Samples, sample.PartSample)
	p.endDTS = sample.dts

	return nil
}

func (p *formatFMP4Part) duration() time.Duration {
	return p.endDTS - p.startDTS
}

func (p *formatFMP4Part) dbCreatingPaths() {
	if p.s.f.a.stor.DbUseCodeMP_Contract && p.s.f.a.stor.UseDbPathStream {
		p.s.path = fmt.Sprintf(p.s.f.a.free+Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat), p.s.f.a.agent.CodeMp, p.s.f.a.agent.PathStream)
	}

	if p.s.f.a.stor.UseDbPathStream {
		p.s.path = fmt.Sprintf(p.s.f.a.free+Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat), p.s.f.a.agent.PathStream)
	}

	if p.s.f.a.stor.DbUseCodeMP_Contract {
		p.s.path = fmt.Sprintf(p.s.f.a.free+Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat), p.s.f.a.agent.CodeMp)
	}

	if !p.s.f.a.stor.DbUseCodeMP_Contract && !p.s.f.a.stor.UseDbPathStream {
		p.s.path = fmt.Sprintf(p.s.f.a.free + Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat))
	}
}

func (p *formatFMP4Part) localCreatePath() {
	if len(p.s.f.a.agent.PathFormats) == 0 {
		p.s.path = Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat)
	} else {
		if p.s.f.a.stor.Use {
			p.s.f.a.free = getMostFreeDiskGroup(p.s.f.a.agent.PathFormats)
			p.dbCreatingPaths()
		} else {
			p.s.f.a.free = getMostFreeDiskGroup(p.s.f.a.agent.PathFormats)
			p.s.path = fmt.Sprintf(p.s.f.a.free + Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat))
		}
	}
}
