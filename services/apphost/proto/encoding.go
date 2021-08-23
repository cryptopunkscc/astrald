package proto

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

func writeBuffer(writer io.Writer, buf []byte) error {
	var l = uint16(len(buf))

	if err := binary.Write(writer, binary.BigEndian, &l); err != nil {
		return err
	}

	_, err := writer.Write(buf)

	return err
}

func readBuffer(reader io.Reader) ([]byte, error) {
	var l uint16

	if err := binary.Read(reader, binary.BigEndian, &l); err != nil {
		return nil, err
	}

	var buf = make([]byte, l)

	_, err := io.ReadFull(reader, buf)
	return buf, err
}

func writeJSON(writer io.Writer, v interface{}) error {
	jsonBytes, err := json.Marshal(v)
	if err != nil {
		return err
	}

	return writeBuffer(writer, jsonBytes)
}

func readJSON(reader io.Reader, v interface{}) error {
	buf, err := readBuffer(reader)
	if err != nil {
		return err
	}

	return json.Unmarshal(buf, v)
}
