package record

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aler9/writerseeker"
	"github.com/bluenviron/mediacommon/pkg/formats/fmp4"

	"github.com/bluenviron/mediamtx/internal/logger"
)

var drives []interface{}

var free string

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

	var ws writerseeker.WriterSeeker
	err := part.Marshal(&ws)
	if err != nil {
		return err
	}

	_, err = f.Write(ws.Bytes())
	return err
}

type formatFMP4Part struct {
	s              *formatFMP4Segment
	sequenceNumber uint32
	startDTS       time.Duration

	created    time.Time
	partTracks map[*formatFMP4Track]*fmp4.PartTrack
	endDTS     time.Duration
}

func newFormatFMP4Part(
	s *formatFMP4Segment,
	sequenceNumber uint32,
	startDTS time.Duration,
) *formatFMP4Part {
	return &formatFMP4Part{
		s:              s,
		startDTS:       startDTS,
		sequenceNumber: sequenceNumber,
		created:        timeNow(),
		partTracks:     make(map[*formatFMP4Track]*fmp4.PartTrack),
	}
}

func (p *formatFMP4Part) close() error {
	if p.s.fi == nil {

		var err error

		if p.s.f.a.stor.DbDrives {

			p.s.f.a.pathStream, err = p.s.f.a.stor.Req.SelectPathStream(fmt.Sprintf(p.s.f.a.stor.Sql.GetPathStream, p.s.f.a.agent.StreamName))

			if err != nil {
				return err
			}

			data, err := p.s.f.a.stor.Req.SelectData(p.s.f.a.stor.Sql.GetDrives)

			if err != nil {
				return err
			}
			for _, line := range data {
				drives = append(drives, line[0].(string))
			}
			free = getMostFreeDisk(drives)

			if p.s.f.a.stor.DbUseCodeMP {
				p.s.f.a.codeMp, err = p.s.f.a.stor.Req.SelectPathStream(fmt.Sprintf(p.s.f.a.stor.Sql.GetCodeMP, p.s.f.a.agent.StreamName))
				if err != nil {
					return err
				}
				p.s.path = fmt.Sprintf(free+path(p.created).encode(p.s.f.a.pathFormat), p.s.f.a.codeMp, p.s.f.a.pathStream)
			} else {
				p.s.path = fmt.Sprintf(free+path(p.created).encode(p.s.f.a.pathFormat), p.s.f.a.pathStream)
			}

		} else {
			p.s.path = path(p.created).encode(p.s.f.a.pathFormat)
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

		if p.s.f.a.stor.Use {
			paths := strings.Split(p.s.path, "/")
			pathRec := strings.Join(paths[:len(paths)-1], "/")
			err := p.s.f.a.stor.Req.ExecQuery(
				fmt.Sprintf(
					p.s.f.a.stor.Sql.InsertPath,
					pathRec+"/",
					paths[len(paths)-1],
					time.Now().Format("2006-01-02 15:04:05"),
					p.s.f.a.pathStream,
				),
			)
			if err != nil {
				os.Remove(p.s.path)
				return err
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

func (p *formatFMP4Part) record(track *formatFMP4Track, sample *sample) error {
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
