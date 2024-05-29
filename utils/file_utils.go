package utils

import (
	"errors"
	"fmt"
	"os"
)

func GetFileSize(filepath string) (int64, error) {
	fi, err := os.Stat(filepath)
	if err != nil {
		return 0, err
	}
	// get the size
	return fi.Size(), nil
}

func EnsureDir(dirName string) error {
	err := os.Mkdir(dirName, 0766)
	if err == nil {
		return nil
	}
	if os.IsExist(err) {
		// check that the existing path is a directory
		info, err := os.Stat(dirName)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return errors.New("path exists but is not a directory")
		}
		return nil
	}
	return err
}

func GetFileSizeInMb(filePath string) (float64, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	// get the size in Mb
	size := float64(fi.Size()) / 1000 / 1000
	return size, nil
}

func GetFileName(filePath string) string {
	fi, err := os.Stat(filePath)
	if err != nil {
		fmt.Println(err)
		return ""
	}
	return fi.Name()
}

func CreateDirByPath(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func RemoveDirWithFiles(dirPath string) error {
	err := os.RemoveAll(dirPath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
