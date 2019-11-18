package migrations

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

var __1_create_initial_tables_up_sql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x94\x90\xc1\x6a\x83\x40\x10\x86\xef\x3e\xc5\x7f\x8c\xd0\x37\xf0\xb4\xd5\x49\x59\xaa\x6b\x58\x47\x30\xa7\x45\xa2\x84\x6d\xa9\x11\x9d\xf7\xa7\xd4\xc6\x66\xa1\xa6\x25\x73\xdc\x99\xf9\xf6\xfb\x27\xb5\xa4\x98\xc0\xea\x39\x27\xe8\x3d\x4c\xc9\xa0\x46\x57\x5c\x61\xee\x45\xfc\x70\x9e\xb1\x8b\x00\xc0\x77\x08\x4b\x1b\xa6\x17\xb2\x38\x58\x5d\x28\x7b\xc4\x2b\x1d\xa1\x6a\x2e\xb5\x49\x2d\x15\x64\xf8\x69\xd9\xfa\x81\x7c\x17\x53\xc3\xcb\x1f\xa6\xce\xf3\x28\x4e\xa2\xe8\x0f\x81\x37\x3f\xb5\x4e\xfc\x47\xef\x64\x6a\x4f\xef\x7e\x38\x6f\xab\x3c\xe2\xb3\x30\xef\x44\x59\xbd\x82\xc9\xae\x9f\x4f\x93\x1f\xc5\x5f\x86\x5f\xfa\xd7\x80\xd2\x4e\xd2\x77\xae\x95\x1b\x30\x53\x4c\xac\x0b\x42\x46\x7b\x55\xe7\x8c\xb4\xb6\x96\x0c\xbb\xaf\xc7\x8a\x55\x71\x58\x77\x2f\xe3\x78\x67\x37\x3c\x8e\x36\x19\x35\xff\x1e\xc7\xad\xd9\x4a\xb3\xd1\xdd\x5d\xbb\x71\xf2\x20\x35\x08\xb8\x0d\xbe\x0d\xc4\xc9\x67\x00\x00\x00\xff\xff\xb4\x99\x90\x74\x4f\x02\x00\x00")

func _1_create_initial_tables_up_sql() ([]byte, error) {
	return bindata_read(
		__1_create_initial_tables_up_sql,
		"1_create_initial_tables.up.sql",
	)
}

var _migrations_go = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x01\x00\x00\xff\xff\x00\x00\x00\x00\x00\x00\x00\x00")

func migrations_go() ([]byte, error) {
	return bindata_read(
		_migrations_go,
		"migrations.go",
	)
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() ([]byte, error){
	"1_create_initial_tables.up.sql": _1_create_initial_tables_up_sql,
	"migrations.go": migrations_go,
}
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func func() ([]byte, error)
	Children map[string]*_bintree_t
}
var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	"1_create_initial_tables.up.sql": &_bintree_t{_1_create_initial_tables_up_sql, map[string]*_bintree_t{
	}},
	"migrations.go": &_bintree_t{migrations_go, map[string]*_bintree_t{
	}},
}}
