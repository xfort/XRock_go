package xutil

import (
	"crypto/md5"
	"io"
	"os"
)

// 计算 大文件 md5,
func Md5File(absFilepath string) ([]byte, error) {
	file, err := os.Open(absFilepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	hashMd5 := md5.New()
	_, err = io.Copy(hashMd5, file)
	if err != nil {
		return nil, err
	}
	return hashMd5.Sum(nil), nil
}
