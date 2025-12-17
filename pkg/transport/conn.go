package transport

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var defaultPrivateKeys = []string{
	"~/.ssh/id_ed25519",
	"~/.ssh/id_rsa",
	"~/.ssh/id_dsa",
	"~/.ssh/id_ecdsa",
}

// Connect 函數用於建立 SSH 連線
// 它會嘗試使用 PrivateKey 進行認證, 如果失敗或未指定, 則嘗試使用密碼認證.
func Connect(host string, port uint16, username string, key string, forcePassword bool) (*ssh.Client, error) {
	portStr := "22"
	if port > 0 {
		portStr = strconv.FormatUint(uint64(port), 10)
	}
	addr := net.JoinHostPort(host, portStr)

	if forcePassword {
		key = ""
	} else if key == "" {
		key = findDefaultKey()
	}

	// 準備身份驗證方法
	var authMethods []ssh.AuthMethod
	if key != "" {
		// 嘗試使用公鑰認證 (如果 PrivateKey 欄位非空)
		keyPath, err := expandPath(key)
		if err == nil {
			signer, err := signerFromKeyFile(keyPath)
			if err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	clientConfig := &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,                 // 預設先嘗試 PublicKeys 認證
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 實際生產環境中應使用已知主機金鑰驗證
	}

	// 嘗試連線
	client, err := ssh.Dial("tcp", addr, clientConfig)

	// 5. 如果連線失敗, 且錯誤提示是權限不足 (可能需要密碼), 則提示使用者輸入密碼
	if err != nil && strings.Contains(err.Error(), "unable to authenticate") {
		fmt.Printf("使用者密碼: ")

		// 安全地從終端機讀取密碼, 不顯示回顯
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // 讀取完畢後換行

		if err != nil {
			return nil, fmt.Errorf("密碼錯誤: %w", err)
		}

		password := string(passwordBytes)

		// 使用密碼重新配置 Auth
		clientConfig.Auth = []ssh.AuthMethod{
			ssh.Password(password),
		}

		// 使用密碼認證再次嘗試連線
		client, err = ssh.Dial("tcp", addr, clientConfig)
		if err != nil {
			return nil, fmt.Errorf("密碼認證連線失敗: %w", err)
		}

		return client, nil

	} else if err != nil {
		// 其他連線錯誤 (如主機找不到、連線逾時等)
		return nil, fmt.Errorf("SSH 連線失敗: %w", err)
	}

	return client, nil
}

// 展開路徑中的 ~ (例如 ~/.ssh/id_rsa)
func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, path[1:]), nil
	}
	return path, nil
}

// 檢查預設路徑中是否存在私鑰檔案，並傳回第一個找到的路徑
func findDefaultKey() string {
	curOS := runtime.GOOS
	if curOS == "linux" || curOS == "darwin" || curOS == "windows" {
		for _, path := range defaultPrivateKeys {
			keyPath, err := expandPath(path)
			if err == nil {
				// 檢查檔案是否存在且可讀
				if _, err := os.Stat(keyPath); err == nil {
					return path
				}
			}
		}
	}

	return ""
}

// 從檔案路徑建立 ssh.Signer
func signerFromKeyFile(file string) (ssh.Signer, error) {
	buf, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	// 嘗試解析不帶密碼保護的私鑰
	signer, err := ssh.ParsePrivateKey(buf)
	if err == nil {
		return signer, nil
	}

	// 如果解析失敗 (可能是因為私鑰有密碼保護), 則提示使用者輸入金鑰密碼
	// TODO: 測試有密碼保護的 private key
	if strings.Contains(err.Error(), "cannot decode private keys") {
		fmt.Printf("輸入私鑰檔案密碼: ")
		passphraseBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return nil, fmt.Errorf("私鑰密碼輸入失敗: %w", err)
		}
		passphrase := passphraseBytes

		// 使用密碼再次嘗試解析私鑰
		signer, err = ssh.ParsePrivateKeyWithPassphrase(buf, passphrase)
		if err != nil {
			return nil, fmt.Errorf("使用密碼解析私鑰失敗: %w", err)
		}

		return signer, nil
	}

	// 其他解析錯誤
	return nil, fmt.Errorf("無法解析私鑰: %w", err)
}
