package codec

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/lukasjarosch/skipper/data"
	"gopkg.in/yaml.v3"
)

type YamlCodec struct{}

func NewYamlCodec() YamlCodec {
	return YamlCodec{}
}

func (codec YamlCodec) Unmarshal(in []byte) (data.Map, error) {
	var out data.Map
	err := codec.UnmarshalTarget(in, &out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (codec YamlCodec) UnmarshalTarget(in []byte, target interface{}) error {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		return fmt.Errorf("cannot decode with path: target must be a pointer")
	}

	dec := yaml.NewDecoder(bytes.NewReader(in))
	dec.KnownFields(true)
	err := dec.Decode(target)
	if err != nil {
		return err
	}

	return nil
}

// DecodeBytesWithPath attempts to resolve the given path within the bytes and
// decodes them into the given target interface.
// The target interface must be a pointer.
func (codec YamlCodec) UnmarshalPath(in []byte, path data.Path, target interface{}) error {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		return fmt.Errorf("cannot decode with path: target must be a pointer")
	}

	var out data.Map
	err := yaml.Unmarshal(in, &out)
	if err != nil {
		return err
	}

	tree, err := out.Get(path)
	if err != nil {
		return err
	}

	b, err := codec.Marshal(tree)
	if err != nil {
		return err
	}

	err = codec.UnmarshalTarget(b, target)
	if err != nil {
		return err
	}
	return nil
}

func (codec YamlCodec) Marshal(in interface{}) ([]byte, error) {
	out, err := yaml.Marshal(in)
	if err != nil {
		return nil, err
	}
	return out, nil
}
