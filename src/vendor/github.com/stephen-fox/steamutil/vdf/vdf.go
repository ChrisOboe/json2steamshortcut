package vdf

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"strings"
)

const (
	// null is the null ASCII character.
	null = "\x00"
	// soh is the 'Start of Header' ASCII character.
	soh = "\x01"
	// stx is the 'Start of Text' ASCII character.
	stx = "\x02"

	stringField = soh
	intField    = stx
	sliceField  = null
)

const (
	V1 FormatVersion = 1
)

type FormatVersion int

type Vdf interface {
	Append(object Object)
	Objects() []Object
	String() (string, error)
}

type defaultVdf struct {
	objects []Object
	config  Config
}

func (o *defaultVdf) Append(object Object) {
	o.objects = append(o.objects, object)
}

func (o *defaultVdf) Objects() []Object {
	return o.objects
}

func (o *defaultVdf) String() (string, error) {
	sb := &strings.Builder{}

	switch o.config.Version {
	case V1:
		break
	default:
		return "", errors.New("The specified format is not supported - " + string(o.config.Version))
	}

	sb.Write(o.config.header)

	for _, object := range o.objects {
		for _, field := range object.Fields() {
			err := field.Append(sb, o.config.Version)
			if err != nil {
				return "", err
			}
		}

		sb.Write(o.config.objectDelim)
	}

	sb.Write(o.config.footer)

	return sb.String(), nil
}

type Config struct {
	Name        string
	Version     FormatVersion
	header      []byte
	footer      []byte
	objectDelim []byte
}

type Constructor interface {
	Build(objects []Object) Vdf
	Read(r io.Reader) (Vdf, error)
}

type v1Constructor struct {
	config Config
}

func (o *v1Constructor) Build(objects []Object) Vdf {
	return &defaultVdf{
		objects: objects,
		config:  o.config,
	}
}

func (o *v1Constructor) Read(r io.Reader) (Vdf, error) {
	result := &defaultVdf{
		config: o.config,
	}

	s := bufio.NewScanner(r)
	s.Split(o.split)

	for s.Scan() {
		text := s.Text()

		if len(strings.TrimSpace(text)) > 0 {
			object, err := ParseRawObject(text, result.config.Version)
			if err != nil {
				return result, err
			}

			result.Append(object)
		}
	}

	err := s.Err()
	if err != nil {
		return result, err
	}

	return result, nil
}

func (o *v1Constructor) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.Index(data, o.config.header); i >= 0 {
		return i + len(o.config.header), nil, nil
	}

	if i := bytes.Index(data, o.config.objectDelim); i >= 0 {
		return i + len(o.config.objectDelim), data[0:i], nil
	}

	if i := bytes.Index(data, o.config.footer); i >= 0 {
		return i + len(o.config.footer), data[0:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

func NewConstructor(config Config) (Constructor, error) {
	if len(config.Name) == 0 {
		return &v1Constructor{}, errors.New("Please specify a name for the vdf")
	}

	switch config.Version {
	case V1:
		config.header = fileHeaderV1(config.Name)
		config.objectDelim = []byte{0, 8, 8}
		config.footer = []byte{8, 8}

		return &v1Constructor{
			config: config,
		}, nil
	}

	return &v1Constructor{}, errors.New("The specified format is not supported - '" + string(config.Version) + "'")
}

func fileHeaderV1(name string) []byte {
	header := []byte{0}
	header = append(header, []byte(name)...)
	header = append(header,0)

	return header
}
