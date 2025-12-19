package transport

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"scopy/pkg/util"

	"github.com/pkg/sftp"
)

// Upload 上傳單一檔案或目錄下所有檔案.
func Upload(
	client *sftp.Client,
	remotePath string,
	localPath string,
	excludes []string,
	remoteSep string,
) error {
	localInfo, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("取得本地路徑 (%s) 資訊: %w", localPath, err)
	}

	if localInfo.IsDir() {
		if localPath == "." {
			localPath, _ = os.Getwd()
		}
		return uploadLocalDir(client, remotePath, localPath, excludes, remoteSep)
	} else {
		if remoteInfo, err := client.Stat(remotePath); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("取得遠端路徑 (%s) 資訊: %w", remotePath, err)
			}
		} else if remoteInfo.IsDir() {
			remotePath = filepath.Join(remotePath, filepath.Base(localPath))
		}

		return uploadLocalFile(client, remotePath, localPath, remoteSep)
	}
}

func uploadLocalDir(
	client *sftp.Client,
	remoteDir string,
	localDir string,
	excludes []string,
	remoteSep string,
) error {
	remoteRoot := remoteDir
	if remoteRoot == "." {
		remoteRoot = filepath.Base(localDir)
	}

	if errWalk := filepath.Walk(localDir, func(localPath string, localInfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("掃描本地檔案系統錯誤: %w", err)
		}

		if isMatched(localPath, excludes) {
			// 不用繼續往下做了
			return nil
		}

		relPath, err := filepath.Rel(localDir, localPath)
		if err != nil {
			return fmt.Errorf("取得本地路徑 (%s) 相對路徑: %w", localPath, err)
		}

		if relPath == "." {
			if remoteStat, err := client.Stat(remoteRoot); err != nil {
				if os.IsNotExist(err) {
					fmt.Printf("建立遠端目錄 %s\n", remoteRoot)
					if err := util.RemoteMkdirAll(client, remoteRoot, remoteSep); err != nil {
						return fmt.Errorf("建立遠端目錄: %w", err)
					}
				} else {
					return fmt.Errorf("取得遠端目錄資訊: %w", err)
				}
			} else if !remoteStat.IsDir() {
				return fmt.Errorf("遠端路徑 (%s) 存在且不是目錄", remoteRoot)
			}
		} else {
			remotePath := filepath.Join(remoteRoot, relPath)
			if localInfo.IsDir() {
				fmt.Printf("建立遠端目錄 %s\n", remotePath)
				if err := util.RemoteMkdirAll(client, remotePath, remoteSep); err != nil {
					return fmt.Errorf("建立遠端目錄: %w", err)
				}
			} else {
				if err := uploadLocalFile(client, remotePath, localPath, remoteSep); err != nil {
					return fmt.Errorf("上傳本地檔案: %w", err)
				}
			}
		}

		return nil
	}); errWalk != nil {
		return errWalk
	}

	return nil
}

func uploadLocalFile(client *sftp.Client, remotePath string, localPath string, remoteSep string) error {
	remotePath = util.ReplaceSepWith(remotePath, remoteSep)

	fmt.Printf("開啟本地檔案 %s\n", localPath)
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("開啟本地檔案 (%s): %w", localPath, err)
	}
	defer localFile.Close()

	remoteDir := filepath.Dir(remotePath)
	if err := util.RemoteMkdirAll(client, remoteDir, remoteSep); err != nil {
		return fmt.Errorf("建立遠端目錄 (%s): %w", remoteDir, err)
	}

	fmt.Printf("建立遠端檔案 %s\n", remotePath)
	remoteFile, err := client.Create(remotePath)
	if err != nil {
		return fmt.Errorf("建立遠端檔案 (%s): %w", remotePath, err)
	}
	defer remoteFile.Close()

	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("複製檔案至遠端: %w", err)
	}

	localStat, err := client.Stat(localPath)
	if err != nil {
		// 只是無法複製屬性, 不管它
		return nil
	}

	mtime := localStat.ModTime()
	client.Chtimes(localPath, mtime, mtime)

	mode := localStat.Mode()
	client.Chmod(localPath, mode)

	return nil
}
