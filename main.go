package main

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/lomifile/sync-me/tree"
)

func main() {
	root := tree.NewNode("../../notes/", "root")

	files, _ := os.ReadDir("../../notes")

	for _, file := range files {
		fmt.Println(file, file.IsDir())
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
