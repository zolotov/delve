package dwarf

import (
	"debug/dwarf"

	"github.com/derekparker/delve/pkg/dwarf/frame"
	"github.com/derekparker/delve/pkg/dwarf/line"
	"github.com/derekparker/delve/pkg/dwarf/reader"
)

type Dwarf struct {
	Frame frame.FrameDescriptionEntries
	Line  line.DebugLines
	Rdr   *reader.Reader

	data  *dwarf.Data
	types map[string]dwarf.Offset
}

func Parse(path string) (*Dwarf, error) {
	exe, err := newExecutable(path)
	if err != nil {
		return nil, err
	}
	d, err := exe.DWARF()
	if err != nil {
		return nil, err
	}
	line, err := parseLine(exe)
	if err != nil {
		return nil, err
	}
	frame, err := parseFrame(exe)
	if err != nil {
		return nil, err
	}
	rdr := reader.New(d)
	return &Dwarf{
		Frame: frame,
		Line:  line,
		Rdr:   rdr,
		data:  d,
		types: loadTypes(rdr),
	}, nil
}

func (d *Dwarf) TypeNamed(name string) (dwarf.Type, error) {
	off, found := d.types[name]
	if !found {
		return nil, reader.TypeNotFoundErr
	}
	return d.data.Type(off)
}

func loadTypes(rdr *reader.Reader) map[string]dwarf.Offset {
	types := make(map[string]dwarf.Offset)
	for entry, err := rdr.NextType(); entry != nil; entry, err = rdr.NextType() {
		if err != nil {
			break
		}
		name, ok := entry.Val(dwarf.AttrName).(string)
		if !ok {
			continue
		}
		if _, exists := dbp.types[name]; !exists {
			types[name] = entry.Offset
		}
	}
	rdr.Seek(0)
	return types
}
