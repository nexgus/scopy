package transport

import "testing"

func TestParseScpCli(t *testing.T) {
	// 1. 定義測試案例結構
	type testCase struct {
		input    string
		expected ScpInfo
		name     string
	}

	// 2. 定義測試案例列表
	cases := []testCase{
		{
			name:  "1. Full Remote Path with Username",
			input: "user@server.example.com:/home/user/file.txt",
			expected: ScpInfo{
				Username: "user",
				Address:  "server.example.com",
				Path:     "/home/user/file.txt",
			},
		},
		{
			name:  "2. Remote Path without Username",
			input: "server.example.com:/var/log/app.log",
			expected: ScpInfo{
				Username: "",
				Address:  "server.example.com",
				Path:     "/var/log/app.log",
			},
		},
		{
			name:  "3. Remote Path with IP Address",
			input: "root@192.168.1.1:~/backup/data.sql",
			expected: ScpInfo{
				Username: "root",
				Address:  "192.168.1.1",
				Path:     "~/backup/data.sql",
			},
		},
		{
			name:  "4. Remote Path with IP and no Username",
			input: "10.0.0.5:/mnt/share/",
			expected: ScpInfo{
				Username: "",
				Address:  "10.0.0.5",
				Path:     "/mnt/share/",
			},
		},
		{
			name:  "5. Local Absolute Path",
			input: "/etc/hosts",
			expected: ScpInfo{
				Username: "",
				Address:  "",
				Path:     "/etc/hosts",
			},
		},
		{
			name:  "6. Local Relative Path",
			input: "local/data.zip",
			expected: ScpInfo{
				Username: "",
				Address:  "",
				Path:     "local/data.zip",
			},
		},
		{
			name:  "7. Path with Hyphens and Underscores",
			input: "my-user_1@dev-server.cloud:path_with-dash.log",
			expected: ScpInfo{
				Username: "my-user_1",
				Address:  "dev-server.cloud",
				Path:     "path_with-dash.log",
			},
		},
		{
			name:  "8. Only Address, Path is empty (Edge Case)",
			input: "host:",
			expected: ScpInfo{
				Username: "",
				Address:  "host",
				Path:     "", // scp 允許路徑為空，代表主目錄
			},
		},
		{
			name:  "9. Only Username and Address, Path is empty (Edge Case)",
			input: "user@host:",
			expected: ScpInfo{
				Username: "user",
				Address:  "host",
				Path:     "",
			},
		},
		{
			name:  "10. Win Remote Path with Drive Letter",
			input: "winserver:C:\\path\\to\\file.txt",
			expected: ScpInfo{
				Username: "",
				Address:  "winserver",
				Path:     "C:\\path\\to\\file.txt",
			},
		},
		{
			name:  "11. Win Remote Path with User and Forward Slashes",
			input: "admin@server.local:/D:/data/backup.db",
			expected: ScpInfo{
				Username: "admin",
				Address:  "server.local",
				Path:     "/D:/data/backup.db",
			},
		},
		{
			name:  "12. Local Win Absolute Path",
			input: "E:\\Program Files (x86)\\app.exe",
			expected: ScpInfo{
				Username: "",
				Address:  "",
				Path:     "E:\\Program Files (x86)\\app.exe",
			},
		},
		{
			name:  "13. Local Win UNC Path (Double Backslash)",
			input: "\\\\unc-server\\share\\document.pdf",
			expected: ScpInfo{
				Username: "",
				Address:  "",
				Path:     "\\\\unc-server\\share\\document.pdf",
			},
		},
		{
			name:  "14. Remote Win UNC Path (Double Forward Slash)",
			input: "user@host.net://sharename/folder/",
			expected: ScpInfo{
				Username: "user",
				Address:  "host.net",
				Path:     "//sharename/folder/",
			},
		},
	}

	// 3. 迭代執行測試案例
	for _, tc := range cases {
		// 使用 t.Run 來執行子測試，方便隔離和報告結果
		t.Run(tc.name, func(t *testing.T) {
			actual := ParseScpCli(tc.input)

			// 檢查 Username 是否匹配
			if actual.Username != tc.expected.Username {
				t.Errorf("Username Mismatch for input %s.\nExpected: %s, Got: %s",
					tc.input, tc.expected.Username, actual.Username)
			}

			// 檢查 Address 是否匹配
			if actual.Address != tc.expected.Address {
				t.Errorf("Address Mismatch for input %s.\nExpected: %s, Got: %s",
					tc.input, tc.expected.Address, actual.Address)
			}

			// 檢查 Path 是否匹配
			if actual.Path != tc.expected.Path {
				t.Errorf("Path Mismatch for input %s.\nExpected: %s, Got: %s",
					tc.input, tc.expected.Path, actual.Path)
			}
		})
	}
}
