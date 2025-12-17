package transport

import (
	"regexp"
)

// ScpInfo 結構體用於儲存解析後的資訊
type ScpInfo struct {
	Username string
	Address  string
	Path     string
}

// ParseScpCli 解析 scp 格式的字串 [user@]host:path. 允許 path 部分包含磁碟機代號 C: 或 UNC 雙斜線 //.
func ParseScpCli(cli string) ScpInfo {
	// 1. ^ - 字串開頭
	// 2. (?P<username>[a-zA-Z0-9_-]+@)? - 可選的使用者名稱 + @
	// 3. (?P<address>[a-zA-Z0-9._-]+) - 主機位址 (至少需要一個字元)
	// 4. : - 遠端路徑分隔符（這是判斷是否為遠端路徑的關鍵）
	// 5. (?P<path>.*) - 剩餘的部分視為路徑
	// 6. $ - 字串結尾
	regex := regexp.MustCompile(`^((?P<username>[a-zA-Z0-9_-]+)@)?(?P<address>[a-zA-Z0-9._-]+):(?P<path>.*)$`)
	match := regex.FindStringSubmatch(cli)

	info := ScpInfo{}
	if match == nil {
		info.Username = ""
		info.Address = ""
		info.Path = cli
	} else {
		names := regex.SubexpNames()
		for idx, name := range names {
			if idx == 0 || match[idx] == "" {
				continue
			}

			switch name {
			case "username":
				info.Username = match[idx]
			case "address":
				info.Address = match[idx]
			case "path":
				info.Path = match[idx]
			}
		}

		if len(info.Address) == 1 && info.Username == "" {
			// 由於 Address 是一個單字母, 且沒有 Username,
			// 我們判定它是磁碟機代號
			info.Username = ""
			info.Address = ""
			info.Path = cli
		}
	}

	return info
}
