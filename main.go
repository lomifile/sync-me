package main

import (
	"fmt"
	"reflect"

	"github.com/lomifile/sync-me/cache"
	"github.com/lomifile/sync-me/config"
	"github.com/lomifile/sync-me/ssh"
	"github.com/lomifile/sync-me/tree"
)

func main() {
	config := config.LoadConfig("./config.json")
	var root *tree.Node
	client := ssh.NewClient(config)
	defer client.Close()
	cacheFile, _ := cache.CheckCacheFile(config.ReceiverFilePath, client)
	if cacheFile {
		fmt.Println("-- Loading cache")
		root = cache.LoadCacheFile(config.ReceiverFilePath, client)

		fmt.Println("-- Loading local file tree")
		fileTree := tree.BuildFileTree(config.SyncFilePath, config.RootName, true)

		if reflect.DeepEqual(root, fileTree) {
			fmt.Println("-- No file tree to sync")
			return
		} else {
			root = fileTree
			cache.UpdateCache(root, config.ReceiverFilePath, client)
		}
	} else {
		fmt.Println("-- Building new cache")
		root = cache.BuildNewCache(config.SyncFilePath, config.ReceiverFilePath, config.RootName, client)
	}

	ssh.HandleSendData(client, root)
}
