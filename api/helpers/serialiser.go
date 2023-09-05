package helpers

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

type SerializerFormat = string

const (
	JSON SerializerFormat = "json"
	YAML                  = "yaml"
)

type Serializer interface {
	Serialize(any) ([]byte, error)
}

type YamlSerializer struct{}

func (y *YamlSerializer) Serialize(a any) ([]byte, error) {
	if a == nil {
		return []byte(""), nil
	}
	return yaml.Marshal(a)
}

type JsonSerializer struct{}

func (j *JsonSerializer) Serialize(a any) ([]byte, error) {
	if a == nil {
		return []byte(""), nil
	}
	return json.MarshalIndent(a, " ", " ")
}
