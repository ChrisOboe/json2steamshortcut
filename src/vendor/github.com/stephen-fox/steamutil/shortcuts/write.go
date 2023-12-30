package shortcuts

import (
	"errors"
	"io"
	"os"

	"github.com/stephen-fox/steamutil/vdf"
)

const (
	defaultFileMode = 0644
)

const (
	Unchanged      UpdateResult = "No changes were made to the file"
	CreatedNewFile UpdateResult = "Created new file"
	UpdatedEntry   UpdateResult = "Updated existing entry in the file"
	AddedNewEntry  UpdateResult = "Added new entry to the file"
)

// UpdateResult describes the result of creating or updating a shortcuts file.
type UpdateResult string

// CreateOrUpdateConfig provides configuration data for creating or updating
// a shortcuts file.
type CreateOrUpdateConfig struct {
	// Path is the path to the shortcuts file.
	Path string

	// Mode is the mode to set the file to if a new file is created.
	// This defaults to defaultFileMode if not specified.
	Mode os.FileMode

	// MatchName is the application name to match.
	MatchName string

	// OnMatch is the function to execute when a shortcut meets the
	// match criteria.
	OnMatch func(name string, match *Shortcut)

	// NoMatch is the function to execute when no shortcut is found
	// for the provided match criteria.
	NoMatch func(name string) (s Shortcut, doNothing bool)
}

// IsValid returns a non-nil error if the configuration is invalid.
func (o *CreateOrUpdateConfig) IsValid() error {
	if len(o.MatchName) == 0 {
		return errors.New("the shortcut name to match cannot be empty")
	}

	if len(o.Path) == 0 {
		return errors.New("the shortcut file path cannot be empty")
	}

	if o.OnMatch == nil {
		return errors.New("the shortcut match function cannot be nil")
	}

	if o.NoMatch == nil {
		return errors.New("the shortcut not matched function cannot be nil")
	}

	if o.Mode == 0 {
		o.Mode = defaultFileMode
	}

	return nil
}

// CreateOrUpdateFile creates or updates a shortcuts file.
func CreateOrUpdateFile(config CreateOrUpdateConfig) (UpdateResult, error) {
	return CreateOrUpdateVdfV1File(config)
}

// CreateOrUpdateFile creates or updates a shortcuts file using the
// VDF v1 format.
func CreateOrUpdateVdfV1File(config CreateOrUpdateConfig) (UpdateResult, error) {
	err := config.IsValid()
	if err != nil {
		return Unchanged, err
	}

	alreadyExists := false

	_, statErr := os.Stat(config.Path)
	if statErr == nil {
		alreadyExists = true
	}

	f, err := os.OpenFile(config.Path, os.O_RDWR|os.O_CREATE, config.Mode)
	if err != nil {
		return Unchanged, err
	}
	defer f.Close()

	var currentScs []Shortcut

	if alreadyExists {
		currentScs, err = ReadVdfV1File(f)
		if err != nil {
			return Unchanged, err
		}
	}

	result := Unchanged

	for i := range currentScs {
		if currentScs[i].AppName == config.MatchName {
			result = UpdatedEntry

			config.OnMatch(config.MatchName, &currentScs[i])

			break
		}
	}

	if result == Unchanged {
		s, doNothing := config.NoMatch(config.MatchName)
		if !doNothing {
			s.Id = len(currentScs)
			currentScs = append(currentScs, s)
			result = AddedNewEntry
		}
	}

	err = OverwriteVdfV1File(f, currentScs)
	if err != nil {
		return Unchanged, err
	}

	if alreadyExists {
		return result, nil
	}

	return CreatedNewFile, nil
}

// OverwriteFile overwrites the shortcuts file using the provided data in
// the VDF v1 format.
func OverwriteVdfV1File(f *os.File, scs []Shortcut) error {
	_, err := f.Seek(0, 0)
	if err != nil {
		return err
	}

	err = f.Truncate(0)
	if err != nil {
		return err
	}

	err = WriteVdfV1(scs, f)
	if err != nil {
		return err
	}

	return nil
}

// WriteVdfV1 writes the provides shortcuts data to the specified io.Writer
// in the VDF v1 format.
func WriteVdfV1(shortcuts []Shortcut, w io.Writer) error {
	config := vdf.Config{
		Name:    header,
		Version: vdf.V1,
	}

	c, err := vdf.NewConstructor(config)
	if err != nil {
		return err
	}

	var objects []vdf.Object

	for _, s := range shortcuts {
		objects = append(objects, s.object())
	}

	v := c.Build(objects)

	s, err := v.String()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(s))
	if err != nil {
		return err
	}

	return nil
}
