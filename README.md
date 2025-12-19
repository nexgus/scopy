# scopy
`winscp` 其實不好用. 如果 Linux/macOS 端的檔名是 UTF8, 在 GUI 內總是變成亂碼. 在 `cmd` 視窗內使用 `scp` 指令又要考慮編碼方式.

總之在 Windows 就是有一些煩人的事!

## Syntax
```
Usage: scopy <source> <target> [flags]

A simplified scp tool (0.1.0 commit f4f051ff)

Arguments:
  <source>    來源路徑
  <target>    目的路徑

Flags:
  -h, --help                     Show context-sensitive help.
  -x, --excludes=EXCLUDES,...    排除的檔案或目錄模式 (pattern), 可用萬用字元
      --port=22                  SSH 埠號. 預設 22
  -k, --key=STRING               私鑰的檔案位置
      --force-password           強迫使用密碼
  -V, --version                  顯示版本訊息
```

## 安裝
-   Windows
    將 `scopy-0.1.0-windows-amd64.exe` 複製到 `C:\Wwindows` 目錄下, 並將檔名改成 `scopy.exe`
-   Linux
    ```bash
    sudo cp scopy-0.1.0-linux-amd64 /usr/sbin
    sudo chmod +x /usr/sbin/scopy-0.1.0-linux-amd64
    sudo ln -s /usr/sbin/scopy-0.1.0-linux-amd64 /usr/sbin/scopy
    ```
-   macOS
    ```bash
    sudo cp scopy-0.1.0-darwin-arm64 /usr/sbin
    sudo chmod +x /usr/sbin/scopy-0.1.0-darwin-arm64
    sudo ln -s /usr/sbin/scopy-0.1.0-darwin-arm64 /usr/sbin/scopy
    ```
