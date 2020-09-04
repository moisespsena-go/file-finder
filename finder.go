package ffinder

import "context"

type Finder interface {
	Find(cb func(file File) error) (err error)
}

type Finders []Finder

func (this Finders) Find(cb func(file File) error) (err error) {
	for _, f := range this {
		if err = f.Find(cb); err != nil {
			return
		}
	}
	return
}

type InputFinders []InputFinder

func (this InputFinders) Setup(ctx context.Context) (err error) {
	for _, f := range this {
		if err = f.Setup(ctx); err != nil {
			return
		}
	}
	return
}

func (this InputFinders) Finder() Finder {
	var finders Finders
	for _, f := range this {
		finder := f.Finder()
		if sl, ok := finder.(Finders); ok {
			finders = append(finders, sl...)
		} else {
			finders = append(finders, finder)
		}
	}
	return finders
}