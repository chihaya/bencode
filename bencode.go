// Copyright 2014 The Chihaya Authors. All rights reserved.
// Use of this source code is governed by the BSD 2-Clause license,
// which can be found in the LICENSE file.

// Package bencode implements bencoding of data as defined in BEP 3 using
// type assertion over reflection for performance.
package bencode

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"
)

// Dict represents a bencode dictionary.
type Dict map[string]interface{}

// NewDict allocates the memory for a Dict.
func NewDict() Dict {
	return make(Dict)
}

// An Encoder writes Bencoded objects to an output stream.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the bencoding of v to the stream.
func (enc *Encoder) Encode(v interface{}) error {
	return marshal(enc.w, v)
}

// Marshal returns the bencoding of v.
func Marshal(v interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := marshal(buf, v)
	return buf.Bytes(), err
}

// Marshaler is the interface implemented by objects that can marshal
// themselves.
type Marshaler interface {
	MarshalBencode() ([]byte, error)
}

// marshal writes types bencoded to an io.Writer
func marshal(w io.Writer, data interface{}) error {
	switch v := data.(type) {
	case Marshaler:
		bencoded, err := v.MarshalBencode()
		if err != nil {
			return err
		}
		_, err = w.Write(bencoded)
		if err != nil {
			return err
		}

	case string:
		marshalString(w, v)

	case int:
		marshalInt(w, int64(v))

	case uint:
		marshalUint(w, uint64(v))

	case int64:
		marshalInt(w, v)

	case uint64:
		marshalUint(w, v)

	case []byte:
		w.Write(v)

	case time.Duration: // Assume seconds
		marshalInt(w, int64(v/time.Second))

	case Dict:
		w.Write([]byte{'d'})
		for key, val := range v {
			marshalString(w, key)
			err := marshal(w, val)
			if err != nil {
				return err
			}
		}
		w.Write([]byte{'e'})

	case []Dict:
		w.Write([]byte{'l'})
		for _, val := range v {
			err := marshal(w, val)
			if err != nil {
				return err
			}
		}
		w.Write([]byte{'e'})

	case []string:
		w.Write([]byte{'l'})
		for _, val := range v {
			err := marshal(w, val)
			if err != nil {
				return err
			}
		}
		w.Write([]byte{'e'})

	default:
		// Although not currently necessary,
		// should handle []interface{} manually; Go can't do it implicitly
		return fmt.Errorf("attempted to marshal unsupported type:\n%t", v)
	}

	return nil
}

func marshalInt(w io.Writer, v int64) {
	w.Write([]byte{'i'})
	w.Write([]byte(strconv.FormatInt(v, 10)))
	w.Write([]byte{'e'})
}

func marshalUint(w io.Writer, v uint64) {
	w.Write([]byte{'i'})
	w.Write([]byte(strconv.FormatUint(v, 10)))
	w.Write([]byte{'e'})
}

func marshalString(w io.Writer, v string) {
	w.Write([]byte(strconv.Itoa(len(v))))
	w.Write([]byte{':'})
	w.Write([]byte(v))
}
