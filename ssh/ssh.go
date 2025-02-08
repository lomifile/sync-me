package ssh

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/lomifile/sync-me/tree"
	"github.com/pkg/sftp"
	"github.com/schollz/progressbar/v3"
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
			fmt.Println("-- [Create]: RootDir: ", PATH_PREFIX+tree.Name, "doesn't exists")
			CreateDir(PATH_PREFIX+tree.Name, client)
		} else {
			fmt.Println("-- [SKIP]: RootDir: ", PATH_PREFIX+tree.Name, "exists")
		}

		PATH_PREFIX = "./" + tree.Name + "/"
	}

	for _, file := range tree.Files {
		exists, err := Exists(PATH_PREFIX+tree.Name+"/"+file.Name, client)
		if err != nil {
			panic(err)
		}

		if !exists {
			fmt.Println("-- [Create]: Item: ", PATH_PREFIX+tree.Name+"/"+file.Name, "doesn't exists")
			SendFileOverSsh(file, tree, client)
		} else {

			current, err := os.Lstat(file.Path)
			if err != nil {
				panic(err)
			}

			if file.ByteSize != current.Size() {
				fmt.Println("-- [UPDATE]: Item: ", PATH_PREFIX+tree.Name+"/"+file.Name, "has changed size from", file.ByteSize, "to", current.Size(), "updating data on server")
			}

			fmt.Println("-- [SKIP]: Item: ", PATH_PREFIX+tree.Name+"/"+file.Name, "already exists")
		}
	}

	for _, child := range tree.Child {
		exists, err := Exists(PATH_PREFIX+child.Name+"/", client)
		if err != nil {
			panic(err)
		}

		if !exists {
			fmt.Println("-- [Create]: Dir: ", PATH_PREFIX+tree.Name+"/"+child.Name, "doesn't exists")
			CreateDir(child.Name, client)
		} else {
			fmt.Println("-- [SKIP]: Dir: ", PATH_PREFIX+tree.Name+"/"+child.Name, "already exists")
		}

		HandleSendData(client, &child)
	}
}

func SendFileOverSsh(file tree.Files, node *tree.Node, client *sftp.Client) error {
	// Open source file
	src, err := os.Open(file.Path)
	if err != nil {
		return fmt.Errorf("unable to open source file: %v", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := client.Create(PATH_PREFIX + node.Name + "/" + file.Name)
	if err != nil {
		return fmt.Errorf("unable to create destination file: %v", err)
	}
	defer dst.Close()

	// Get the file size for progress bar
	fileInfo, err := src.Stat()
	if err != nil {
		return fmt.Errorf("unable to get file info: %v", err)
	}

	// Create the progress bar
	bar := progressbar.NewOptions64(fileInfo.Size(),
		progressbar.OptionSetDescription(fmt.Sprintf("Sending: %s", PATH_PREFIX+node.Name+"/"+file.Name)),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetPredictTime(false),
	)

	// Create a multi-writer to send data to both destination and progress bar
	_, err = io.Copy(io.MultiWriter(dst, bar), src)
	if err != nil {
		return fmt.Errorf("error while copying file: %v", err)
	}

	// After transfer is complete, ensure the progress bar is completed
	bar.Finish()

	fmt.Println()

	return nil
}
