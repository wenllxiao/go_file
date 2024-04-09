package main

import (
	"archive/zip"
	"bytes"
	"github.com/golang-module/carbon"
	"io"
	"os"
	"path/filepath"
)

func main() {

}

// zipFile 压缩文件
func zipFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	fileName := filepath.Base(filePath)
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	zipEntry, err := zw.Create(fileName)
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(zipEntry, file); err != nil {
		return "", err
	}

	if err = zw.Close(); err != nil {
		return "", err
	}
	zipFilePath := filepath.Dir(filePath) + "/" + fileName + ".zip"
	if err = os.WriteFile(zipFilePath, buf.Bytes(), 0644); err != nil {
		return "", err
	}
	return zipFilePath, nil
}

func renameFile(fileName string) (string, error) {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return "", err
	}
	var newFileName string
	newFileName = fileName + "." + carbon.Now().ToShortTimeString()
	if err := os.Rename(fileName, newFileName); err != nil {
		return "", err
	}
	return newFileName, nil
}
