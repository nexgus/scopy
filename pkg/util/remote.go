package util

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/sftp"
)

func RemoteIsDirectory(client *sftp.Client, path string) (bool, error) {
	info, err := client.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return info.IsDir(), nil
}

func RemotePathExists(client *sftp.Client, path string) (bool, error) {
	_, err := client.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func RemoteMkdirAll(client *sftp.Client, remotePath string, remoteSep string) error {
	// TODO: 尚不支援遠端是 Windows 的狀況
	remotePath = ReplaceSepWith(remotePath, remoteSep)

	// 如果是絕對路徑, pathParts 的第一個素是空字串, 如 ["", "path", "starts", "from", "root"]
	pathParts := strings.Split(remotePath, remoteSep)
	if pathParts[0] == "" {
		pathParts[0] = remoteSep
	}

	// 遞迴檢查和建立
	remoteDir := ""
	for _, part := range pathParts {
		remoteDir = filepath.Join(remoteDir, part)
		if remoteDir == remoteSep || remoteDir == "." {
			continue
		}

		if exist, err := RemotePathExists(client, remoteDir); err != nil {
			return fmt.Errorf("檢查遠端路徑是否存在: %w", err)
		} else if !exist {
			if err := client.Mkdir(remoteDir); err != nil {
				return err
			}
		}
	}

	return nil
}

func RemoteScan(client *sftp.Client, path string, pattern string, excludes []string) ([]string, error) {
	regexpPattern := strings.ReplaceAll(pattern, ".", "\\.")
	regexpPattern = strings.ReplaceAll(regexpPattern, "*", ".*")
	regexpPattern = strings.ReplaceAll(regexpPattern, "?", ".")

	// 確保只匹配整個 'base' 名稱（如果 pattern 不包含路徑分隔符）
	// 實際情況可能更複雜，但對於簡單的檔案名匹配，這足夠了。
	re, err := regexp.Compile(regexpPattern)
	if err != nil {
		return nil, fmt.Errorf("編譯正規表達式 %q: %v", regexpPattern, err)
	}

	// 讀取遠端目錄內容
	entries, err := client.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("讀取遠端目錄 %s: %v", path, err)
	}

	var paths []string
	for _, entry := range entries {
		baseName := entry.Name()
		fullPath := path + "/" + baseName // 簡單地連接路徑

		// 檢查是否在排除列表中
		if isExcluded(baseName, excludes) {
			continue
		}

		// 檢查名稱是否匹配 pattern (類似 Glob 的效果)
		// 只有匹配 pattern 的目錄和檔案才會被考慮
		if entry.IsDir() || re.MatchString(baseName) {
			if entry.IsDir() {
				if _paths, err := RemoteScan(client, fullPath, pattern, excludes); err != nil {
					return nil, err
				} else {
					if len(_paths) == 0 {
						paths = append(paths, fullPath)
					} else {
						paths = append(paths, _paths...)
					}
				}
			} else {
				paths = append(paths, fullPath)
			}
		}
	}

	return paths, nil
}
