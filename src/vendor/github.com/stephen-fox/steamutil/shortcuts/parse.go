package shortcuts

import (
	"io"
	"os"
	"strings"

	"github.com/stephen-fox/steamutil/vdf"
)

// ReadFile reads the provided shortcuts data from an os.File.
func ReadFile(f *os.File) ([]Shortcut, error) {
	return ReadVdfV1File(f)
}

// ReadVdfV1File reads the provided shortcuts data from an os.File using the
// VDF v1 format.
func ReadVdfV1File(f *os.File) ([]Shortcut, error) {
	scs, err := ReadVdfV1(f)
	if err != nil {
		return []Shortcut{}, err
	}

	_, err = f.Seek(0, 0)
	if err != nil {
		return []Shortcut{}, err
	}

	return scs, nil
}

// Read reads the provided shortcuts data from an io.Reader.
func Read(r io.Reader) ([]Shortcut, error) {
	return ReadVdfV1(r)
}

// Read reads the provided shortcuts data from an io.Reader using the
// VDF v1 format.
func ReadVdfV1(r io.Reader) ([]Shortcut, error) {
	config := vdf.Config{
		Name:    header,
		Version: vdf.V1,
	}

	c, err := vdf.NewConstructor(config)
	if err != nil {
		return []Shortcut{}, err
	}

	v, err := c.Read(r)
	if err != nil {
		return []Shortcut{}, err
	}

	var scs []Shortcut

	for _, o := range v.Objects() {
		scs = append(scs, objectToShortcut(o))
	}

	return scs, nil
}

func objectToShortcut(object vdf.Object) Shortcut {
	var s Shortcut
	
	for _, f := range object.Fields() {
		switch f.Name() {
		case vdf.IdFieldNameMagicV1:
			s.Id = f.IdValue()
		case appNameField:
			s.AppName = f.StringValue()
		case exePathField:
			s.ExePath = trimDoubleQuote(f.StringValue())
		case startDirField:
			s.StartDir = trimDoubleQuote(f.StringValue())
		case iconPathField:
			s.IconPath = f.StringValue()
		case shortcutPathField:
			s.ShortcutPath = f.StringValue()
		case launchOptionsField:
			s.LaunchOptions = f.StringValue()
		case isHiddenField:
			s.IsHidden = f.BoolValue()
		case allowDesktopConfigField:
			s.AllowDesktopConfig = f.BoolValue()
		case allowOverlayField:
			s.AllowOverlay = f.BoolValue()
		case isOpenVrField:
			s.IsOpenVr = f.BoolValue()
		case lastPlayTimeField:
			s.LastPlayTimeEpoch = f.Int32Value()
		case tagsField:
			s.Tags = f.SliceValue()
		}
	}

	return s
}

func trimDoubleQuote(s string) string {
	return strings.TrimPrefix(strings.TrimSuffix(s, "\""), "\"")
}
