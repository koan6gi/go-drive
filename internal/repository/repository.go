package repository

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"

	repErr "github.com/koan6gi/go-drive/internal/repository/errors"
)

type FileSystem struct {
	mu sync.Mutex
	st *FSItem
}

type FSItem struct {
	Type  int
	Path  string
	Name  string
	Entry map[string]*FSItem
}

// FSItem types
const (
	fsFile = iota
	fsDir
)

// DirEntry represents file/directory information
type DirEntry struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// DirEntry types
const (
	deFile = "file"
	deDir  = "dir"
)

type DirByAlphabet []DirEntry

func (d DirByAlphabet) Len() int      { return len(d) }
func (d DirByAlphabet) Swap(i, j int) { d[i], d[j] = d[j], d[i] }
func (d DirByAlphabet) Less(i, j int) bool {
	return ((d[i].Type == d[j].Type) && (d[i].Name < d[j].Name)) || (d[i].Type < d[j].Type)
}

const StorageDirectory = "./storage"

// TODO: make ext-error types
type Storage interface {
	Lock()
	Unlock()
	CreateFile(path string) (*os.File, error)
	CreateDirectory(path string) error
	GetFile(path string) (*os.File, error)
	Delete(path string) error
	Copy(dest string, src string) error
	Move(dest string, src string) error
	List(path string) (*[]DirEntry, error)
}

var FileStorage Storage

func SetupStorage() error {
	var err error
	FileStorage, err = NewFileStorage()
	return err
}

func NewFileStorage() (*FileSystem, error) {
	storage := &FileSystem{
		mu: sync.Mutex{},
		st: &FSItem{
			Type:  fsDir,
			Path:  StorageDirectory,
			Entry: make(map[string]*FSItem),
		},
	}

	err := os.Mkdir(StorageDirectory, 0777)
	if err != nil && !os.IsExist(err) {
		return nil, &repErr.SystemError{
			Err:     err,
			Content: fmt.Sprintf("can't create local storage dir: %v", err),
		}
	}

	return storage, walkDir(storage.st)
}

func (st *FileSystem) Lock() {
	st.mu.Lock()
}

func (st *FileSystem) Unlock() {
	st.mu.Unlock()
}

func walkDir(d *FSItem) error {
	path := d.Path
	file, err := os.Open(path)
	if err != nil {
		return &repErr.SystemError{
			Err:     err,
			Content: fmt.Sprintf("can't open file: %s: %v", path, err),
		}
	}
	defer file.Close()

	dir, _ := file.Readdir(0)

	for _, v := range dir {
		name := v.Name()
		if name == "." || name == ".." {
			continue
		}

		newItem := &FSItem{
			Type:  fsFile,
			Path:  path + "/" + name,
			Name:  name,
			Entry: nil,
		}
		d.Entry[name] = newItem

		if v.IsDir() {
			newItem.Type = fsDir
			newItem.Entry = make(map[string]*FSItem)
			err := walkDir(newItem)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (st *FileSystem) getParentDirectory(path string) (*FSItem, error) {
	if path == "/" {
		return st.st, nil
	}

	err := &repErr.PathError{
		Content: fmt.Sprintf("bad path: %s", path),
	}

	paths := strings.Split(path, "/")
	if len(paths) == 1 {
		return nil, err
	}

	paths = paths[1 : len(paths)-1]
	if len(paths) == 0 {
		return st.st, nil
	}

	lastIndex := len(paths) - 1

	dirEntry := st.st

	for i, v := range paths {
		newEntry, ok := dirEntry.Entry[v]
		if !ok {
			break
		}

		if i == lastIndex && newEntry.Type == fsDir {
			return newEntry, nil
		}

		dirEntry = newEntry
	}

	return nil, err
}

func (st *FileSystem) getItem(path string) (*FSItem, error) {
	if path == "/" {
		return st.st, nil
	}

	name := path[strings.LastIndex(path, "/")+1:]
	dir, err := st.getParentDirectory(path)
	if err != nil {
		return nil, err
	}

	if item, ok := dir.Entry[name]; ok {
		return item, nil
	}

	return nil, &repErr.PathError{
		Content: fmt.Sprintf("bad path: %s", path),
	}
}

func (st *FileSystem) GetFile(path string) (*os.File, error) {
	item, err := st.getItem(path)
	if err != nil {
		return nil, err
	}

	if item.Type != fsFile {
		return nil, &repErr.PathError{
			Content: fmt.Sprintf("not file: %s", path),
		}
	}

	file, err := os.Open(item.Path)
	if err != nil {
		err = &repErr.SystemError{
			Err:     err,
			Content: fmt.Sprintf("can't open file: %s, %v", item.Path, err),
		}
	}

	return file, err
}

func (st *FileSystem) CreateFile(path string) (*os.File, error) {
	name := path[strings.LastIndex(path, "/")+1:]
	dir, err := st.getParentDirectory(path)
	if err != nil {
		return nil, err
	}

	if _, ok := dir.Entry[name]; ok {
		return nil, &repErr.PathError{
			Content: fmt.Sprintf("path %s is already exist", path),
		}
	}

	newFile := &FSItem{
		Type:  fsFile,
		Name:  name,
		Path:  dir.Path + "/" + name,
		Entry: nil,
	}

	file, err := os.Create(newFile.Path)
	if err != nil {
		return nil, &repErr.SystemError{
			Err:     err,
			Content: fmt.Sprintf("can't create file: %s: %v", newFile.Path, err),
		}
	}

	_ = file.Chmod(0777)

	dir.Entry[name] = newFile

	return file, nil
}

func (st *FileSystem) CreateDirectory(path string) error {
	name := path[strings.LastIndex(path, "/")+1:]
	dir, err := st.getParentDirectory(path)
	if err != nil {
		return err
	}

	if _, ok := dir.Entry[name]; ok {
		return &repErr.PathError{
			Content: fmt.Sprintf("path %s is already exist", path),
		}
	}

	newDir := &FSItem{
		Type:  fsDir,
		Name:  name,
		Path:  dir.Path + "/" + name,
		Entry: make(map[string]*FSItem),
	}

	err = os.Mkdir(newDir.Path, 0777)
	if err != nil {
		return &repErr.SystemError{
			Err:     err,
			Content: fmt.Sprintf("can't create directory: %s: %v", newDir.Path, err),
		}
	}

	dir.Entry[name] = newDir

	return nil
}

func (st *FileSystem) Delete(path string) error {
	if path == "/" {
		return &repErr.PathError{
			Content: "can't delete root",
		}
	}

	item, err := st.getItem(path)
	if err != nil {
		return err
	}
	dir, _ := st.getParentDirectory(path)
	delete(dir.Entry, item.Name)

	err = nil

	switch item.Type {
	case fsFile:
		err = os.Remove(item.Path)
	case fsDir:
		err = os.RemoveAll(item.Path)
	}

	if err != nil {
		err = &repErr.SystemError{
			Err:     err,
			Content: fmt.Sprintf("can't delete file or directory: %s: %v", item.Path, err),
		}
	}

	return err
}

func (st *FileSystem) Move(dest string, src string) error {
	err := st.Copy(dest, src)
	if err != nil {
		return err
	}

	//name := src[strings.LastIndex(src, "/")+1:]
	//newPath := dest+"/"+name

	err = st.Delete(src)
	if err != nil {
		return err // TODO: return storage to normal stage
	}

	return nil
}

func (st *FileSystem) Copy(dest string, src string) error {
	srcFile, err := st.GetFile(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	name := src[strings.LastIndex(src, "/")+1:]
	newPath := dest + "/" + name

	destFile, err := st.CreateFile(newPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil && err != io.EOF {
		destFile.Close()
		_ = st.Delete(newPath)
		return &repErr.SystemError{
			Err:     err,
			Content: fmt.Sprintf("can't copy %s to %s: %v", src, dest, err),
		}
	}

	return nil
}

func (st *FileSystem) List(path string) (*[]DirEntry, error) {
	item, err := st.getItem(path)
	if err != nil {
		return nil, err
	}
	if item.Type != fsDir {
		return nil, fmt.Errorf("%s hot directory", path)
	}

	result := make([]DirEntry, 0)

	for _, v := range item.Entry {
		itemType := deFile
		if v.Type == fsDir {
			itemType = deDir
		}

		result = append(result, DirEntry{
			Name: v.Name,
			Path: v.Path[len(StorageDirectory):],
			Type: itemType,
		})
	}

	sort.Sort(DirByAlphabet(result))

	return &result, nil
}
