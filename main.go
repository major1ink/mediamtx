// main executable.
package main

import (
	"os"

	"github.com/bluenviron/mediamtx/internal/core"
)

func main() {
	// memProfileTicker := time.Tick(15 * time.Second)
	// go func() {
	// 	for range memProfileTicker {
	// 		f, err := os.Create("memprofile.prof")
	// 		if err != nil {
	// 			panic(err)
	// 		}

	// 		err = pprof.WriteHeapProfile(f)
	// 		if err != nil {
	// 			panic(err)
	// 		}

	// 		f.Close()
	// 	}
	// }()
	s, ok := core.New(os.Args[1:])
	if !ok {
		os.Exit(1)
	}
	s.Wait()
}
