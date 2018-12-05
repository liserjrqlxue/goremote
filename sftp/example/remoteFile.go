package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/liserjrqlxue/simple-util"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

var (
	action = flag.String(
		"action",
		"download",
		"action type upload/download",
	)
	User = flag.String(
		"user",
		"john",
		"remote user",
	)
	host = flag.String(
		"host",
		"example.com",
		"remote host",
	)
	password = flag.String(
		"password",
		"",
		"password to user@remote",
	)
	key = flag.String(
		"key",
		"/.ssh/id_rsa",
		"ssh key path",
	)
	port = flag.String(
		"port",
		":22",
		"port to remote",
	)
	srcPath = flag.String(
		"src",
		"",
		"source file path",
	)
	destPath = flag.String(
		"dest",
		"",
		"destination file path",
	)
)

func main() {
	flag.Parse()

	if *srcPath == "" || *destPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	// get host public key
	hostKey := getHostKey(*host)

	config := &ssh.ClientConfig{
		User: *User,
		Auth: []ssh.AuthMethod{
			ssh.Password(*password),
		},
		// HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}

	// connect
	conn, err := ssh.Dial("tcp", *host+*port, config)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(conn)

	// create new SFTP client
	client, err := sftp.NewClient(conn)
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(client)

	if *action == "upload" {
		// create destination file
		dstFile, err := client.Create(*destPath)
		simple_util.CheckErr(err)
		defer simple_util.DeferClose(dstFile)

		// open source file
		srcFile, err := os.Open(*srcPath)
		simple_util.CheckErr(err)

		// copy source file to destination file
		bytes, err := io.Copy(dstFile, srcFile)
		simple_util.CheckErr(err)
		fmt.Printf("%d bytes copied\n", bytes)
	} else {
		// create destination file
		dstFile, err := os.Create(*destPath)
		simple_util.CheckErr(err)
		defer simple_util.DeferClose(dstFile)

		// open source file
		srcFile, err := client.Open(*srcPath)
		simple_util.CheckErr(err)

		// copy source file to destination file
		bytes, err := io.Copy(dstFile, srcFile)
		simple_util.CheckErr(err)
		fmt.Printf("%d bytes copied\n", bytes)

		// flush in-memory copy
		err = dstFile.Sync()
		simple_util.CheckErr(err)
	}
}

func getHostKey(host string) ssh.PublicKey {
	// parse OpenSSH known_hosts file
	// ssh or use ssh-keyscan to get initial key

	usr, err := user.Current()
	simple_util.CheckErr(err)

	file, err := os.Open(filepath.Join(usr.HomeDir, ".ssh", "known_hosts"))
	simple_util.CheckErr(err)
	defer simple_util.DeferClose(file)

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				log.Fatalf("error parsing %q: %v", fields[2], err)
			}
			break
		}
	}

	if hostKey == nil {
		log.Fatalf("no hostkey found for %s", host)
	}

	return hostKey
}
