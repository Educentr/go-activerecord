package foo

import "github.com/Educentr/go-activerecord/internal/pkg/parser/testdata/ds"

type Beer struct{}

type Foo struct {
	Key      string
	Bar      ds.AppInfo
	BeerData []Beer
	MapData  map[string]any
}
