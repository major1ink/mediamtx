package record

import (
	"testing"
	"time"

	"github.com/bluenviron/mediamtx/internal/conf"
	"github.com/bluenviron/mediamtx/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestLocalCreatePathMPEGTSSegment(t *testing.T) {
	// Создаем экземпляр формата MPEGTS
	func () {
	pathFormats := make(map[string]string)
	pathFormats["./recordings"] = "1"
	format := &formatMPEGTSSegment{
		f: &formatMPEGTS{
			a: &agentInstance{
				agent: &Agent{
					PathFormats: pathFormats,
					PathStream:  "1",
				},
				pathFormat: "/%v/%Y-%m-%d/%path-%s-%f",
				stor: storage.Storage{
					Use:                  true,
				},
				switches: conf.Switches{
						UsePathStream: true,
					},
			},
		},
		startNTP: time.Date(2008, 11, 7, 11, 22, 4, 123456000, time.Local),
	}
	format.localCreatePath()
	require.Equal(t, "./recordings/1/2008-11-07/-1226046124-123456", format.path)	
	}()
	func () {
	pathFormats := make(map[string]string)
	pathFormats["./recordings"] = "1"
	format := &formatMPEGTSSegment{
		f: &formatMPEGTS{
			a: &agentInstance{
				agent: &Agent{
					PathFormats: pathFormats,
					CodeMp:  "1",
				},
				pathFormat: "/%v/%Y-%m-%d/%path-%s-%f",
				stor: storage.Storage{
					Use:                  true,
				},
				switches: conf.Switches{
					UseCodeMP_Contract: true,
					},
			},
		},
		startNTP: time.Date(2008, 11, 7, 11, 22, 4, 123456000, time.Local),
	}
	format.localCreatePath()
	require.Equal(t, "./recordings/1/2008-11-07/-1226046124-123456", format.path)	
	}()
	
}
