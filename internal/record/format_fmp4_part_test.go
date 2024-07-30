package record

import (
	"testing"
	"time"

	"github.com/bluenviron/mediamtx/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestLocalCreatePathFMP4Part(t *testing.T) {

 func () {
 	pathFormats := make(map[string]string)
	pathFormats["./recordings"] = "1"
	format := &formatFMP4Part{
		s: &formatFMP4Segment{
			f: &formatFMP4{
				a: &agentInstance{
					agent: &Agent{
						PathFormats: pathFormats,
						PathStream:  "1",
					},
					pathFormat: "/%v/%Y-%m-%d/%path-%s-%f",
					stor: storage.Storage{
						Use:                  true,
						UseDbPathStream:      true,
					},
				},
			},
			startNTP: time.Date(2008, 11, 7, 11, 22, 4, 123456000, time.Local),
		},
	}
	format.localCreatePath()
	require.Equal(t, "./recordings/1/2008-11-07/-1226046124-123456", format.s.path)
	}()

 func () {
 	pathFormats := make(map[string]string)
	pathFormats["./recordings"] = "1"
	format := &formatFMP4Part{
		s: &formatFMP4Segment{
			f: &formatFMP4{
				a: &agentInstance{
					agent: &Agent{
						PathFormats: pathFormats,
						CodeMp:  "1",
					},
					pathFormat: "/%v/%Y-%m-%d/%path-%s-%f",
					stor: storage.Storage{
						Use:                  true,
						DbUseCodeMP_Contract:      true,
					},
				},
			},
			startNTP: time.Date(2008, 11, 7, 11, 22, 4, 123456000, time.Local),
		},
	}
	format.localCreatePath()
	require.Equal(t, "./recordings/1/2008-11-07/-1226046124-123456", format.s.path)
	}()
}
