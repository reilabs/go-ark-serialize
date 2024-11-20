package go_ark_serialize

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

func deserializeSettable(data io.Reader, v reflect.Value, compress, validate bool) (int, error) {
	n := 0
	vPtr := v.Addr()
	if vD := vPtr.MethodByName("CanonicalDeserializeWithMode"); vD.IsValid() {
		vs := vD.Call([]reflect.Value{reflect.ValueOf(data), reflect.ValueOf(compress), reflect.ValueOf(validate)})
		if len(vs) != 2 {
			return n, fmt.Errorf("DeserializeCanonical: invalid return values in custom deserializer")
		}
		n := int(vs[0].Int())
		err := vs[1].Interface()
		if err != nil {
			return n, err.(error)
		}
		return n, nil
	}
	switch v.Kind() {
	case reflect.Uint8:
		buf := make([]byte, 1)
		r, err := data.Read(buf)
		n += r
		if err != nil {
			return n, err
		}
		v.SetUint(uint64(buf[0]))
		return n, nil
	case reflect.Uint64:
		buf := make([]byte, 8)
		r, err := data.Read(buf)
		n += r
		if err != nil {
			return n, err
		}
		v.SetUint(binary.LittleEndian.Uint64(buf))
		return n, nil
	case reflect.Array:
		for i := 0; i < v.Len(); i++ {
			r, err := deserializeSettable(data, v.Index(i), compress, validate)
			n += r
			if err != nil {
				return n, err
			}
		}
		return n, nil
	case reflect.Slice:
		var len uint64
		r, err := CanonicalDeserializeWithMode(data, &len, compress, validate)
		n += r
		if err != nil {
			return n, err
		}
		v.Set(reflect.MakeSlice(v.Type(), int(len), int(len)))
		for i := 0; i < int(len); i++ {
			r, err := deserializeSettable(data, v.Index(i), compress, validate)
			n += r
			if err != nil {
				return n, err
			}
		}
		return n, nil
	case reflect.Struct:
		for i := range v.NumField() {
			r, err := deserializeSettable(data, v.Field(i), compress, validate)
			n += r
			if err != nil {
				return n, err
			}
		}
		return n, nil
	}

	return n, fmt.Errorf("CanonicalDeserialize: unsupported type %v", v.Type())
}

func CanonicalDeserializeWithMode(data io.Reader, v any, compress, validate bool) (n int, err error) {
	vR := reflect.ValueOf(v)
	if vR.Kind() != reflect.Ptr {
		err = fmt.Errorf("CanonicalDeserialize: v must be a pointer")
		return
	}
	vE := vR.Elem()
	return deserializeSettable(data, vE, compress, validate)

}
