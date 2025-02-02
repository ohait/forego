package enc

import (
	"fmt"
	"reflect"

	"github.com/ohait/forego/ctx"
)

type String string

var _ Node = String("")

func (this String) native() any {
	return string(this)
}

func (this String) GoString() string {
	return fmt.Sprintf("enc.String{%q}", string(this))
}

func (this String) String() string {
	return fmt.Sprintf("%q", string(this))
}

func (this String) unmarshalInto(c ctx.C, handler Handler, into reflect.Value) error {
	//log.Debugf(c, "%v.unmarshalInto(%#v)", this, into)
	switch into.Kind() {
	case reflect.String:
		into.SetString(string(this))
	case reflect.Interface:
		v := reflect.ValueOf(this.native())
		into.Set(v)
	default:
		return ctx.NewErrorf(c, "can't unmarshal %s %T into %v", handler.path, this, into.Type())
	}
	return nil
}

func (this String) AsTime() (Time, error) {
	var t Time
	return t, t.Parse(string(this))
}

func (this String) MustTime() Time {
	var t Time
	err := t.Parse(string(this))
	if err != nil {
		panic(err)
	}
	return t
}

func (this String) AsDuration() (Duration, error) {
	var d Duration
	return d, d.Parse(string(this))
}

func (this String) MustDuration() Duration {
	var d Duration
	err := d.Parse(string(this))
	if err != nil {
		panic(err)
	}
	return d
}

type Bool bool

var _ Node = Bool(true)

func (this Bool) native() any {
	return bool(this)
}

func (this Bool) GoString() string {
	if this {
		return "enc.Bool{true}"
	} else {
		return "enc.Bool{false}"
	}
}

func (this Bool) String() string {
	if this {
		return "true"
	} else {
		return "false"
	}
}

func (this Bool) unmarshalInto(c ctx.C, handler Handler, into reflect.Value) error {
	//log.Debugf(c, "%v.unmarshalInto(%#v)", this, into)
	switch into.Kind() {
	case reflect.Bool:
		into.SetBool(bool(this))
	case reflect.Interface:
		into.Set(reflect.ValueOf(bool(this)))
	default:
		return ctx.NewErrorf(c, "can't unmarshal %s %T into %v", handler.path, this, into.Type())
	}
	return nil
}

/*
type Time time.Time

var _ Node = Time(time.Time{})

func (this Time) native() any {
	return time.Time(this)
}

func (this Time) GoString() string {
	return fmt.Sprintf("enc.Time{%s}", time.Time(this).Format(time.RFC3339))
}

func (this Time) String() string {
	return time.Time(this).Format(time.RFC3339)
}

func (this Time) unmarshalInto(c ctx.C, handler Handler, into reflect.Value) error {
	log.Debugf(c, "%v.unmarshalInto(%#v)", this, into)
	switch into.Kind() {
	case reflect.Struct:
		into.Set(reflect.ValueOf(time.Time(this)))
	default:
		return ctx.NewErrorf(c, "can't unmarshal %T into %v", this, into.Type())
	}
	return nil
}
*/
