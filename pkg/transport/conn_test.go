package transport

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 測試 expandPath 函數
func TestExpandPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Should expand tilde (~)",
			input:    "~/.ssh/id_rsa",
			expected: filepath.Join(homeDir, ".ssh", "id_rsa"),
		},
		{
			name:     "Should not change absolute path",
			input:    "/etc/hosts",
			expected: "/etc/hosts",
		},
		{
			name:     "Should not change relative path",
			input:    "./config.yaml",
			expected: "./config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Windows 上測試絕對路徑時需要處理 C:
			if runtime.GOOS == "windows" && !strings.Contains(tt.input, "~") {
				path, err := expandPath(tt.input)
				assert.NoError(t, err)
				// 這裡可能需要更精確的 Windows 路徑比較，但基本邏輯是正確的
				// 簡單檢查一下開頭是否正確
				if !strings.HasPrefix(tt.input, "~") {
					assert.Equal(t, tt.input, path)
				}
				return
			}

			path, err := expandPath(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, path)
		})
	}
}

// 測試 findDefaultKey 函數
func TestFindDefaultKey(t *testing.T) {
	// 設置一個臨時的 HOME 目錄用於測試
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)        // for Linux/macOS
	os.Setenv("USERPROFILE", tempDir) // for Windows

	// 1. 測試找不到金鑰的情況
	t.Run("No_Key_Found", func(t *testing.T) {
		key := findDefaultKey()
		assert.Empty(t, key, "Expected no key to be found in empty temp directory")
	})

	// 2. 測試找到金鑰的情況 (模擬建立 id_rsa)
	t.Run("Key_Found", func(t *testing.T) {
		// 創建 ~/.ssh 目錄和 id_rsa 檔案
		sshDir := filepath.Join(tempDir, ".ssh")
		err := os.MkdirAll(sshDir, 0700)
		require.NoError(t, err)

		keyPath := filepath.Join(sshDir, "id_rsa")
		err = os.WriteFile(keyPath, []byte("fake private key content"), 0600)
		require.NoError(t, err)

		// 測試 findDefaultKey
		foundKey := findDefaultKey()
		assert.Equal(t, "~/.ssh/id_rsa", foundKey, "Expected to find ~/.ssh/id_rsa")

		// 測試找到優先級更高的 key (id_ed25519)
		edKeyPath := filepath.Join(sshDir, "id_ed25519")
		err = os.WriteFile(edKeyPath, []byte("fake ed25519 key content"), 0600)
		require.NoError(t, err)

		foundKey = findDefaultKey()
		assert.Equal(t, "~/.ssh/id_ed25519", foundKey, "Expected to find id_ed25519 due to higher priority")
	})

	// 恢復環境變數 (重要)
	t.Cleanup(func() {
		os.Unsetenv("HOME")
		os.Unsetenv("USERPROFILE")
	})
}
