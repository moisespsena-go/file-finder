package ffinder

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dustin/go-humanize"
)



type InputFinder interface {
	Setup(ctx context.Context) (err error)
	Finder() Finder
}

type File interface {
	fmt.Stringer
	os.FileInfo
	GetPath() string
	Reader() (r io.ReadCloser, err error)
}

type RealFile struct {
	os.FileInfo
	Path       string
	RealPath   string
	ReaderFunc func() (r io.ReadCloser, err error)
}

func (this *RealFile) String() (s string) {
	if this == nil {
		return
	}
	return this.Name() + " <" + humanize.Bytes(uint64(this.Size())) +
		" " + this.Mode().String() +
		" " + this.ModTime().Format(time.RFC3339)+">"
}

func (this RealFile) GetPath() string {
	return this.Path
}

func (this RealFile) Reader() (r io.ReadCloser, err error) {
	return this.ReaderFunc()
}

func NewRealFile(path, realPath string, info os.FileInfo) *RealFile {
	return &RealFile{
		info,
		path,
		realPath,
		func() (r io.ReadCloser, err error) {
			return os.Open(realPath)
		},
	}
}
