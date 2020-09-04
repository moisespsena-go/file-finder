package main

import (
	"compress/gzip"
	"fmt"
	"io"

	ffinder "github.com/moisespsena-go/file-finder"
	"github.com/moisespsena-go/ibfs"
)

func main() {
	dir := &ffinder.DirInput{
		Dir:       ".",
		Recursive: true,
	}
	dir.Excludes.Parse("?.+(git|idea)/")
	if err := dir.Setup(nil); err != nil {
		panic(err)
	}

	var tree, err = ibfs.NewTree(ibfs.ImportFinder(dir.Finder()))

	if err != nil {
		panic(err)
	}

	var treew = tree.NewWriter()
	treew.Options.CompressorFactory = func(file *ibfs.File, dst io.Writer) (cw io.WriteCloser, err error) {
		gz := gzip.NewWriter(dst)
		return gz, nil
	}
	treew.Options.Compressable = func(file *ibfs.File) bool {
		return file.Name() == "README.md"
	}
	treew.HWalkCheckSum(func(file *ibfs.File, h ibfs.HashHeader) error {
		fmt.Println(file.GetPath(), "->", fmt.Sprintf("%x", h))
		return nil
	})
}

func fileToString(f *ffinder.File) {

}
