package copy

import (
	"io"
	"os"
	"path"
	"path/filepath"
)

type ICopy interface {
	CopyReader(file io.Reader, filePath string) error
}

type LocalCopy struct {
	BasePath string
}

func NewLocalCopy(basePath string) *LocalCopy {
	return &LocalCopy{BasePath: basePath}
}

func (l LocalCopy) CopyFile(file *os.File, filePath string) error {
	realPath := path.Join(l.BasePath, filePath)
	err := os.MkdirAll(filepath.Dir(realPath), 0777)
	if err != nil {
		return err
	}
	newFile, err := os.Create(realPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(newFile, file)
	return err
}
func (l LocalCopy) CopyReader(file io.Reader, filePath string) error {
	realPath := path.Join(l.BasePath, filePath)
	err := os.MkdirAll(filepath.Dir(realPath), 0777)
	if err != nil {
		return err
	}
	newFile, err := os.Create(realPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(newFile, file)
	return err
}
