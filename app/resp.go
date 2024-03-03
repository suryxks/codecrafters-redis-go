package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Resp struct {
	reader *bufio.Reader
}
type Value struct {
	typ    string
	str    string
	num    int
	bulk   string
	array  []Value
	length int
}

func (v *Value) Marshal() []byte {
	switch v.typ {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()

	case "string":
		return v.marshalString()

	default:
		return []byte{}

	}
}
func (v *Value) marshalArray() []byte {
	len := len(v.array)
	var bytes []byte
	bytes = append(bytes, []byte(ARRAY)...)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < len; i++ {
		bytes = append(bytes, v.array[i].Marshal()...)
	}

	return bytes
}
func (v *Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, []byte(STRING)...)
	bytes = append(bytes, strconv.Itoa(len(v.str))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, []byte(v.str)...)
	bytes = append(bytes, '\r', '\n')
	return bytes
}
func (v *Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, []byte(BULK)...)
	bytes = append(bytes, strconv.Itoa(len(v.bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}
func (r *Resp) readLine() (line []byte, n int, err error) {

	for {
		b, err := r.reader.ReadByte()
		if err != nil {
			return nil, 0, err
		}
		n += 1
		line = append(line, b)

		if len(line) >= 2 && strings.Contains(string(line), string("\r\n")) {
			break
		}
	}

	return line[:len(line)-2], n, nil
}
func (r *Resp) readInteger() (x int, n int, err error) {
	line, n, err := r.readLine()
	if err != nil {
		return 0, 0, err
	}
	i64, err := strconv.ParseInt(string(line[0]), 10, 64)
	if err != nil {
		return 0, n, err
	}

	return int(i64), n, nil
}

func (r *Resp) Read() (Value, error) {
	_type, err := r.reader.ReadByte()

	if err != nil {
		return Value{}, err
	}

	switch string(_type) {
	case ARRAY:
		return r.readArray()
	case BULK:
		return r.readBulk()
	case STRING:
		return r.readString()
	default:
		fmt.Printf("Unknown type: %v", string(_type))
		return Value{}, nil
	}

}
func (r *Resp) readArray() (Value, error) {
	v := Value{}
	v.typ = "array"

	length, err := r.reader.ReadByte()

	if err != nil {
		return v, err
	}
	lengthInt, err := parseInt(length)
	if err != nil {
		return v, err
	}
	r.reader.ReadByte()
	r.reader.ReadByte()

	for i := 0; i < lengthInt; i++ {

		value, err := r.Read()
		if err != nil {
			return v, err
		}
		v.array = append(v.array, value)
	}

	return v, nil
}

func (r *Resp) readBulk() (Value, error) {
	v := Value{}
	v.typ = "bulk"

	length, err := r.reader.ReadByte()
	if err != nil {
		return v, err
	}
	r.reader.ReadByte()
	r.reader.ReadByte()

	line, _, err := r.readLine()

	if err != nil {
		return v, err
	}
	v.bulk = string(line)
	lengthInt, err := parseInt(length)
	if err != nil {
		return v, err
	}
	v.length = lengthInt

	return v, nil

}
func (r *Resp) readString() (Value, error) {
	line, _, err := r.reader.ReadLine()
	if err != nil {
		return Value{}, err
	}
	return Value{typ: "string", str: string(line)}, nil
}

const (
	STRING  string = "+"
	ARRAY   string = "*"
	ERROR   string = "-"
	INTEGER string = ":"
	BULK    string = "$"
)

func NewResp(rd io.Reader) *Resp {
	return &Resp{bufio.NewReader(rd)}
}

type Writer struct {
	writer io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{writer: w}

}
func (w *Writer) Write(v Value) error {
	var bytes = v.Marshal()

	_, err := w.writer.Write(bytes)
	if err != nil {
		return err
	}

	return nil
}
func parseInt(length byte) (number int, err error) {
	i64, err := strconv.ParseInt(string(length), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(i64), nil
}
