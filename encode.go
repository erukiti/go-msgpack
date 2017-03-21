package msgpack

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
)

type Encoder struct {
	writer io.Writer
}

func NewEncoder(writer io.Writer) Encoder {
	return Encoder{
		writer,
	}
}

func (e *Encoder) WriteByte(b byte) {
	buf := []byte{b}
	e.writer.Write(buf)
}

func (e *Encoder) Encode(data interface{}) error {
	v := reflect.ValueOf(data)

	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			e.WriteByte(0xc0)
			return nil
		}

		v = v.Elem()
	}

	if !v.IsValid() {
		return fmt.Errorf("invalid")
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n := v.Int()

		if n < math.MinInt32 || n > math.MaxInt32 {
			e.WriteByte(0xd3)
			binary.Write(e.writer, binary.BigEndian, uint64(n))
		} else if n < math.MinInt16 || n > math.MaxInt16 {
			e.WriteByte(0xd2)
			binary.Write(e.writer, binary.BigEndian, uint32(n))
		} else if n < math.MinInt8 || n > math.MaxInt8 {
			e.WriteByte(0xd1)
			binary.Write(e.writer, binary.BigEndian, uint16(n))
		} else if n < -32 {
			e.WriteByte(0xd0)
			e.WriteByte(byte(n))
		} else if n >= -32 && n <= math.MaxInt8 {
			e.WriteByte(byte(n))
		} else {
			return fmt.Errorf("?")
		}
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n := v.Uint()

		if n > math.MaxInt32 {
			e.WriteByte(0xcf)
			binary.Write(e.writer, binary.BigEndian, uint64(n))
		} else if n > math.MaxInt16 {
			e.WriteByte(0xce)
			binary.Write(e.writer, binary.BigEndian, uint32(n))
		} else if n > math.MaxInt8 {
			e.WriteByte(0xcd)
			binary.Write(e.writer, binary.BigEndian, uint16(n))
		} else {
			e.WriteByte(0xcc)
			e.WriteByte(byte(n))
		}
		return nil

	case reflect.Bool:
		if v.Bool() {
			e.WriteByte(0xc3)
		} else {
			e.WriteByte(0xc2)
		}
		return nil

	case reflect.String:
		ln := v.Len()
		if ln <= 0x1f {
			e.WriteByte(byte(0xa0 + ln))
		} else if ln <= 0xff {
			e.WriteByte(byte(0xd9))
			e.WriteByte(byte(ln))
		} else if ln <= 0xffff {
			e.WriteByte(byte(0xda))
			binary.Write(e.writer, binary.BigEndian, uint16(ln))
		} else if ln <= 0xffffffff {
			e.WriteByte(byte(0xdb))
			binary.Write(e.writer, binary.BigEndian, uint32(ln))
		} else {
			return fmt.Errorf("String over length: %d", ln)
		}
		io.WriteString(e.writer, v.String())
		return nil

	case reflect.Float64:
		e.WriteByte(0xcb)
		binary.Write(e.writer, binary.BigEndian, math.Float64bits(v.Float()))
		return nil

	case reflect.Float32:
		e.WriteByte(0xca)
		binary.Write(e.writer, binary.BigEndian, math.Float32bits(float32(v.Float())))
		return nil

	case reflect.Slice:
		if v.IsNil() {
			return e.encodeArray(v, 0)
		} else {
			return e.encodeArray(v, v.Len())
		}

	case reflect.Array:
		return e.encodeArray(v, v.Len())

	case reflect.Map:
		var ln int
		if v.IsNil() {
			ln = 0
		} else {
			ln = v.Len()
		}

		if ln <= 0x0f {
			e.WriteByte(byte(0x80 + ln))
		} else if ln <= 0xffff {
			e.WriteByte(0xde)
			binary.Write(e.writer, binary.BigEndian, uint16(ln))
		} else if ln <= 0xffffffff {
			e.WriteByte(0xdf)
			binary.Write(e.writer, binary.BigEndian, uint32(ln))
		} else {
			return fmt.Errorf("Map over length: %d", ln)
		}

		for _, key := range v.MapKeys() {
			e.Encode(key.Interface())
			e.Encode(v.MapIndex(key).Interface())
		}
		return nil

	case reflect.Struct:
		ln := v.NumField()

		if ln <= 0x0f {
			e.WriteByte(byte(0x80 + ln))
		} else if ln <= 0xffff {
			e.WriteByte(0xde)
			binary.Write(e.writer, binary.BigEndian, uint16(ln))
		} else if ln <= 0xffffffff {
			e.WriteByte(0xdf)
			binary.Write(e.writer, binary.BigEndian, uint32(ln))
		} else {
			return fmt.Errorf("Struct over length: %d", ln)
		}

		for i := 0; i < ln; i++ {
			field := v.Type().Field(i)
			ar := strings.SplitN(field.Tag.Get("msgpack"), "=", 2)
			var name string
			if ar[0] == "" {
				name = field.Name
			} else {
				name = ar[0]
			}

			e.Encode(name)
			e.Encode(v.Field(i).Interface())
		}
		return nil
	}

	return fmt.Errorf("error: unknown")
}

func (e *Encoder) encodeBytes(v reflect.Value, ln int) error {
	if ln <= 0xff {
		e.WriteByte(0xc4)
		e.WriteByte(byte(ln))
	} else if ln <= 0xffff {
		e.WriteByte(0xc5)
		binary.Write(e.writer, binary.BigEndian, uint16(ln))
	} else if ln <= 0xffffffff {
		e.WriteByte(0xc6)
		binary.Write(e.writer, binary.BigEndian, uint16(ln))
	} else {
		return fmt.Errorf("bytes overflow length: %d", ln)
	}
	e.writer.Write(v.Slice(0, ln).Bytes())
	return nil
}

func (e *Encoder) encodeArray(v reflect.Value, ln int) error {
	if v.Type().Elem().String() == "uint8" {
		return e.encodeBytes(v, ln)
	}

	if ln <= 0x0f {
		e.WriteByte(byte(0xa0 + ln))
	} else if ln <= 0xffff {
		e.WriteByte(byte(0xdc))
		binary.Write(e.writer, binary.BigEndian, uint16(ln))
	} else if ln <= 0xffffffff {
		e.WriteByte(byte(0xdd))
		binary.Write(e.writer, binary.BigEndian, uint32(ln))
	} else {
		return fmt.Errorf("Array over length: %d", ln)
	}

	for i := 0; i < ln; i++ {
		err := e.Encode(v.Index(i).Interface())
		if err != nil {
			return fmt.Errorf("Array encode error: %s", err.Error())
		}
	}
	return nil
}
