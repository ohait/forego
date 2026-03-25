package enc

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ohait/forego/ctx"
)

type Bytes []byte

var _ Node = Bytes(nil)

func (this Bytes) native() any {
	return []byte(this)
}

func (this Bytes) MarshalJSON() ([]byte, error) {
	return json.Marshal([]byte(this))
}

func (this Bytes) GoString() string {
	return fmt.Sprintf("enc.Bytes(%#v)", []byte(this))
}

func (this Bytes) String() string {
	return base64.StdEncoding.EncodeToString(this)
}

func (this Bytes) unmarshalInto(c ctx.C, handler Handler, into reflect.Value) error {
	switch into.Kind() {
	case reflect.Interface:
		into.Set(reflect.ValueOf([]byte(this)))
		return nil
	case reflect.Slice:
		if into.Type().Elem().Kind() != reflect.Uint8 {
			break
		}
		out := make([]byte, len(this))
		copy(out, this)
		into.SetBytes(out)
		return nil
	case reflect.Array:
		if into.Type().Elem().Kind() != reflect.Uint8 {
			break
		}
		if into.Len() != len(this) {
			return ctx.NewErrorf(c, "expected %v, got %d bytes instead", into.Type(), len(this))
		}
		reflect.Copy(into, reflect.ValueOf([]byte(this)))
		return nil
	}
	return ctx.NewErrorf(c, "can't unmarshal %s %T into %v", handler.path, this, into.Type())
}
