package ssh

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/lomifile/sync-me/tree"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

var PATH_PREFIX = "./"

func NewClient() *sftp.Client {
	config := &ssh.ClientConfig{
		User: "butters",
		Auth: []ssh.AuthMethod{
			ssh.Password("00259641"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", "butters.local:22", config)
	if err != nil {
		panic(err)
	}

	sftp, err := sftp.NewClient(client)
	if err != nil {
		panic(err)
	}

	return sftp
}

func Exists(path string, client *sftp.Client) (bool, error) {
	_, err := client.Stat(path)
	if err == nil {
		return true, nil
	}

	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	return false, err
}

func CreateDir(path string, client *sftp.Client) {
	err := client.MkdirAll(PATH_PREFIX + path)
	if err != nil {
		panic(err)
	}
}

func HandleSendData(client *sftp.Client, tree *tree.Node) {
	if tree.Root {
		exists, err := Exists(PATH_PREFIX+tree.Name, client)
		if err != nil {
			panic(err)
		}

		if !exists {
			fmt.Println("RootDir: ", PATH_PREFIX+tree.Name, "doesn't exists")
			CreateDir(PATH_PREFIX+tree.Name, client)
		} else {
			fmt.Println("RootDir: ", PATH_PREFIX+tree.Name, "exists")
		}

		PATH_PREFIX = "./" + tree.Name + "/"
	}

	for _, file := range tree.Files {
		exists, err := Exists(PATH_PREFIX+tree.Name+"/"+file.Name, client)
		if err != nil {
			panic(err)
		}

		if !exists {
			fmt.Println("Item: ", PATH_PREFIX+tree.Name+"/"+file.Name, "doesn't exists")
			SendFileOverSsh(file, tree, client)
		} else {
			fmt.Println("Item: ", PATH_PREFIX+tree.Name+"/"+file.Name, "already exists")
		}
	}

	for _, child := range tree.Child {
		exists, err := Exists(PATH_PREFIX+child.Name+"/", client)
		if err != nil {
			panic(err)
		}

		if !exists {
			fmt.Println("Dir: ", PATH_PREFIX+tree.Name+"/"+child.Name, "doesn't exists")
			CreateDir(child.Name, client)
		} else {
			fmt.Println("Dir: ", PATH_PREFIX+tree.Name+"/"+child.Name, "already exists")
		}

		HandleSendData(client, &child)
	}
}

func SendFileOverSsh(file tree.Files, node *tree.Node, client *sftp.Client) error {
	src, err := os.Open(file.Path)
	if err != nil {
		panic(err)
	}

	defer src.Close()

	dst, err := client.Create(PATH_PREFIX + node.Name + "/" + file.Name)
	if err != nil {
		panic(err)
	}

	fmt.Println("Sending: ", PATH_PREFIX+node.Name+"/"+file.Name)
	defer dst.Close()

	if _, err := dst.ReadFrom(src); err != nil {
		return err
	}

	return nil
}
