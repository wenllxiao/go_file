package main

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// UnZipFile 解压缩 ZIP 文件到指定目录
func UnZipFile(src, dest string) (string, error) {
	// 打开 ZIP 文件
	r, err := zip.OpenReader(src)
	if err != nil {
		return "", err
	}
	defer r.Close()
	var isFirst = true
	var filePath string
	// 遍历 ZIP 文件中的每个文件
	for _, f := range r.File {
		// 跳过隐藏文件
		if strings.HasPrefix(f.Name, ".") {
			continue
		}
		if isFirst {
			topLevelDir := ConvertFileNameToUTF8(strings.Split(f.Name, "/")[0])
			filePath = filepath.Join(dest, topLevelDir)
			isFirst = false
		}
		f.Name = ConvertFileNameToUTF8(f.Name)
		// 构建目标文件路径
		fpath := filepath.Join(dest, f.Name)
		// 如果是目录，则创建目录
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return "", err
			}
			continue
		}

		// 创建文件所在的目录
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return "", err
		}

		// 打开目标文件
		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", err
		}

		// 打开 ZIP 文件中的文件
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		// 将 ZIP 文件中的内容复制到目标文件
		_, err = io.Copy(outFile, rc)
		// 关闭文件
		outFile.Close()
		rc.Close()

		if err != nil {
			return "", err
		}
	}
	return filePath, nil
}

// ConvertFileNameToUTF8 将文件名转换为 UTF-8 编码
func ConvertFileNameToUTF8(name string) string {
	if utf8.ValidString(name) {
		return name
	}
	decodeGBK, err := decodeGBK(name)
	if err == nil {
		return decodeGBK
	}
	return name
}

// decodeGBK 尝试将GBK编码的字符串转换为UTF-8
func decodeGBK(s string) (string, error) {
	reader := transform.NewReader(strings.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
