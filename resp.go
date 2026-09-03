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

func (v Value) Marshal() []byte {

	switch v.typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "null":
		return v.marshalNull()
	case "error":
		return v.marshalError()
	default:
		return []byte{}
	}

}

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, []byte(STRING)...)
	bytes = append(bytes, []byte(v.str)...)
	bytes = append(bytes, []byte("\r\n")...)

	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, []byte(BULK)...)
	bytes = append(bytes, []byte(strconv.Itoa(len(v.bulk)))...)
	bytes = append(bytes, []byte("\r\n")...)
	bytes = append(bytes, []byte(v.bulk)...)
	bytes = append(bytes, []byte("\r\n")...)
	return bytes
}

func (v Value) marshalArray() []byte {
	len := len(v.array)

	var bytes []byte
	bytes = append(bytes, []byte(ARRAY)...)
	bytes = append(bytes, []byte(strconv.Itoa(len))...)
	bytes = append(bytes, []byte("\r\n")...)

	for _, value := range v.array {
		bytes = append(bytes, value.Marshal()...)
	}
	return bytes

}

func (v Value) marshalNull() []byte {
	return []byte("$-1\r\n")
}

func (v Value) marshalError() []byte {

	var bytes []byte
	bytes = append(bytes, []byte(ERROR)...)
	bytes = append(bytes, []byte(v.str)...)
	bytes = append(bytes, []byte("\r\n")...)

	return bytes
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}

}

func (w *Writer) Write(v Value) error {

	b := v.Marshal()
	_, err := w.writer.Write(b)
	if err != nil {
		return err
	}
	return nil
}
