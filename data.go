package skipper

import (
	"log"
	"strings"
)

type Data map[string]interface{}

const IdentifierSeparator = "."

func (d Data) HasKey(k string) bool {
	if _, ok := d[k]; ok {
		return true
	}
	return false
}

func (d Data) Get(k string) Data {
	return d[k].(Data)
}

// HasValueAtIdentifier returns true if the given identifier path points to a value.
func (d Data) HasValueAtIdentifier(path string) bool {
	if d.GetByIdentifier(path) == nil {
		return false
	}
	return true
}

// GetByIdentifier returns a value given a dot-separated identifier.
// TODO: add support for array indexing (e.g. 'foo.bar.3.baz')
func (d Data) GetByIdentifier(path string) interface{} {
	var segments = strings.Split(path, IdentifierSeparator)

	obj := d

	for i, v := range segments {

		// we found the last segment of the identifier path
		if i == len(segments)-1 {
			return obj[v]
		}

		switch obj[v].(type) {
		case Data:
			obj = Data(obj[v].(Data))
		default:
			return nil
		}
	}

	return obj
}

func (d *Data) SetByIdentifier(path string, value interface{}) {
	var segments = strings.Split(path, IdentifierSeparator)

	obj := (*d)

	for i, v := range segments {

		// we found the last segment of the identifier path
		if i == len(segments)-1 {
			log.Println("REPLACE HERE", obj[v])
			obj[v] = value
			(*d) = obj
			return
		}

		switch obj[v].(type) {
		case Data:
			obj = Data(obj[v].(Data))
		}
	}
}

// MergeReplace merges the existing Data with the given.
// If a key already exists, the passed data has precedence and it's value will be used.
func (d Data) MergeReplace(data Data) Data {
	out := make(Data, len(d))
	for k, v := range d {
		out[k] = v
	}
	for k, v := range data {
		if v, ok := v.(Data); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(Data); ok {
					out[k] = bv.MergeReplace(v)
					continue
				}
			}
		}
		if v, ok := v.([]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.([]interface{}); ok {
					out[k] = append(bv, v...)
					continue
				}
			}
		}
		out[k] = v
	}

	return out
}
