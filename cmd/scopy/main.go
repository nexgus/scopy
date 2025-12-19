package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tp "scopy/pkg/transport"
	"scopy/pkg/util"

	"github.com/alecthomas/kong"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var args struct {
	Source        string           `arg:"" name:"source" help:"來源路徑"`
	Target        string           `arg:"" name:"target" help:"目的路徑"`
	Exclude       []string         `short:"x" help:"排除的檔案或目錄模式 (pattern), 可用萬用字元"`
	Port          uint16           `default:"22" help:"SSH 埠號. 預設 22"`
	Key           string           `short:"k" help:"私鑰的檔案位置"`
	ForcePassword bool             `help:"強迫使用密碼"`
	Version       kong.VersionFlag `short:"V" help:"顯示版本訊息"`
}

func main() {
	kong.Parse(
		&args,
		kong.Name(filepath.Base(os.Args[0])),
		kong.Description(fmt.Sprintf("A simplified scp tool (%s commit %s)", VersionString, GitCommitHash)),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		kong.Vars{"version": fmt.Sprintf("%s (commit %s)", VersionString, GitCommitHash)},
	)

	srcInfo := tp.ParseScpCli(args.Source)
	dstInfo := tp.ParseScpCli(args.Target)

	var (
		remote     *ssh.Client
		err        error
		isDownload bool
	)
	if len(srcInfo.Username) > 0 && len(srcInfo.Address) > 0 {
		remote, err = tp.Connect(srcInfo.Address, args.Port, srcInfo.Username, args.Key, args.ForcePassword)
		if err != nil {
			exit("連線至 %s 時發生錯誤: %v.", srcInfo.Address, err)
		} else {
			defer remote.Close()
			isDownload = true
		}
	} else if len(dstInfo.Username) > 0 && len(dstInfo.Address) > 0 {
		remote, err = tp.Connect(dstInfo.Address, args.Port, dstInfo.Username, args.Key, args.ForcePassword)
		if err != nil {
			exit("連線至 %s 時發生錯誤: %s.", dstInfo.Address, err)
		} else {
			defer remote.Close()
		}
	} else {
		exit("沒有或不正確地設定遠端.")
	}

	if client, err := sftp.NewClient(remote); err != nil {
		exit("無法建立 SFTP: %s.", err)
	} else {
		defer client.Close()

		path, err := client.RealPath(".")
		if err != nil {
			exit("嘗試檢查遠端時發生錯誤: %s.", err)
		}
		remoteSep := "\\"
		if strings.HasPrefix(path, "/") {
			remoteSep = "/"
		}

		if isDownload {
			if _, err := client.Stat(srcInfo.Path); err != nil {
				if os.IsNotExist(err) {
					exit("遠端路徑 %s 不存在.", srcInfo.Path)
				} else {
					exit("取得遠端路徑資訊時發生錯誤: %s.", err)
				}
			} else {
				if err := tp.Download(client, srcInfo.Path, dstInfo.Path, args.Exclude, remoteSep); err != nil {
					exit("下載時發生錯誤: %s.", err)
				}
			}
		} else {
			if !util.PathExists(srcInfo.Path) {
				exit("本地路徑 %s 不存在.", srcInfo.Path)
			}

			if err := tp.Upload(client, dstInfo.Path, srcInfo.Path, args.Exclude, remoteSep); err != nil {
				exit("上傳時發生錯誤: %s.", err)
			}
		}
	}
}

func exit(format string, a ...any) {
	fmt.Printf(format, a...)
	os.Exit(1)
}
