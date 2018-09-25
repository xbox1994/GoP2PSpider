package types

import "fmt"

type QueryParam struct {
	Q     string
	Start int
}

type QueryResult struct {
	Hits      int64
	Start     int
	Query     string
	PrevStart int
	NextStart int
	Items     []interface{}
}

type Meta struct {
	Hash  string
	Name  string
	Size  uint64
	Files []File
}

func (meta *Meta) String() string {
	return fmt.Sprintf("Hash: %s\nName: %s\nSize: %v\nFiles: %v\n", meta.Hash, meta.Name, meta.Size, meta.Files)
}

type File struct {
	Path string
	Size uint64
}

func (file *File) String() string {
	return fmt.Sprintf("Path: %s, Size: %v", file.Path, file.Size)
}
