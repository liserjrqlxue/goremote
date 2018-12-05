package sftp

import (
	"fmt"
	"github.com/liserjrqlxue/simple-util"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
)

func Upload(conn *ssh.Client, src, dest string) {
	// create new SFTP client
	client, err := sftp.NewClient(conn)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(client)

	srcFile, err := os.Open(src)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(srcFile)

	dstFile, err := client.Create(dest)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(dstFile)

	// copy source file to destination file
	bytes, err := io.Copy(dstFile, srcFile)
	simple_util.CheckErr(err)
	fmt.Printf("%d bytes upload\n", bytes)
}

func Download(conn *ssh.Client, src, dest string) {
	// create new SFTP client
	client, err := sftp.NewClient(conn)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(client)

	srcFile, err := client.Open(src)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(srcFile)

	dstFile, err := os.Create(dest)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(dstFile)

	// copy source file to destination file
	bytes, err := io.Copy(dstFile, srcFile)
	simple_util.CheckErr(err)
	fmt.Printf("%d bytes download\n", bytes)

	// flush in-memory copy
	err = dstFile.Sync()
	simple_util.CheckErr(err)
}
