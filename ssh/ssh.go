package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/lomifile/sync-me/config"
	"github.com/lomifile/sync-me/tree"
	"github.com/pkg/sftp"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/crypto/ssh"
)

const CHUNK_SIZE = 4096

var PATH_PREFIX = "./"

func RunComparisonOnFiles(src, dest string, client *sftp.Client) bool {
	f1, err := os.Open(src)
	if err != nil {
		panic(err)
	}
	defer f1.Close()

	f2, err := client.Open(dest)
	if err != nil {
		panic(err)
	}
	defer f2.Close()

	buf1 := make([]byte, CHUNK_SIZE)
	buf2 := make([]byte, CHUNK_SIZE)

	for {
		n1, err1 := io.ReadFull(f1, buf1)
		n2, err2 := io.ReadFull(f2, buf2)

		if n1 != n2 {
			return false
		}

		if !bytes.Equal(buf1[:n1], buf2[:n2]) {
			return false
		}

		if err1 == io.EOF && err2 == io.EOF {
			return true
		}

		if err1 == io.EOF || err2 == io.EOF {
			return false
		}

		if err1 != nil && err1 != io.ErrUnexpectedEOF {
			panic(err1)
		}
		if err2 != nil && err2 != io.ErrUnexpectedEOF {
			panic(err2)
		}
	}
}

func NewClient(config *config.Config) *sftp.Client {
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", config.Host, config.Port), &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
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

func BuildFilePath(tree *tree.Node, file *tree.Files) string {
	if tree.Root {
		return PATH_PREFIX + file.Name
	} else {
		return PATH_PREFIX + tree.Name + "/" + file.Name
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
		path := BuildFilePath(tree, &file)

		exists, err := Exists(path, client)
		if err != nil {
			panic(err)
		}

		if !exists {
			fmt.Println("-- [Create]: Item: ", path, "doesn't exists")
			err := SendFileOverSsh(file, path, client)
			if err != nil {
				panic(err)
			}

		} else {
			isSame := RunComparisonOnFiles(file.Path, path, client)
			if !isSame {
				fmt.Println("-- [UPDATE]: Item: ", path, "updating data on server")
				err := SendFileOverSsh(file, path, client)
				if err != nil {
					panic(err)
				}
			}
		}

		fmt.Println("-- [SKIP]: Item: ", path, "already exists")
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

func SendFileOverSsh(file tree.Files, path string, client *sftp.Client) error {
	src, err := os.Open(file.Path)
	if err != nil {
		return fmt.Errorf("unable to open source file: %v", err)
	}
	defer src.Close()

	dst, err := client.Create(path)
	if err != nil {
		return fmt.Errorf("unable to create destination file: %v", err)
	}
	defer dst.Close()

	fileInfo, err := src.Stat()
	if err != nil {
		return fmt.Errorf("unable to get file info: %v", err)
	}

	bar := progressbar.NewOptions64(fileInfo.Size(),
		progressbar.OptionSetDescription(fmt.Sprintf("-- [Send]: %s", path)),
		progressbar.OptionSetWriter(os.Stdout),
		progressbar.OptionSetWidth(50),
		progressbar.OptionSetPredictTime(false),
	)

	_, err = io.Copy(io.MultiWriter(dst, bar), src)
	if err != nil {
		return fmt.Errorf("error while copying file: %v", err)
	}

	bar.Finish()

	fmt.Println()

	return nil
}
