package main

import (
	"fmt"
	"reflect"

	"github.com/lomifile/sync-me/cache"
	"github.com/lomifile/sync-me/ssh"
	"github.com/lomifile/sync-me/tree"
)

var rootPath = "../../notes/"

func main() {
	var root *tree.Node
	client := ssh.NewClient()
	defer client.Close()
	cacheFile, _ := cache.CheckCacheFile("./notes", client)
	if cacheFile {
		fmt.Println("-- Loading cache")
		root = cache.LoadCacheFile("./notes", client)

		fmt.Println("-- Loading local file tree")
		fileTree := tree.BuildFileTree(rootPath, "notes", true)

		if reflect.DeepEqual(root, fileTree) {
			fmt.Println("-- No file tree to sync")
			return
		} else {
			root = fileTree
			cache.UpdateCache(root, "./notes", client)
		}
	} else {
		fmt.Println("-- Building new cache")
		root = cache.BuildNewCache("../../notes/", "./notes", "notes", client)
	}

	ssh.HandleSendData(client, root)
}
