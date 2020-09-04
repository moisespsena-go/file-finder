package ffinder

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type ContextKey uint8

const (
	LoadDirConfigFile ContextKey = iota
)

func WithDirConfigFile(name string, ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, LoadDirConfigFile, name)
}

func GetDirConfigFile(ctx context.Context) (name string) {
	if ctx != nil {
		if value := ctx.Value(LoadDirConfigFile); value != nil {
			return value.(string)
		}
	}
	return
}

type Input interface {
	InputFinder
	GetTrimPrefix() string
	GetPrefix() string
	IsRecursive() bool
	Include(pth string, info os.FileInfo) bool
	Exclude(pth string, info os.FileInfo) bool
}

type DirInput struct {
	Dir        string
	TrimPrefix string `mapstructure:"strip_prefix" yaml:"strip_prefix"`
	Prefix     string
	Recursive  bool
	Includes   FileMatchers
	Excludes   FileMatchers
	Sources    []DirInput
}

func (this DirInput) Setup(ctx context.Context) (err error) {
	return this.SetupTree(ctx, "")
}

func (this *DirInput) SetupTree(ctx context.Context, prefix string) (err error) {
	if cfgName := GetDirConfigFile(ctx); cfgName != "" {
		for _, ext := range []string{"yaml", "yml"} {
			pth := filepath.Join(this.Dir, cfgName+"."+ext)
			if r, err := os.Open(pth); err == nil {
				if err = yaml.NewDecoder(r).Decode(this); err != nil {
					return fmt.Errorf("decode `%v` failed: %v", pth, err.Error())
				}
			} else if !os.IsNotExist(err) {
				return err
			}
		}
	}
	if this.TrimPrefix == "." {
		if this.Dir == "." {
			this.TrimPrefix = ""
		} else {
			this.TrimPrefix = this.Dir
		}
	}
	this.Prefix = path.Join(prefix, this.Prefix)
	for _, e := range this.Sources {
		if err = e.SetupTree(ctx, this.Prefix); err != nil {
			return
		}
	}
	return
}

func (this DirInput) Finder() (finder Finder) {
	return this.finder("")
}

func (this DirInput) finder(prefix string) (finders Finders) {
	finders = append(finders, &this)
	for _, e := range this.Sources {
		finders = append(finders, e.finder(this.Prefix)...)
	}
	return
}

func (this DirInput) PathOf(realPath string) (pth string) {
	tpf := this.TrimPrefix
	if tpf == "" {
		tpf = this.Dir
	}
	pth = strings.TrimPrefix(filepath.ToSlash(realPath), this.TrimPrefix)
	if this.Prefix != "" {
		pth = path.Join(this.Prefix, pth)
	} else {
		pth = path.Clean(pth)
	}
	return
}

func (this DirInput) Find(cb func(file File) error) (err error) {
	if this.Recursive {
		return filepath.Walk(this.Dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if this.Include(path, info) || !this.Exclude(path, info) {
				if !info.IsDir() {
					return cb(NewRealFile(this.PathOf(path), path, info))
				}
			} else if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		})
	}
	var files []os.FileInfo
	if files, err = ioutil.ReadDir(this.Dir); err != nil {
		return
	}
	for _, info := range files {
		if info.IsDir() {
			continue
		}
		if this.Include(info.Name(), info) || !this.Exclude(info.Name(), info) {
			if err = cb(NewRealFile(this.PathOf(info.Name()), filepath.Join(this.Dir, info.Name()), info)); err != nil {
				return
			}
		}
	}
	return
}

func (this DirInput) GetTrimPrefix() string {
	return this.TrimPrefix
}

func (this DirInput) GetPrefix() string {
	return this.Prefix
}

func (this DirInput) IsRecursive() bool {
	return this.Recursive
}

func (this DirInput) Include(pth string, info os.FileInfo) bool {
	return this.Includes.Match(pth, info)
}

func (this DirInput) Exclude(pth string, info os.FileInfo) bool {
	return this.Excludes.Match(pth, info)
}
