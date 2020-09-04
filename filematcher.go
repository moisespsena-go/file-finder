package ffinder

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/gobwas/glob/syntax"

	"github.com/gobwas/glob"
)

type FileMatcher interface {
	Match(pth string, info os.FileInfo) bool
}

type FileMatchers []FileMatcher

func (this FileMatchers) Match(pth string, info os.FileInfo) bool {
	for _, e := range this {
		if e.Match(pth, info) {
			return true
		}
	}
	return false
}

func (this *FileMatchers) Parse(items ...string) (err error) {
	for _, s := range items {
		var m FileMatcher
		switch s[0] {
		case '#':
			continue

		default:
			if m, err = ParseFileMatcher(s); err != nil {
				return
			}
		}
		*this = append(*this, m)
	}
	return
}

func ParseFileMatchers(items ...string) (include, exclude FileMatchers, err error) {
	for i, s := range items {
		if s[0] == '!' {
			if err = exclude.Parse(s[1:]); err != nil {
				err = fmt.Errorf("parse#%d as exclude failed: %v", i, err)
				return nil, nil, err
			}
		} else {
			if err = include.Parse(s[1:]); err != nil {
				err = fmt.Errorf("parse#%d as include failed: %v", i, err)
				return nil, nil, err
			}
		}
	}
	return
}

func ParseFileMatcher(s string) (fm FileMatcher, err error) {
	var m FileMatcher
	isDir := strings.HasSuffix(s, "/")
	if isDir {
		s = strings.TrimSuffix(s, "/")
	}
	switch s[0] {
	case '?':
		// is regex
		if m, err = FileMatchRegex(s[1:]); err != nil {
			return nil, fmt.Errorf("parse regex %q: %v", s, err)
		}
	default:
		var special bool
		for _, r := range s {
			if syntax.Special(byte(r)) {
				special = true
				break
			}
		}
		if special {
			if m, err = FileMatchGlob(s); err != nil {
				return nil, fmt.Errorf("parse glob %q: %v", s, err)
			}
		} else {
			m = FileMatcherFunc(func(pth string, info os.FileInfo) bool {
				return info.Name() == s
			})
		}
	}

	if isDir {
		return FileMatcherFunc(func(pth string, info os.FileInfo) bool {
			if !info.IsDir() {
				return false
			}
			return m.Match(pth, info)
		}), nil
	}

	return FileMatcherFunc(func(pth string, info os.FileInfo) bool {
		if info.IsDir() {
			return false
		}
		return m.Match(pth, info)
	}), nil
}

type FileMatcherFunc func(pth string, info os.FileInfo) bool

func (this FileMatcherFunc) Match(pth string, info os.FileInfo) bool {
	return this(pth, info)
}

func FileMatchRegex(pattern string) (acceptor FileMatcher, err error) {
	if regexPattern, err := regexp.Compile(pattern); err != nil {
		return nil, err
	} else {
		return FileMatcherFunc(func(pth string, info os.FileInfo) bool {
			return regexPattern.MatchString(pth)
		}), nil
	}
}

func FileMatchGlob(pattern string) (acceptor FileMatcher, err error) {
	if globPattern, err := glob.Compile(pattern); err != nil {
		return nil, err
	} else {
		return FileMatcherFunc(func(pth string, info os.FileInfo) bool {
			return globPattern.Match(pth)
		}), nil
	}
}
