package main

import (
	"encoding/json"
	"os"

	"github.com/google/uuid"
	"github.com/lomifile/sync-me/ssh"
	"github.com/lomifile/sync-me/tree"
)

func main() {
	var root *tree.Node
	if _, err := os.Stat("./cache.json"); err != nil {
		bytes, err := os.ReadFile("./cache.json")
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

		bytes := tree.ExportTree(root)
		file, err := os.Create("./cache.json")
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
