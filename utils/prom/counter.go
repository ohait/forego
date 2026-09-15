package prom

import (
	"fmt"
	"io"

	"github.com/ohait/forego/utils/sync"
)

type Counter struct {
	Desc   string
	Labels []string
	val    sync.Map[string, *counter]
}

type counter struct {
	m   sync.Mutex
	sum float64
}

func (this *Counter) Observe(val float64, labels ...string) {
	this.val.GetOrStore(stringify(this.Labels, labels), &counter{}).observe(val)
}

func (this *counter) observe(val float64) {
	//defer log.Printf("(%p).Observe(%v)", this, val)
	this.m.Lock()
	defer this.m.Unlock()
	this.sum += val
}

func (this *Counter) Print(name string, w io.Writer) error {
	first := true
	return this.val.RangeErr(func(l string, v *counter) error {
		if first {
			first = false
			_, err := fmt.Fprintf(w, "# HELP %s %s\n", name, this.Desc)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "# TYPE %s counter\n", name)
			if err != nil {
				return err
			}
		}
		v.m.Lock()
		defer v.m.Unlock()
		if l == "" {
			_, err := fmt.Fprintf(w, "%s %f\n", name, v.sum)
			return err
		}
		_, err := fmt.Fprintf(w, "%s{%s} %f\n", name, l, v.sum)
		return err
	})
}
