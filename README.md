go-msgpack
==========

This is msgpack implement by Golang.

example.
--------

```go
package main

import (
	"bytes"
	"fmt"
	"github.com/erukiti/go-msgpack"
	"github.com/erukiti/go-util"
)

type Fuga struct {
	Piyo string `msgpack:"piyo"`
}

type Hoge struct {
	Fuga Fuga `msgpack:"fuga"`
}

func main() {
	buf := []byte{0x81, 0xa4, 'f', 'u', 'g', 'a', 0x81, 0xa4, 'p', 'i', 'y', 'o', 0xa4, 'p', 'i', 'y', 'o'}
	r := bytes.NewBuffer(buf)
	decoder := msgpack.NewDecoder(r)

	var v0 map[string]string
	var v1 int
	var v2 Hoge

	value, ind, _ := decoder.Decode(&v0, &v1, &v2)
	fmt.Printf("result index: %d\n", ind)
	switch ind {
	case -1:
		util.Dump(value)
	case 0:
		util.Dump(v0)
	case 1:
		util.Dump(v1)
	case 2:
		util.Dump(v2)
	}
}
```

