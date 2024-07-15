package record

import (
	"testing"
	"time"

	"github.com/bluenviron/mediamtx/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestLocalCreatePathMPEGTSSegment(t *testing.T) {
	// Создаем экземпляр формата MPEGTS
	pathFormats := make(map[string]string)
	pathFormats["./recordings"] = "1"
	format := &formatMPEGTSSegment{
		f: &formatMPEGTS{
			a: &agentInstance{
				agent: &Agent{
					PathFormats: pathFormats,
					CodeMp:      "1",
					PathStream:  "1",
				},
				pathFormat: "/%v/%v/%Y-%m-%d/%path-%s-%f",
				stor: storage.Storage{
					Use:                  true,
					DbUseCodeMP_Contract: true,
					UseDbPathStream:      true,
				},
			},
		},
		startNTP: time.Date(2008, 11, 7, 11, 22, 4, 123456000, time.Local),
	}
	format.localCreatePath()
	require.Equal(t, "./recordings/1/1/2008-11-07/-1226046124-123456", format.path)
}
