package shortcuts

import (
	"strings"

	"github.com/stephen-fox/steamutil/vdf"
)

const (
	header = "shortcuts"

	appNameField            = "AppName"
	exePathField            = "Exe"
	startDirField           = "StartDir"
	iconPathField           = "icon"
	shortcutPathField       = "ShortcutPath"
	launchOptionsField      = "LaunchOptions"
	isHiddenField           = "IsHidden"
	allowDesktopConfigField = "AllowDesktopConfig"
	allowOverlayField       = "AllowOverlay"
	isOpenVrField           = "OpenVR"
	lastPlayTimeField       = "LastPlayTime"
	tagsField               = "tags"
)

// Shortcut represents a single shortcut data structure.
type Shortcut struct {
	Id                 int
	AppName            string
	ExePath            string
	StartDir           string
	IconPath           string
	ShortcutPath       string
	LaunchOptions      string
	IsHidden           bool
	AllowDesktopConfig bool
	AllowOverlay       bool
	IsOpenVr           bool
	LastPlayTimeEpoch  int32
	Tags               []string
}

// Equals returns true if the Shortcut is the same as another.
func (o *Shortcut) Equals(s Shortcut) bool {
	if o.Id != s.Id {
		return false
	}

	if o.AppName != s.AppName {
		return false
	}

	if o.ExePath != s.ExePath {
		return false
	}

	if o.StartDir != s.StartDir {
		return false
	}

	if o.IconPath != s.IconPath {
		return false
	}

	if o.ShortcutPath != s.ShortcutPath {
		return false
	}

	if o.LaunchOptions != s.LaunchOptions {
		return false
	}

	if o.IsHidden != s.IsHidden {
		return false
	}

	if o.AllowDesktopConfig != s.AllowDesktopConfig {
		return false
	}

	if o.AllowOverlay != s.AllowOverlay {
		return false
	}

	if o.IsOpenVr != s.IsOpenVr {
		return false
	}

	if o.LastPlayTimeEpoch != s.LastPlayTimeEpoch {
		return false
	}

	if len(o.Tags) != len(s.Tags) {
		return false
	}

	for i := range o.Tags {
		if o.Tags[i] != s.Tags[i] {
			return false
		}
	}

	return true
}

func (o *Shortcut) object() vdf.Object {
	object := vdf.NewEmptyObject()

	object.Append(vdf.NewIdField(o.Id))

	object.Append(vdf.NewStringField(appNameField, o.AppName))

	object.Append(vdf.NewStringField(exePathField, appendDoubleQuotesIfNeeded(o.ExePath)))

	object.Append(vdf.NewStringField(startDirField, appendDoubleQuotesIfNeeded(o.StartDir)))

	object.Append(vdf.NewStringField(iconPathField, o.IconPath))

	object.Append(vdf.NewStringField(shortcutPathField, o.ShortcutPath))

	object.Append(vdf.NewStringField(launchOptionsField, o.LaunchOptions))

	object.Append(vdf.NewBoolField(isHiddenField, o.IsHidden))

	object.Append(vdf.NewBoolField(allowDesktopConfigField, o.AllowDesktopConfig))

	object.Append(vdf.NewBoolField(allowOverlayField, o.AllowOverlay))

	object.Append(vdf.NewBoolField(isOpenVrField, o.IsOpenVr))

	object.Append(vdf.NewInt32Field(lastPlayTimeField, o.LastPlayTimeEpoch))

	object.Append(vdf.NewSliceField(tagsField, o.Tags))

	return object
}

func appendDoubleQuotesIfNeeded(s string) string {
	doubleQuote := "\""

	if !strings.HasPrefix(s, doubleQuote) {
		s = doubleQuote + s
	}

	if !strings.HasSuffix(s, doubleQuote) {
		s = s + doubleQuote
	}

	return s
}
