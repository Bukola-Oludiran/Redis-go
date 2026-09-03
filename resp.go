package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = "+"
	ERROR   = "-"
	INTEGER = ":"
	BULK    = "$"
	ARRAY   = "*"
)

type Value struct {
	typ   string
	str   string
	num   int
	bulk  string
	array []Value
}

type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd io.Reader) *Resp {

	return &Resp{reader: bufio.NewReader(rd)}
}

func (r *Resp) readLine() (line []byte, err error, n int) {
	for {

		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, err, 0

		}
		n += 1
		line = append(line, b)
		if len(line) >= 2 && line[len(line)-2] == '\r' {
			break
		}
	}

	return line[:len(line)-2], nil, n

}

func (r *Resp) readInteger() (x int, err error, n int) {

	line, err, n := r.readLine()
	if err != nil {
		return 0, err, 0
	}

	i64, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, err, n

	}
	return int(i64), nil, n

}

func (r *Resp) Read() (Value, error) {

	_type, err := r.reader.ReadByte()

	if err != nil {
		return Value{}, err
	}

	switch string(_type) {
	case BULK:
		return r.readBulk()
	case ARRAY:
		return r.readArray()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil

	}

}

func (r *Resp) readArray() (Value, error) {

	v := Value{}
	v.typ = "array"

	length, err, _ := r.readInteger()
	if err != nil {
		return v, err
	}

	v.array = make([]Value, length)

	for i := 0; i < length; i++ {
		val, err := r.Read()
		if err != nil {
			return v, err
		}
		v.array[i] = val
	}

	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{}
	v.typ = "bulk"

	length, err, _ := r.readInteger()
	if err != nil {
		return v, err
	}

	bulk := make([]byte, length)

	r.reader.Read(bulk)

	v.bulk = string(bulk)

	r.readLine()

	return v, nil
}
