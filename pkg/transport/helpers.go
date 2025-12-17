package transport

import (
	"fmt"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func ToSFTP(client *ssh.Client) (*sftp.Client, error) {
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("無法建立 SFTP 客戶端: %w", err)
	}

	return sftpClient, nil
}
