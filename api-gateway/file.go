package main

import (
	"os"
	"path"
)

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

func createFile(filename string) (*os.File, error) {
	if checkFileIsExist(filename) {
		return os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	}

	err := os.MkdirAll(path.Dir(filename), 0777)
	if err != nil {
		return nil, err
	}

	return os.Create(filename)
}
