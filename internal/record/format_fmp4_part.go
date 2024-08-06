package record

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
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
	if !p.s.f.a.agent.Pathrecord {
		p.s.f.a.agent.ChConfigSet <- []struct {
			Name   string
			Record bool
		}{{Name: p.s.f.a.agent.PathName, Record: false}}
		err := fmt.Errorf("status_record = 0")
		return err

	}
	if p.s.fi == nil {
		if p.s.f.a.stor.DbDrives && p.s.f.a.stor.Use {
			// проверка на использование бд, если бд не используеться будет записываться по локальным путям

			p.s.f.a.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", p.s.f.a.stor.Sql.GetDrives))
			data, err := p.s.f.a.stor.Req.SelectData(p.s.f.a.stor.Sql.GetDrives)
			if err != nil {
				//записываем ошибку в лог и пробуем создать путь по локальному пути
				p.s.f.a.agent.Log(logger.Error, "%v", err)
				p.localCreatePath()
			} else {
				p.s.f.a.agent.Log(logger.Debug, "The result of executing the sql query: %v", data)
				if len(data) == 0 {
					p.s.f.a.agent.Log(logger.Error, "ERROR:  No values were received in response to the request")
					p.localCreatePath()
				} else {
					idDisks := make(map[string]int32)
					drives := []interface{}{}
					for _, line := range data {
						idDisks[line[1].(string)] = line[0].(int32)
						drives = append(drives, line[1].(string))
					}
					p.s.f.a.free = getMostFreeDisk(drives)
					p.s.f.a.idDsk = strconv.Itoa(int(idDisks[p.s.f.a.free]))
					p.dbCreatingPaths()
				}

			}
		} else {
			p.localCreatePath()
		}
		var err error
		if p.s.f.a.agent.PathStream == "0" && p.s.f.a.stor.Use && p.s.f.a.stor.UseDbPathStream {
			p.s.f.a.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", fmt.Sprintf(p.s.f.a.agent.Stor.Sql.GetPathStream, p.s.f.a.agent.StreamName)))
			p.s.f.a.agent.Status_record, p.s.f.a.agent.PathStream, err = p.s.f.a.agent.Stor.Req.SelectPathStream(fmt.Sprintf(p.s.f.a.agent.Stor.Sql.GetPathStream, p.s.f.a.agent.StreamName))
			if err != nil {
				p.s.f.a.agent.Log(logger.Error, "%s", err)
				p.s.f.a.agent.PathStream = "0"
				p.s.f.a.agent.Status_record = 0
			} else {
				p.s.f.a.agent.Log(logger.Debug, "The result of executing the sql query: %b, %s", p.s.f.a.agent.Status_record, p.s.f.a.agent.PathStream)
				if p.s.f.a.agent.Status_record == 0 {
					p.s.f.a.agent.ChConfigSet <- []struct {
						Name   string
						Record bool
					}{{Name: p.s.f.a.agent.PathName, Record: false}}
					err := fmt.Errorf("status_record = 0")
					return err
				}
			}

		}
		if p.s.f.a.agent.CodeMp == "0" && p.s.f.a.stor.Use && p.s.f.a.stor.DbUseCodeMP_Contract {
			p.s.f.a.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", fmt.Sprintf(p.s.f.a.agent.Stor.Sql.GetCodeMP, p.s.f.a.agent.StreamName)))
			p.s.f.a.agent.CodeMp, err = p.s.f.a.agent.Stor.Req.SelectCodeMP_Contract(fmt.Sprintf(p.s.f.a.agent.Stor.Sql.GetCodeMP, p.s.f.a.agent.StreamName))
				if err != nil {
					p.s.f.a.agent.Log(logger.Error, "%s", err)
					p.s.f.a.agent.CodeMp = "0"
				} else {
					p.s.f.a.agent.Log(logger.Debug, "The result of executing the sql query: %s", p.s.f.a.agent.CodeMp)
				}
		}

		p.s.f.a.agent.Log(logger.Debug, "creating segment %s", p.s.path)

		err = os.MkdirAll(filepath.Dir(p.s.path), 0o755)
		if err != nil {
			return err
		}

		fi, err := os.Create(p.s.path)
		if err != nil {
			return err
		}
		p.s.f.a.timeStart = p.s.startNTP.Format("2006-01-02 15:04:05")
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
	if p.s.f.a.stor.DbUseCodeMP_Contract {
		if p.s.f.a.agent.CodeMp != "0" {
			p.s.path = fmt.Sprintf(p.s.f.a.free+Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat), p.s.f.a.agent.CodeMp)
			return
		}
		p.s.path = fmt.Sprintf(p.s.f.a.free+Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat), fmt.Sprintf("code_mp_cam/%v", p.s.f.a.agent.PathName))
		return
	}
	if p.s.f.a.stor.UseDbPathStream {
		if p.s.f.a.agent.PathStream != "0" {
			p.s.path = fmt.Sprintf(p.s.f.a.free+Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat), p.s.f.a.agent.PathStream)
			return
		}
		p.s.path = fmt.Sprintf(p.s.f.a.free+Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat), fmt.Sprintf("stream/%v", p.s.f.a.agent.PathName))
		return
	}
		p.s.path = fmt.Sprintf(p.s.f.a.free + Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat))
}

func (p *formatFMP4Part) localCreatePath() {
	if len(p.s.f.a.agent.PathFormats) == 0 {
		p.s.path = Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat)
	} else {
		if p.s.f.a.stor.Use {
			p.s.f.a.free = getMostFreeDiskGroup(p.s.f.a.agent.PathFormats)
			for id, path := range p.s.f.a.agent.PathFormats {
				if p.s.f.a.free == path {
					p.s.f.a.idDsk = id
					break
				}
			}
			p.dbCreatingPaths()
		} else {
			p.s.f.a.free = getMostFreeDiskGroup(p.s.f.a.agent.PathFormats)
			p.s.path = fmt.Sprintf(p.s.f.a.free + Path{Start: p.s.startNTP}.Encode(p.s.f.a.pathFormat))
			p.s.f.a.idDsk = "0"
		}
	}
}
