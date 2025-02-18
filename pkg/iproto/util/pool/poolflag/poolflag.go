package poolflag

import (
	"bytes"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

type flagSetWrapper struct {
	*flag.FlagSet
}

func (f flagSetWrapper) Float64Slice(name string, def []float64, desc string) *[]float64 {
	v := new(float64slice)
	*v = float64slice(def)
	f.Var(v, name, desc)

	return (*[]float64)(v)
}

type float64slice []float64

func (f *float64slice) Set(v string) (err error) {
	var (
		values = strings.Split(v, ",")
		vs     = make([]float64, len(values))
	)

	for i, v := range values {
		vs[i], err = strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return err
		}
	}

	*f = float64slice(vs)

	return nil
}

func (f *float64slice) String() string {
	var buf bytes.Buffer

	for i, f := range *f {
		if i != 0 {
			buf.WriteString(", ")
		}

		fmt.Fprintf(&buf, "%f", f)
	}

	return buf.String()
}
