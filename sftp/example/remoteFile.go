package main

import (
	"bufio"
	"flag"
	"github.com/liserjrqlxue/goremote/sftp"
	"github.com/liserjrqlxue/simple-util"
	"golang.org/x/crypto/ssh"
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

	if *action == "upload" {
		sftp.Upload(conn, *srcPath, *destPath)
	} else {
		sftp.Download(conn, *srcPath, *destPath)
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
