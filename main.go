package main

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"

	"github.com/google/uuid"
	"github.com/lomifile/sync-me/ssh"
	"github.com/lomifile/sync-me/tree"
)

var rootPath = "../../notes/"

func CheckCacheFile() (bool, error) {
	_, err := os.Stat(rootPath + ".cache")
	if err == nil {
		return true, nil
	}

	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	return false, err
}

func main() {
	var root *tree.Node
	cache, _ := CheckCacheFile()
	if cache {
		bytes, err := os.ReadFile(rootPath + ".cache/cache.json")
		if err != nil {
			panic(err)
		}
		json.Unmarshal(bytes, root)
	} else {

		root = tree.NewNode("../../notes/", "notes", true)

		files, _ := os.ReadDir("../../notes")

		for _, file := range files {
			if file.IsDir() {
				tree.Insert("../../notes/"+file.Name(), uuid.NewString(), file, root)
			} else {
				info, _ := file.Info()
				root.Files = append(root.Files, tree.Files{
					Name:     file.Name(),
					ByteSize: info.Size(),
					Path:     root.Path,
				})
			}
		}

		err := os.MkdirAll(rootPath+".cache", 0750)
		if err != nil {
			panic(err)
		}

		bytes := tree.ExportTree(root)
		file, err := os.Create(rootPath + ".cache/cache.json")
		if err != nil {
			panic(err)
		}

		defer file.Close()
		file.Write(bytes)
	}

	client := ssh.NewClient()
	defer client.Close()
	ssh.HandleSendData(client, root)
}
