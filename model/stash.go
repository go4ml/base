package model

import (
	"go4ml.xyz/base/fu"
	"go4ml.xyz/iokit"
	"go4ml.xyz/zorros"
	"io"
)

type ModelStash struct {
	iteration int
	pattern   string
	files     []iokit.TemporaryFile
}

func NewStash(histlen int, pattern string) *ModelStash {
	return &ModelStash{
		pattern: pattern,
		files:   make([]iokit.TemporaryFile, histlen+1),
	}
}

func (ms *ModelStash) Length() int {
	return fu.Mini(ms.iteration+1, len(ms.files))
}

func (ms *ModelStash) Output(iteration int) (out iokit.Output, err error) {
	ms.iteration = iteration
	f := ms.files[ms.iteration%len(ms.files)]
	if f == nil {
		if f, err = iokit.Tempfile(ms.pattern); err != nil {
			return
		}
		ms.files[ms.iteration%len(ms.files)] = f
	} else {
		if err = f.Truncate(); err != nil {
			return
		}
	}
	return iokit.Writer(f.(io.Writer)), nil
}

func (ms *ModelStash) Reader(iteration int) (rd io.Reader, err error) {
	if iteration > ms.iteration || (ms.iteration-iteration) > len(ms.files) {
		return nil, zorros.Errorf("iteration %v is out of stash [%v,%v]",
			iteration,
			fu.Maxi(ms.iteration-len(ms.files), 0),
			ms.iteration)
	}
	f := ms.files[iteration%len(ms.files)]
	if err = f.Reset(); err != nil {
		return
	}
	return f, nil
}

func (ms *ModelStash) Close() error {
	for _, f := range ms.files {
		if f != nil {
			f.Close()
		}
	}
	return nil
}
