package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
)

type (
	Shortcuts []Shortcut
	Shortcut  struct {
		AppId         uint32   `json:"AppId"`
		Appname       string   `json:"AppName"`
		Exe           string   `json:"Exe"`
		StartDir      string   `json:"StartDir"`
		LaunchOptions string   `json:"LaunchOptions"`
		Icon          string   `json:"Icon"`
		Tags          []string `json:"Tags"`
		// ShortcutPath        string
		// IsHidden            uint32
		// AllowDesktopConfig  uint32
		// AllowOverlay        uint32
		// OpenVR              uint32
		// Devkit              uint32
		// DevkitGameId        string
		// DevkitOverrideAppId uint32
		// LastPlayTime        uint32
		// FlatpakAppId        string
	}
)

func writeVdfString(buf *bytes.Buffer, key string, value string) error {
	err := binary.Write(buf, binary.LittleEndian, []byte{0x01})
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.LittleEndian, []byte(key))
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.LittleEndian, []byte{0x00})
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.LittleEndian, []byte(value))
	if err != nil {
		return err
	}
	return binary.Write(buf, binary.LittleEndian, []byte{0x00})
}

func writeVdfInt(buf *bytes.Buffer, key string, value uint32) error {
	err := binary.Write(buf, binary.LittleEndian, []byte{0x02})
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.LittleEndian, []byte(key))
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.LittleEndian, []byte{0x00})
	if err != nil {
		return err
	}
	return binary.Write(buf, binary.LittleEndian, value)
}

func writeVdfStringSlice(buf *bytes.Buffer, key string, value []string) error {
	err := binary.Write(buf, binary.LittleEndian, []byte{0x00})
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.LittleEndian, []byte(key))
	if err != nil {
		return err
	}
	err = binary.Write(buf, binary.LittleEndian, []byte{0x00})
	if err != nil {
		return err
	}
	for c, v := range value {
		err = binary.Write(buf, binary.LittleEndian, []byte{0x01})
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, []byte(strconv.Itoa(c)))
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, []byte{0x00})
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, []byte(v))
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, []byte{0x00})
		if err != nil {
			return err
		}
	}
	return binary.Write(buf, binary.LittleEndian, []byte{0x08})
}

func (ss Shortcuts) ToVdf() ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	err := binary.Write(buf, binary.LittleEndian, []byte{0x00})
	if err != nil {
		return []byte{}, err
	}
	err = binary.Write(buf, binary.LittleEndian, []byte("shortcuts"))
	if err != nil {
		return []byte{}, err
	}
	err = binary.Write(buf, binary.LittleEndian, []byte{0x00})
	if err != nil {
		return []byte{}, err
	}

	for i, s := range ss {
		err = binary.Write(buf, binary.LittleEndian, []byte{0x00})
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, []byte(strconv.Itoa(i)))
		if err != nil {
			return err
		}
		err = binary.Write(buf, binary.LittleEndian, []byte{0x00})
		if err != nil {
			return err
		}
		err = writeVdfInt(buf, "appid", s.AppId)
		if err != nil {
			return []byte{}, err
		}
		err = writeVdfString(buf, "appname", s.Appname)
		if err != nil {
			return []byte{}, err
		}
		err = writeVdfString(buf, "exe", s.Exe)
		if err != nil {
			return []byte{}, err
		}
		err = writeVdfString(buf, "StartDir", s.StartDir)
		if err != nil {
			return []byte{}, err
		}
		err = writeVdfString(buf, "LaunchOptions", s.LaunchOptions)
		if err != nil {
			return []byte{}, err
		}
		err = writeVdfString(buf, "icon", s.Icon)
		if err != nil {
			return []byte{}, err
		}
		err = writeVdfStringSlice(buf, "tags", s.Tags)
		if err != nil {
			return []byte{}, err
		}
		err = binary.Write(buf, binary.LittleEndian, []byte{0x08})
		if err != nil {
			return []byte{}, err
		}

	}
	err = binary.Write(buf, binary.LittleEndian, []byte{0x08, 0x08})
	if err != nil {
		return []byte{}, err
	}

	return buf.Bytes(), nil
}

func hash(s string) uint32 {
	h := fnv.New32()
	h.Write([]byte(s))
	return h.Sum32()
}

func main() {
	var filepath string
	flag.StringVar(&filepath, "file", "-", "the path to the json. when - is used it's read from stdin")

	flag.Parse()

	stream := os.Stdin

	if filepath != "-" {
		file, err := os.Open(filepath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Couldn't open file %s: %s\n", filepath, err.Error())
			os.Exit(1)
		}
		defer file.Close()
		stream = file
	}

	var s Shortcuts
	err := json.NewDecoder(stream).Decode(&s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't parse json: %s\n", err.Error())
		os.Exit(1)
	}

	// generate reproducable id if it's not explicitely set
	for i, shortcut := range s {
		if shortcut.AppId == 0 {
			s[i].AppId = hash(shortcut.Appname)
		}
	}

	b, err := s.ToVdf()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	_, err = fmt.Print(string(b))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
