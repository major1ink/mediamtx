package recorder

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
	"github.com/bluenviron/mediamtx/internal/recordstore"
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
	if !p.s.f.ai.agent.Pathrecord {
		p.s.f.ai.agent.ChConfigSet <- []struct {
			Name   string
			Record bool
		}{{Name: p.s.f.ai.agent.PathName, Record: false}}
		err := fmt.Errorf("status_record = 0")
		return err

	}
	if p.s.fi == nil {
		var err error
			switch{
			case p.s.f.ai.agent.ClientGRPC.Use:
			if p.s.f.ai.switches.GetDrives	{	
			p.s.f.ai.agent.Log(logger.Debug, "sending a request to receive disks")
			r,err := p.s.f.ai.agent.ClientGRPC.Select(p.s.f.ai.agent.StreamName,"MountPoint")
			if err != nil {
			p.s.f.ai.agent.Log(logger.Error, "%v", err)
			p.localCreatePath()
			} else {
				p.s.f.ai.agent.Log(logger.Debug, "The result of executing the query: %v", r.MapDisks)
				if len (r.MapDisks) == 0 {
					p.s.f.ai.agent.Log(logger.Error, "ERROR:  No values were received in response to the request")
					p.localCreatePath()
				} else {
					drives:=[]interface{}{}
					for path := range r.MapDisks {
						drives = append(drives, path)
					}
					p.s.f.ai.free = getMostFreeDisk(drives)
					p.s.f.ai.idDsk = strconv.Itoa(int(r.MapDisks[p.s.f.ai.free]))
					p.CreatingPaths()
				}

			}}
			if p.s.f.ai.switches.UsePathStream && p.s.f.ai.agent.PathStream != "0" {
			p.s.f.ai.agent.Log(logger.Debug, "A request has been sent to receive Cod_mp and status_record")
			r, err :=p.s.f.ai.agent.ClientGRPC.Select(p.s.f.ai.agent.StreamName, "CodeMP")
			if err != nil {
				p.s.f.ai.agent.Log(logger.Error, "%s", err)
				p.s.f.ai.agent.Status_record=0
				p.s.f.ai.agent.PathStream="0"
			} else {
				p.s.f.ai.agent.Log(logger.Debug, "response received from GRPS: %s", r)
				p.s.f.ai.agent.PathStream = r.CodeMP
				p.s.f.ai.agent.Status_record = int8(r.StatusRecord)
				if p.s.f.ai.agent.Status_record == 0 {
					p.s.f.ai.agent.ChConfigSet <- []struct {
						Name   string
						Record bool
					}{{Name: p.s.f.ai.agent.PathName, Record: false}}
					err := fmt.Errorf("status_record = 0")
					return err
				}
			}
		}

				if p.s.f.ai.agent.Switches.UseCodeMP_Contract {
			p.s.f.ai.agent.Log(logger.Debug, "A request has been sent to receive CodeMP_Contract")
			r,err:= p.s.f.ai.agent.ClientGRPC.Select(p.s.f.ai.agent.StreamName, "CodeMP_Contract")
			if err != nil {
				p.s.f.ai.agent.Log(logger.Error, "%s", err)
				p.s.f.ai.agent.CodeMp="0"
			} else {
				p.s.f.ai.agent.Log(logger.Debug, "response received from GRPS: %s", r)
			}
		}

			case p.s.f.ai.stor.Use:
				if p.s.f.ai.switches.GetDrives	{
			p.s.f.ai.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", p.s.f.ai.stor.Sql.GetDrives))
			data, err := p.s.f.ai.stor.Req.SelectData(p.s.f.ai.stor.Sql.GetDrives)
			if err != nil {
				//записываем ошибку в лог и пробуем создать путь по локальному пути
				p.s.f.ai.agent.Log(logger.Error, "%v", err)
				p.localCreatePath()
			} else {
				p.s.f.ai.agent.Log(logger.Debug, "The result of executing the sql query: %v", data)
				if len(data) == 0 {
					p.s.f.ai.agent.Log(logger.Error, "ERROR:  No values were received in response to the request")
					p.localCreatePath()
				} else {
					idDisks := make(map[string]int16)
					drives := []interface{}{}
					for _, line := range data {
						idDisks[line[1].(string)] = line[0].(int16)
						drives = append(drives, line[1].(string))
					}
					p.s.f.ai.free = getMostFreeDisk(drives)
					p.s.f.ai.idDsk = strconv.Itoa(int(idDisks[p.s.f.ai.free]))

					p.CreatingPaths()
				}

			}
			}

			if p.s.f.ai.agent.PathStream == "0" && p.s.f.ai.switches.UsePathStream {
			p.s.f.ai.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", fmt.Sprintf(p.s.f.ai.agent.Stor.Sql.GetPathStream, p.s.f.ai.agent.StreamName)))
			p.s.f.ai.agent.Status_record, p.s.f.ai.agent.PathStream, err = p.s.f.ai.agent.Stor.Req.SelectPathStream(fmt.Sprintf(p.s.f.ai.agent.Stor.Sql.GetPathStream, p.s.f.ai.agent.StreamName))
			if err != nil {
				p.s.f.ai.agent.PathStream = "0"
				p.s.f.ai.agent.Status_record = 0
				p.s.f.ai.agent.Log(logger.Error, "%s", err)
			} else {
				p.s.f.ai.agent.Log(logger.Debug, "The result of executing the sql query: %b, %s", p.s.f.ai.agent.Status_record, p.s.f.ai.agent.PathStream)
				if p.s.f.ai.agent.Status_record == 0 {
					p.s.f.ai.agent.ChConfigSet <- []struct {
						Name   string
						Record bool
					}{{Name: p.s.f.ai.agent.PathName, Record: false}}
					err := fmt.Errorf("status_record = 0")
					return err
				}
			}

		}
		if p.s.f.ai.agent.CodeMp == "0" && p.s.f.ai.switches.UseCodeMP_Contract {
			p.s.f.ai.agent.Log(logger.Debug, fmt.Sprintf("SQL query sent:%s", fmt.Sprintf(p.s.f.ai.agent.Stor.Sql.GetCodeMP, p.s.f.ai.agent.StreamName)))
			p.s.f.ai.agent.CodeMp, err = p.s.f.ai.agent.Stor.Req.SelectCodeMP_Contract(fmt.Sprintf(p.s.f.ai.agent.Stor.Sql.GetCodeMP, p.s.f.ai.agent.StreamName))
			if err != nil {
				p.s.f.ai.agent.Log(logger.Error, "%s", err)
				p.s.f.ai.agent.CodeMp = "0"
			} else {
				p.s.f.ai.agent.Log(logger.Debug, "The result of executing the sql query: %s", p.s.f.ai.agent.CodeMp)
			}
		}

			default:
				p.localCreatePath()
		}
		p.s.f.ai.agent.Log(logger.Debug, "creating segment %s", p.s.path)
		// p.s.path = recordstore.Path{Start: p.s.startNTP}.Encode(p.s.f.ai.pathFormat)
		// p.s.f.ai.Log(logger.Debug, "creating segment %s", p.s.path)

		err = os.MkdirAll(filepath.Dir(p.s.path), 0o755)
		if err != nil {
			return err
		}

		fi, err := os.Create(p.s.path)
		if err != nil {
			return err
		}

		p.s.f.ai.agent.OnSegmentCreate(p.s.path)

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

func (p *formatFMP4Part) CreatingPaths() {
	if p.s.f.ai.switches.UseCodeMP_Contract {
		if p.s.f.ai.agent.CodeMp != "0" {
			p.s.path = fmt.Sprintf(p.s.f.ai.free+recordstore.Path{Start: p.s.startNTP}.Encode(p.s.f.ai.pathFormat), p.s.f.ai.agent.CodeMp)
			return
		}
		p.s.path = fmt.Sprintf(p.s.f.ai.free+recordstore.Path{Start: p.s.startNTP}.Encode(p.s.f.ai.pathFormat), "code_mp_cam")
		return
	}
	if p.s.f.ai.switches.UsePathStream {
		if p.s.f.ai.agent.PathStream != "0" {
			p.s.path = fmt.Sprintf(p.s.f.ai.free+recordstore.Path{Start: p.s.startNTP}.Encode(p.s.f.ai.pathFormat), p.s.f.ai.agent.PathStream)
			return
		}
		p.s.path = fmt.Sprintf(p.s.f.ai.free+recordstore.Path{Start: p.s.startNTP}.Encode(p.s.f.ai.pathFormat), "stream")
		return
	}
		p.s.path = fmt.Sprintf(p.s.f.ai.free + recordstore.Path{Start: p.s.startNTP}.Encode(p.s.f.ai.pathFormat))
}

func (p *formatFMP4Part) localCreatePath() {
	if len(p.s.f.ai.agent.PathFormats) == 0 {
		p.s.path = recordstore.Path{Start: p.s.startNTP}.Encode(p.s.f.ai.pathFormat)
	} else {
		if p.s.f.ai.stor.Use || p.s.f.ai.agent.ClientGRPC.Use{
			p.s.f.ai.free = getMostFreeDiskGroup(p.s.f.ai.agent.PathFormats)
			p.s.f.ai.idDsk = p.s.f.ai.agent.PathFormats[p.s.f.ai.free]
			p.CreatingPaths()
		} else {
			p.s.f.ai.free = getMostFreeDiskGroup(p.s.f.ai.agent.PathFormats)
			p.s.path = fmt.Sprintf(p.s.f.ai.free + recordstore.Path{Start: p.s.startNTP}.Encode(p.s.f.ai.pathFormat))
			p.s.f.ai.idDsk = "0"
		}
	}
}
