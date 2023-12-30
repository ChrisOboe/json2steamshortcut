package vdf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const (
	// IdFieldNameMagicV1 is used to identify the ID field, as the
	// field is unnamed in the V1 format.
	IdFieldNameMagicV1 = "reserved_id_v1"

	idFieldMaxLength = 10
)

type Object interface {
	Fields() []Field
	Append(field Field)
}

type defaultObject struct {
	fields []Field
}

func (o *defaultObject) Fields() []Field {
	return o.fields
}

func (o *defaultObject) Append(field Field) {
	o.fields = append(o.fields, field)
}

type RawObjectParser interface {
	Parse() (Object, error)
}

type v1ObjectParser struct {
	gotId  bool
	raw    string
	result Object
}

func (o *v1ObjectParser) Parse() (Object, error) {
	if len(o.raw) == 0 {
		return o.result, nil
	}

	if !o.gotId {
		err := o.parseId()
		if err != nil {
			return o.result, err
		}
	}

	currentValueType, err := o.parseCurrentValueType()
	if err != nil {
		return o.result, err
	}

	currentFieldName, isEnd, err := o.parseFieldName()
	if err != nil {
		return o.result, err
	}

	if isEnd {
		return o.result, nil
	}

	value, err := o.value(currentValueType)
	if err != nil {
		return o.result, err
	}

	f := &defaultField{
		name:      currentFieldName,
		valueType: currentValueType,
	}

	switch currentValueType {
	case stringValue:
		f.stringValue = value
	case int32Value:
		f.int32Value = parseRawInt32Value(value)
	case sliceValue:
		f.sliceValue = parseSlice(value)
	}

	o.result.Append(f)

	return o.Parse()
}

func (o *v1ObjectParser) parseId() error {
	if string(o.raw[0]) == null {
		o.get(1, "")
	}

	// Find the ID integer.
	numBytes := 0
	for i := range o.raw {
		if numBytes > idFieldMaxLength {
			return errors.New("ID field is grater than maximum size of " + strconv.Itoa(idFieldMaxLength))
		} else if unicode.IsDigit(rune(o.raw[i])) {
			numBytes = numBytes + 1
		} else {
			break
		}
	}

	if numBytes == 0 {
		return errors.New("Failed to locate a valid ID string")
	}

	// Drop the ID + null.
	idRaw, ok := o.get(numBytes + 1, "")
	if !ok {
		return errors.New("Failed to cut ID field - index out of range")
	}

	id, err := strconv.Atoi(idRaw[0:len(idRaw) - 1])
	if err != nil {
		return errors.New("Failed to parse ID - " + err.Error())
	}

	o.result.Append(&defaultField{
		name:      IdFieldNameMagicV1,
		valueType: idValue,
		idValue:   id,
	})

	o.gotId = true

	return nil
}

func (o *v1ObjectParser) parseCurrentValueType() (FieldValueType, error) {
	value, ok := o.get(1, "")
	if !ok {
		return stringValue, errors.New("Failed to read type field - no bytes remaining")
	}

	var currentValueType FieldValueType

	switch string(value[0]) {
	case sliceField:
		currentValueType = sliceValue
	case intField:
		currentValueType = int32Value
	case stringField:
		currentValueType = stringValue
	default:
		return stringValue, fmt.Errorf("%s, %x", "Invalid field type", value)
	}

	return currentValueType, nil
}

func (o *v1ObjectParser) parseFieldName() (name string, isEof bool, err error) {
	numToGet := strings.Index(o.raw, null)

	if numToGet < 0 {
		// No null terminator is present.
		numToGet = len(o.raw)
	} else {
		// Include the null terminator as well.
		numToGet = numToGet + 1
	}

	// Drop the field name and the null terminator.
	v, ok := o.get(numToGet, null)
	if !ok {
		return "", false, errors.New("Field name is out of range")
	}

	if len(v) == 0 {
		return "", false, errors.New("Field name is empty")
	}

	if !unicode.IsLetter(rune(v[0])) {
		return "", false, errors.New("Field name does not start with a letter")
	}

	return v, false, nil
}

func (o *v1ObjectParser) value(current FieldValueType) (string, error) {
	var numToGet int
	var trim string

	switch current {
	case stringValue:
		numToGet = strings.Index(o.raw, null) + 1
		trim = null
	case int32Value:
		numToGet = 4
	case sliceValue:
		// TODO: Jank - but this file format is awful. Halp.
		if strings.HasSuffix(o.raw, null) {
			numToGet = strings.LastIndex(o.raw, null) + 1
			trim = null
		} else {
			numToGet = len(o.raw)
		}
	default:
		return "", errors.New("Unknown field type - " + strconv.Itoa(int(current)))
	}

	value, ok := o.get(numToGet, trim)
	if !ok {
		return "", errors.New("Failed to read value field")
	}

	return value, nil
}

func (o *v1ObjectParser) get(numberOfBytes int, trim string) (string, bool) {
	if isIndexOutsideString(numberOfBytes - 1, o.raw) {
		return "", false
	}

	value := o.raw[0:numberOfBytes]

	o.raw = o.raw[numberOfBytes:]

	if len(trim) > 0 {
		value = strings.TrimSuffix(value, trim)
	}

	return value, true
}

func NewEmptyObject() Object {
	return &defaultObject{}
}

func NewObject(fields []Field) Object {
	return &defaultObject{
		fields: fields,
	}
}

func ParseRawObject(rawData string, version FormatVersion) (Object, error) {
	p, err := NewRawObjectParser(rawData, version)
	if err != nil {
		return &defaultObject{}, err
	}

	return p.Parse()
}

func NewRawObjectParser(rawData string, version FormatVersion) (RawObjectParser, error) {
	switch version {
	case V1:
		return &v1ObjectParser{
			raw:    rawData,
			result: &defaultObject{},
		}, nil
	}

	return &v1ObjectParser{}, errors.New("Format version " + string(version) + " is not supported")
}

func parseRawInt32Value(raw string) int32 {
	var i int32

	err := binary.Read(strings.NewReader(raw), binary.LittleEndian, &i)
	if err != nil {
		return 0
	}

	return i
}

func parseSlice(raw string) []string {
	var values []string

	raw = strings.TrimPrefix(raw, soh)

	for i, s := range strings.Split(raw, null + soh) {
		id, v, wasParsed := parseSliceField(s)
		if i != id {
			return values
		}

		if wasParsed {
			values = append(values, v)
		}
	}

	return values
}

func parseSliceField(raw string) (int, string, bool) {
	values := strings.Split(raw, null)

	if len(values) < 2 {
		return 0, "", false
	}

	i, err := strconv.Atoi(values[0])
	if err != nil {
		return 0, "", false
	}

	return i, values[1], true
}

func isIndexOutsideString(index int, s string) bool {
	totalIndexes := len(s) - 1

	if totalIndexes - index < 0 {
		return true
	}

	return false
}
