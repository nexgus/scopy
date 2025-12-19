package util

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ReplaceSepWith(src string, sep string) string {
	srcSep := "/"
	if sep == "/" {
		srcSep = "\\"
	}

	return strings.ReplaceAll(src, srcSep, sep)
}

func isExcluded(base string, excludes []string) bool {
	for _, excludePattern := range excludes {
		match, err := filepath.Match(excludePattern, base)
		if err != nil {
			// 如果模式本身有錯誤，應該記錄下來，但這裡為了不中斷流程假設不匹配
			// 實際應用中可能需要更嚴格的錯誤處理
			fmt.Fprintf(os.Stderr, "警告: 排除模式無效 (%s): %v\n", excludePattern, err)
			continue
		}

		if match {
			return true
		}
	}

	return false
}
