package transport

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"scopy/pkg/util"

	"github.com/pkg/sftp"
)

// Download 從遠端下載一個檔案或目錄到本地指定的路徑.
func Download(
	client *sftp.Client,
	remotePath string,
	localPath string,
	excludes []string,
) error {
	remoteInfo, err := client.Stat(remotePath)
	if err != nil {
		return fmt.Errorf("取得遠端路徑資訊: %w", err)
	}

	if remoteInfo.IsDir() {
		return downloadRemoteDir(client, remotePath, localPath, excludes)
	} else {
		if localInfo, err := os.Stat(localPath); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("取得本地路徑資訊: %w", err)
			}
		} else if localInfo.IsDir() {
			localPath = filepath.Join(localPath, filepath.Base(remotePath))
		}

		fmt.Printf("複製遠地檔案 %s\n", remotePath)
		return downloadRemoteFile(client, remotePath, localPath)
	}
}

func createLocalDir(client *sftp.Client, remoteDir string, localDir string) error {
	remoteStat, err := client.Stat(remoteDir)
	if err != nil {
		return fmt.Errorf("取得遠地目錄資訊: %w", err)
	}

	mode := remoteStat.Mode()
	if err := os.MkdirAll(localDir, mode); err != nil {
		return fmt.Errorf("建立本地目錄: %w", err)
	}

	mtime := remoteStat.ModTime()
	os.Chtimes(localDir, mtime, mtime)

	return nil
}

func downloadRemoteDir(
	client *sftp.Client,
	remoteDir string,
	localDir string,
	excludes []string,
) error {
	localRoot := localDir
	if localRoot == "." {
		localRoot = filepath.Base(remoteDir)
	}

	walker := client.Walk(remoteDir)
	for walker.Step() {
		if walker.Err() != nil {
			return fmt.Errorf("開始搜尋目錄: %w", walker.Err())
		}

		remotePath := walker.Path()
		if isMatched(remotePath, excludes) {
			continue
		}

		relPath, err := filepath.Rel(remoteDir, remotePath)
		if err != nil {
			return fmt.Errorf("取得相對路徑: %w", err)
		}

		if relPath == "." {
			if util.PathExists(localRoot) {
				if !util.IsDirectory(localRoot) {
					return fmt.Errorf("本地路徑 (%s) 存在且不是目錄", localRoot)
				}
			} else {
				fmt.Printf("建立本地目錄 %s\n", localRoot)
				if err := createLocalDir(client, remotePath, localRoot); err != nil {
					return fmt.Errorf("建立本地目錄: %w", err)
				}
			}
		} else {
			remoteStat, err := client.Stat(remotePath)
			if err != nil {
				return fmt.Errorf("取得遠地目錄資訊: %w", err)
			}

			localPath := filepath.Join(localRoot, relPath)
			if remoteStat.IsDir() {
				fmt.Printf("複製遠地目錄 %s\n", remotePath)
				if err := createLocalDir(client, remotePath, localPath); err != nil {
					return fmt.Errorf("建立本地目錄: %w", err)
				}
			} else {
				fmt.Printf("複製遠地檔案 %s\n", remotePath)
				if err := downloadRemoteFile(client, remotePath, localPath); err != nil {
					return fmt.Errorf("下載遠地檔案: %w", err)
				}
			}
		}
	}

	return nil
}

func downloadRemoteFile(client *sftp.Client, remotePath string, localPath string) error {
	remoteFile, err := client.Open(remotePath)
	if err != nil {
		return fmt.Errorf("開啟遠地檔案: %w", err)
	}
	defer remoteFile.Close()

	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return fmt.Errorf("建立本地目錄: %w", err)
	}

	localFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("建立本地檔案: %w", err)
	}
	defer localFile.Close()

	_, err = io.Copy(localFile, remoteFile)
	if err != nil {
		return fmt.Errorf("複製檔案: %w", err)
	}

	// Windows 必須確保緩衝區寫入磁碟才能做 chtime 與 chmod
	if err := localFile.Sync(); err != nil {
		fmt.Printf("[警告] 無法同步本地檔案: %v", err)
	}

	remoteStat, err := client.Stat(remotePath)
	if err != nil {
		// 只是無法複製屬性, 不管它
		return nil
	}

	mtime := remoteStat.ModTime()
	os.Chtimes(localPath, mtime, mtime)

	mode := remoteStat.Mode()
	os.Chmod(localPath, mode)

	return nil
}

func isMatched(path string, patterns []string) bool {
	for idx, pattern := range patterns {
		patterns[idx] = filepath.ToSlash(pattern)
	}

	segments := strings.SplitSeq(filepath.ToSlash(path), "/")
	for segment := range segments {
		if segment == "" {
			continue
		}

		for _, pattern := range patterns {
			matched, err := filepath.Match(pattern, segment)
			if err == nil && matched {
				return true
			}
		}
	}

	return false
}
