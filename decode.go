package msgpack

import (
	"fmt"
	"io"
	"reflect"
)

var (
	_interfaceType = reflect.TypeOf((*interface{})(nil)).Elem()
)

type Decoder struct {
	reader io.ByteReader
}

func NewDecoder(reader io.ByteReader) Decoder {
	return Decoder{
		reader,
	}
}

func (d *Decoder) next() byte {
	for {
		c, err := d.reader.ReadByte()
		if err != nil {
			// if err == io.EOF {
			// 	continue
			// }
			fmt.Printf("error: %s\n", err)
			panic("I/O error")
		}
		return c

	}
}

func (d *Decoder) decodeMap(n int) (interface{}, error) {
	var err error
	var keyType, valType reflect.Type

	keys := [16]interface{}{}
	keyType = nil
	vals := [16]interface{}{}
	valType = nil

	for i := 0; i < n; i++ {
		keys[i], err = d.decode()
		if err != nil {
			return nil, err
		}
		vals[i], err = d.decode()
		if err != nil {
			return nil, err
		}

		_keyType := reflect.TypeOf(keys[i])
		if _keyType.Kind() == reflect.Map {
			return nil, fmt.Errorf("Map is don't be map key.")
		}
		if keyType == nil {
			keyType = _keyType
		} else if keyType != _keyType {
			keyType = _interfaceType
		}

		_valType := reflect.TypeOf(vals[i])
		if valType == nil {
			valType = _valType
		} else if valType != _valType {
			valType = _interfaceType
		}
	}

	if keyType == nil {
		keyType = _interfaceType
	}
	if valType == nil {
		valType = _interfaceType
	}

	result := reflect.MakeMap(reflect.MapOf(keyType, valType))
	for i := 0; i < n; i++ {
		result.SetMapIndex(reflect.ValueOf(keys[i]), reflect.ValueOf(vals[i]))
	}

	return result.Interface(), nil
}

func (d *Decoder) decodeArray(n int) (interface{}, error) {
	var err error
	var _typeArray reflect.Type

	_typeArray = nil
	arr := [16]interface{}{}
	for i := 0; i < n; i++ {
		arr[i], err = d.decode()
		if err != nil {
			return nil, err
		}
		_type := reflect.TypeOf(arr[i])
		if _typeArray == nil {
			_typeArray = _type
		} else if _type != _typeArray {
			_typeArray = _interfaceType
		}
	}
	if _typeArray == nil {
		_typeArray = _interfaceType
	}
	result := reflect.MakeSlice(reflect.SliceOf(_typeArray), 0, 0)
	for i := 0; i < n; i++ {
		result = reflect.Append(result, reflect.ValueOf(arr[i]))
	}
	return result.Interface(), nil
}

func (d *Decoder) decodeString(n int) (interface{}, error) {
	// var err error

	buf := [32]byte{}
	for i := 0; i < n; i++ {
		buf[i] = d.next()
	}
	return string(buf[:n]), nil
}

func (d *Decoder) decodeBinary(n int) (interface{}, error) {
	// var err error

	buf := [32]byte{}
	for i := 0; i < n; i++ {
		buf[i] = d.next()
	}
	return buf[:n], nil
}

func (d *Decoder) decode() (interface{}, error) {
	c := d.next()
	var value interface{}
	var err error

	switch c & 0xf0 {
	case 0x00, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70:
		value = int(c)
	case 0x80:
		value, err = d.decodeMap(int(c & 0x0f))
		if err != nil {
			return value, err
		}
	case 0x90:
		value, err = d.decodeArray(int(c & 0x0f))
		if err != nil {
			return value, err
		}
	case 0xa0, 0xb0:
		value, err = d.decodeString(int(c & 0x1f))
		if err != nil {
			return value, err
		}
	case 0xe0, 0xf0:
		value = -int(c & 0x1f)
	default:
		switch c {
		case 0xc0:
			value = nil
		case 0xc2:
			value = false
		case 0xc3:
			value = true
		case 0xc4:
			ln := int(d.next())
			value, err = d.decodeBinary(ln)
			if err != nil {
				return value, err
			}
		case 0xc5:
			ln := int(d.next()) << 8
			ln += int(d.next())
			value, err = d.decodeBinary(ln)
			if err != nil {
				return value, err
			}
		case 0xc6:
			ln := int(d.next()) << 24
			ln += int(d.next()) << 16
			ln += int(d.next()) << 8
			ln += int(d.next())
			value, err = d.decodeBinary(ln)
			if err != nil {
				return value, err
			}
		case 0xcc:
			n := uint(d.next())
			value = n
		case 0xcd:
			n := uint(d.next()) << 8
			n += uint(d.next())
			value = n
		case 0xce:
			n := uint(d.next()) << 24
			n += uint(d.next()) << 16
			n += uint(d.next()) << 8
			n += uint(d.next())
			value = n
		case 0xcf:
			n := uint(d.next()) << 56
			n += uint(d.next()) << 48
			n += uint(d.next()) << 40
			n += uint(d.next()) << 32
			n += uint(d.next()) << 24
			n += uint(d.next()) << 16
			n += uint(d.next()) << 8
			n += uint(d.next())
			value = n
			// case 0xd0:
			// 	n := uint(d.next())
			// 	return n, nil
			// case 0xd1:
			// 	n := uint(d.next()) << 8
			// 	n += uint(d.next())
			// 	return n, nil
			// case 0xd2:
			// 	n := uint(d.next()) << 24
			// 	n += uint(d.next()) << 16
			// 	n += uint(d.next()) << 8
			// 	n += uint(d.next())
			// 	return n, nil
			// case 0xd3:
			// 	n := uint(d.next()) << 56
			// 	n += uint(d.next()) << 48
			// 	n += uint(d.next()) << 40
			// 	n += uint(d.next()) << 32
			// 	n += uint(d.next()) << 24
			// 	n += uint(d.next()) << 16
			// 	n += uint(d.next()) << 8
			// 	n += uint(d.next())
			// 	return n, nil
		case 0xd9:
			ln := int(d.next())
			value, err = d.decodeString(ln)
			if err != nil {
				return value, err
			}
		case 0xda:
			ln := int(d.next()) << 8
			ln += int(d.next())
			value, err = d.decodeString(ln)
			if err != nil {
				return value, err
			}
		case 0xdb:
			ln := int(d.next()) << 24
			ln += int(d.next()) << 16
			ln += int(d.next()) << 8
			ln += int(d.next())
			value, err = d.decodeString(ln)
			if err != nil {
				return value, nil
			}
		case 0xdc:
			ln := int(d.next()) << 8
			ln += int(d.next())
			value, err = d.decodeArray(ln)
			if err != nil {
				return value, err
			}
		case 0xdd:
			ln := int(d.next()) << 24
			ln += int(d.next()) << 16
			ln += int(d.next()) << 8
			ln += int(d.next())
			value, err = d.decodeArray(ln)
			if err != nil {
				return value, err
			}
		case 0xde:
			ln := int(d.next()) << 8
			ln += int(d.next())
			value, err = d.decodeMap(ln)
			if err != nil {
				return value, err
			}
		case 0xdf:
			ln := int(d.next()) << 24
			ln += int(d.next()) << 16
			ln += int(d.next()) << 8
			ln += int(d.next())
			value, err = d.decodeMap(ln)
			if err != nil {
				return value, err
			}
		default:
			return nil, fmt.Errorf("not implemented %02x", c)
		}
	}

	return value, nil
}

func (d *Decoder) bindObject(value interface{}, ptrList ...interface{}) int {

	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		for ind, ptr := range ptrList {
			intPtr, ok := ptr.(*int)
			if ok {
				*intPtr = int(v.Int())
				return ind
			}
			int8Ptr, ok := ptr.(*int8)
			if ok {
				*int8Ptr = int8(v.Int())
				return ind
			}
			int16Ptr, ok := ptr.(*int16)
			if ok {
				*int16Ptr = int16(v.Int())
				return ind
			}
			int32Ptr, ok := ptr.(*int32)
			if ok {
				*int32Ptr = int32(v.Int())
				return ind
			}
			int64Ptr, ok := ptr.(*int64)
			if ok {
				*int64Ptr = int64(v.Int())
				return ind
			}
			uintPtr, ok := ptr.(*uint)
			if ok {
				*uintPtr = uint(v.Int())
				return ind
			}
			uint8Ptr, ok := ptr.(*uint8)
			if ok {
				*uint8Ptr = uint8(v.Int())
				return ind
			}
			uint16Ptr, ok := ptr.(*uint16)
			if ok {
				*uint16Ptr = uint16(v.Int())
				return ind
			}
			uint32Ptr, ok := ptr.(*uint32)
			if ok {
				*uint32Ptr = uint32(v.Int())
				return ind
			}
			uint64Ptr, ok := ptr.(*uint64)
			if ok {
				*uint64Ptr = uint64(v.Int())
				return ind
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		for ind, ptr := range ptrList {
			intPtr, ok := ptr.(*int)
			if ok {
				*intPtr = int(v.Uint())
				return ind
			}
			int8Ptr, ok := ptr.(*int8)
			if ok {
				*int8Ptr = int8(v.Uint())
				return ind
			}
			int16Ptr, ok := ptr.(*int16)
			if ok {
				*int16Ptr = int16(v.Uint())
				return ind
			}
			int32Ptr, ok := ptr.(*int32)
			if ok {
				*int32Ptr = int32(v.Uint())
				return ind
			}
			int64Ptr, ok := ptr.(*int64)
			if ok {
				*int64Ptr = int64(v.Uint())
				return ind
			}
			uintPtr, ok := ptr.(*uint)
			if ok {
				*uintPtr = uint(v.Uint())
				return ind
			}
			uint8Ptr, ok := ptr.(*uint8)
			if ok {
				*uint8Ptr = uint8(v.Uint())
				return ind
			}
			uint16Ptr, ok := ptr.(*uint16)
			if ok {
				*uint16Ptr = uint16(v.Uint())
				return ind
			}
			uint32Ptr, ok := ptr.(*uint32)
			if ok {
				*uint32Ptr = uint32(v.Uint())
				return ind
			}
			uint64Ptr, ok := ptr.(*uint64)
			if ok {
				*uint64Ptr = uint64(v.Uint())
				return ind
			}
		}
	case reflect.String:
		for ind, ptr := range ptrList {
			strPtr, ok := ptr.(*string)
			if ok {
				*strPtr = v.String()
				return ind
			}
		}
	case reflect.Slice:
		for ind, ptr := range ptrList {
			v2 := reflect.ValueOf(ptr)
			if v2.Kind() != reflect.Ptr {
				continue
			}
			if v2.Elem().Kind() != reflect.Slice && v2.Elem().Kind() != reflect.Array {
				continue
			}
			if v.Len() == 0 {
				v2.Elem().SetLen(0)
				return ind
			}

			if v2.Elem().Type().Elem() != v.Type().Elem() {
				continue
			}

			v2.Elem().Set(v)
			return ind
		}
	case reflect.Map:
		for ind, ptr := range ptrList {
			v2 := reflect.ValueOf(ptr)
			if v2.Kind() != reflect.Ptr {
				continue
			}
			if v2.Elem().Kind() == reflect.Map {
				if v2.Elem().Type().Elem() != v.Type().Elem() || v2.Elem().Type().Key() != v.Type().Key() {
					continue
				}
				v2.Elem().Set(v)
				return ind
			}
			if v2.Elem().Kind() != reflect.Struct {
				continue
			}
			if v2.Elem().NumField() != v.Len() {
				continue
			}
			isFailed := false
			for i := 0; i < v2.Elem().NumField(); i++ {
				field := v2.Elem().Type().Field(i)
				var name string
				if field.Tag.Get("msgpack") != "" {
					name = field.Tag.Get("msgpack")
				} else {
					name = field.Name
				}

				mapIndex := v.MapIndex(reflect.ValueOf(name))
				if !mapIndex.IsValid() {
					isFailed = true
					break
				}

				fv := v2.Elem().Field(i)
				v3 := reflect.New(fv.Type())

				childInd := d.bindObject(mapIndex.Interface(), v3.Interface())
				if childInd == -1 {
					isFailed = true
					break
				}
				fv.Set(v3.Elem())
			}
			if isFailed {
				v2.Elem().Set(reflect.New(v2.Elem().Type()).Elem())
				continue
			}
			return ind
		}
	}

	return -1
}

func (d *Decoder) Decode(ptrList ...interface{}) (interface{}, int, error) {
	value, err := d.decode()
	if err != nil {
		return value, -1, err
	}
	ind := d.bindObject(value, ptrList...)
	return value, ind, nil
}
