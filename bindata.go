package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _bindataGenesisDb = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x57\xcd\x6e\x1b\xb7\x13\x27\x25\xd9\x49\x14\x18\x86\x81\xff\xbf\x2e\x1a\xa0\xe0\xcd\x71\x61\x25\xfc\x5a\x92\x1b\xa0\x40\x94\xd4\x68\x8a\xb8\xa9\xeb\x38\x41\x7c\x32\xf8\x31\x6b\x2d\x2c\x69\x5d\xed\xca\x88\x0e\x3d\x28\x41\x7a\x2f\x7a\xcc\x0b\xf4\x98\x63\x9f\xa0\xb7\x9e\xf2\x00\x7d\x86\x02\x3d\x14\xe8\xa5\xd8\x95\x93\xca\xf9\x6c\x61\x37\x29\x1a\xff\x20\x90\xe2\x70\xf6\xc7\x99\x1d\xee\x70\x78\xf3\xcb\xb5\xb4\x00\x92\x64\x83\x9e\x2d\x88\x40\x35\x84\x31\xba\x4c\x08\x42\x68\x1e\x21\xd4\x44\x7f\xa2\x8e\x10\x6a\x4c\x8d\x31\x7a\x3d\xe6\x51\xeb\x97\x1f\x66\x4a\x65\xfc\x7b\x39\x3e\x35\xe9\x4e\xf0\xef\xc0\x6c\x6d\x2a\x36\xcd\x93\xd8\xbc\x6b\x38\xd3\x28\x3f\xf0\x59\xfc\x00\xe1\x5f\xf1\xcf\xf8\x47\xfc\x08\x7f\x8b\x1f\xbc\x6d\xab\xde\x39\x9c\xad\x9f\xc3\xb7\x61\x90\xa7\x59\x7f\xf6\xc3\x7a\x1b\xaf\x0f\x60\x3f\xcd\x86\xf9\x95\x6e\xe6\x77\xaf\xd9\xbc\x73\x33\xdd\xe9\xdb\x62\x38\x80\xc6\xff\xeb\x17\x9f\x9f\xae\xbd\x5f\xd7\x67\x3e\x85\x3e\xe4\x69\x7e\xa3\xd2\xbb\x06\x36\x74\xd3\x3e\xfc\xaf\xde\xc2\x57\x07\x60\x8b\x6c\xb0\x3e\x74\xdd\xd4\x5f\x87\xd1\x4c\xb9\xdc\x81\xb0\x3e\x57\x25\xf2\x65\x54\xfe\x5e\x89\xfb\x9b\xf8\xec\xfc\x78\xe1\xde\x68\xbc\x70\x6f\x19\xb5\xd9\x25\x2b\x3c\xd5\x90\x28\xef\x00\xb8\x54\x09\x17\xcc\x2a\x96\x24\x42\x39\xe7\x83\x01\x13\x29\xa1\xb9\x00\xa1\xa9\x00\x27\x64\x04\x3e\x84\x58\x88\xa0\x35\x8d\x2d\x70\x41\xa3\x58\x50\x26\xa8\xa2\x9a\x5b\xa3\xa4\xf1\x20\x02\xe5\x94\x51\x45\xcd\x94\x44\x50\x46\x35\x15\x92\x53\x4a\xa5\xf6\x54\x0b\x2b\x22\x16\x38\x33\x21\x66\xcc\x03\x37\x41\x47\x54\x48\x2d\x8c\x71\x41\x44\x11\x88\x60\x7d\x14\x80\x25\xb1\x00\x29\x55\x62\x79\xa2\x0c\x97\xc1\xb8\x48\x1a\x23\x3d\x00\x58\xc3\xa8\x89\x65\x1c\x39\x70\x41\x85\x10\x12\xc7\x29\x17\x89\x66\xda\x0a\x99\x50\x15\x59\x65\x63\xed\x8d\x70\x89\xe2\xc2\x80\x67\x41\x33\xe9\x8f\xc3\x6f\x19\x51\xce\x19\xa5\x49\x14\x2b\xa1\x63\x1e\x7b\xca\x20\xf1\xda\x38\x03\xcc\x49\x1e\x22\xcd\x99\x95\x4a\x7b\xad\xb9\x8a\xbd\xf4\x89\x0a\x4c\x50\x4a\x19\xa3\x49\xac\xbc\xf4\x2c\xe1\xd4\x53\xce\x29\xa7\xc2\x79\x15\x45\xdc\x26\x5e\x79\xa7\x59\xc4\x13\xae\x13\xc9\xbd\x62\x4c\x26\x06\x2c\xa7\x2e\x61\x52\x05\xea\xa8\xa1\xd2\x49\xe9\x82\x94\xc2\xaa\x84\x26\xb1\x31\xcd\x2a\xff\x3f\x46\xf8\xf1\x1b\xda\xeb\x27\xf8\xa7\x31\x6e\xce\x8e\x17\xc6\x0b\x67\x5e\xbb\x55\x11\x42\xa7\xe7\xf0\x6f\xa8\x86\xef\x20\x7c\x07\xff\x84\x5b\xb8\xf5\xb6\x8d\xff\xaf\x20\xc7\xa7\x16\x17\x17\xf1\x78\xb5\xb0\xae\x0b\xdb\x3d\x28\x6c\xd5\xd4\xae\x6e\xac\xb6\x37\x57\xc9\x66\xfb\xca\xda\x2a\xa9\x44\xe4\x7c\x93\x10\x42\x76\x61\x44\x9e\xe0\x76\x7b\xe3\xea\xb5\xf6\x06\xb9\xf1\xc5\x26\xb9\x71\x6b\x6d\x8d\xac\x6f\x7c\xf6\x79\x7b\x63\x8b\x5c\x5f\xdd\x5a\xa9\xb4\xf7\x6d\x77\x08\x87\xb5\x9b\xcb\xcb\xb5\xd9\xc5\x8f\x17\x31\x4a\xfb\x01\xee\xe6\x5f\x75\xd3\x02\xb6\xed\xb0\xc8\xaa\xf1\xf6\xc4\x08\x56\x75\x65\x2d\x7f\xba\xb4\xb3\xcc\xff\x35\xf4\x11\x42\xb7\xca\xe6\x55\x58\x6e\xbc\x92\x7c\x17\x46\xf9\x36\xab\xba\x99\xfb\xe7\xeb\x95\xf7\xdf\x5c\x9c\x78\x5f\xcd\x95\x4d\xe3\xb0\xf7\xa5\xe8\xc0\xfb\x6c\x8f\x1c\xc2\x21\xf7\x57\x26\xb2\x56\x8b\x2c\xb5\x97\xca\xbb\x0a\xb1\x21\xa4\xfd\x9d\x15\xb2\xb4\x31\x19\x0f\x60\x3f\xdb\x4d\xfb\x3b\x15\xd7\xde\xd0\xed\xc2\x68\xbb\x63\xf3\xce\x0b\xdf\xe6\xca\x84\x2b\xed\x93\xa2\xf3\xf4\xea\xb3\x54\x8c\xf6\xe0\x52\x07\xee\x2e\x4d\x71\x90\x97\x46\xe4\x80\xa3\x03\x77\x5b\xd0\xf7\x59\x80\x50\x3d\x96\xa7\x3b\x7f\x65\xe9\xdc\xf6\x9e\x2e\x6c\xf3\x69\x83\x9f\xb0\x4c\x8e\xfb\x97\x2d\x5e\x69\x95\x71\x0c\xb6\xb0\xcf\x98\x38\x99\x9c\xda\x30\xe4\xfc\x14\xff\xca\xb4\x89\xcb\xcd\xea\xd4\x7f\xed\x91\x75\xf4\x43\xef\xcd\x7c\x74\x27\xf8\xdb\x98\xc3\x5f\xa3\x06\xda\x42\xf8\x3b\xb4\x85\x2d\x7a\x84\xae\x1f\x13\xf1\xad\x5a\xe3\xe2\x78\xee\xb9\x52\x95\xd1\x23\x81\x8d\x2f\x34\x1a\xed\x7b\xad\x97\x57\xc8\x82\x4a\x35\x29\xaf\x82\xa3\x42\x5b\x50\xde\x09\xaf\x40\x68\xef\xbc\x01\x1f\xc5\xdc\x59\x0d\x10\x38\xa8\xc0\x0c\x28\x2b\xbc\x85\xe0\x25\x70\x30\x2c\x32\x0c\x5c\xac\xad\x77\x4a\x1f\xb0\x48\x15\x0c\x04\xae\x5d\x14\xb4\xb1\x26\xa2\xb1\x63\xe0\x4c\x24\x63\x1f\x3b\xe5\x4c\xc2\x3c\xb5\x4c\x48\x4f\xbd\xb6\x09\x17\xce\xc5\xc2\x3a\xe3\xbd\x91\x1c\x78\xf8\xa0\x5e\x3f\xd7\x3a\xa8\xb6\xd3\xfd\x6c\x60\xf3\xcb\x3b\x3d\x9b\x76\x2f\xf8\xac\x87\x10\x7a\xef\xf9\x37\xb7\x8b\x1b\x7a\xcc\x5e\x58\xce\x73\xca\x74\x8b\xca\x16\x37\x97\xc8\x7a\x3e\xf2\x1d\x08\xd0\x4d\x3d\xf1\x59\x6f\x2f\x1b\xf6\x43\x99\xd3\xc0\xe7\x85\xcd\x47\xa4\x97\xed\x43\x4e\x7c\x37\xcb\x61\x40\x8a\x8c\xd8\xbd\xbd\x41\xb6\x6f\xbb\xe5\xff\xa2\xb4\x87\xac\x6f\xde\xfc\xa4\xca\xff\xf8\x21\xc2\xdf\xe3\x87\xc7\x14\xf9\x13\x1c\x0f\xce\xce\xd6\xcf\xcd\x1f\x5c\x0b\xd9\xed\x99\x46\x6b\xbc\xf0\xec\x55\xee\xe8\x89\xf9\x8f\x00\x00\x00\xff\xff\x49\x1a\xba\x58\x00\x14\x00\x00")

func bindataGenesisDbBytes() ([]byte, error) {
	return bindataRead(
		_bindataGenesisDb,
		"bindata/genesis.db",
	)
}

func bindataGenesisDb() (*asset, error) {
	bytes, err := bindataGenesisDbBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "bindata/genesis.db", size: 5120, mode: os.FileMode(420), modTime: time.Unix(1494057426, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
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
var _bindata = map[string]func() (*asset, error){
	"bindata/genesis.db": bindataGenesisDb,
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
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"bindata": &bintree{nil, map[string]*bintree{
		"genesis.db": &bintree{bindataGenesisDb, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

