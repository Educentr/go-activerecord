package foo

import "github.com/Educentr/go-activerecord/v3/internal/pkg/parser/testdata/ds"

type Beer struct{}

type Foo struct {
	Key      string
	Bar      ds.AppInfo
	BeerData []Beer
	MapData  map[string]any
}
