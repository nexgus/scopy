package util

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		logrus.Warnf("獲取路徑資訊失敗: %v", err)
		return false
	}

	return info.IsDir()
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		logrus.Warnf("獲取路徑資訊失敗: %v", err)
		return false
	}

	return true
}

func Scan(path string, pattern string, excludes []string) ([]string, error) {
	pathPattern := fmt.Sprintf("%s/%s", path, pattern)
	matches, err := filepath.Glob(pathPattern)
	if err != nil {
		return nil, fmt.Errorf("搜尋 %s 發生錯誤: %v", pathPattern, err)
	}

	var paths []string
	for _, _path := range matches {
		if isExcluded(filepath.Base(_path), excludes) {
			continue
		}

		if IsDirectory(_path) {
			if _paths, err := Scan(_path, pattern, excludes); err != nil {
				return nil, err
			} else {
				if len(_paths) == 0 {
					paths = append(paths, _path)
				} else {
					paths = append(paths, _paths...)
				}
			}
		} else {
			paths = append(paths, _path)
		}
	}

	return paths, nil
}
