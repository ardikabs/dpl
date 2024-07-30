package ioutils

import (
	"io"
	"os"
	"path/filepath"
)

func ReadFile(filename string) ([]byte, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return content, nil
}

func ReadAndOpenFile(filename string) ([]byte, io.WriteCloser, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	writer, err := createOrOpenFile(filename)
	if err != nil {
		return nil, nil, err
	}

	return content, writer, nil
}

func createOrOpenFile(filename string) (*os.File, error) {
	if err := ensureDirExists(filename); err != nil {
		return nil, err
	}

	return os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
}

func ensureDirExists(filename string) error {
	basedir := filepath.Dir(filename)

	_, err := os.Stat(basedir)
	if os.IsNotExist(err) {
		return os.MkdirAll(basedir, 0755)
	}

	return err
}
