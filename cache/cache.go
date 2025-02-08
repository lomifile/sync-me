package cache

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"os"

	"github.com/google/uuid"
	"github.com/lomifile/sync-me/tree"
	"github.com/pkg/sftp"
)

func BuildNewCache(localSrc, src, rootName string, client *sftp.Client) *tree.Node {
	root := tree.NewNode(localSrc, rootName, true)

	files, _ := os.ReadDir(localSrc)

	for _, file := range files {
		if file.IsDir() {
			tree.Insert(localSrc+file.Name(), uuid.NewString(), file, root)
		} else {
			info, _ := file.Info()
			root.Files = append(root.Files, tree.Files{
				Name:     file.Name(),
				ByteSize: info.Size(),
				Path:     root.Path,
			})
		}
	}

	err := client.MkdirAll(src + "/.cache")
	if err != nil {
		panic(err)
	}

	bytes := tree.ExportTree(root)
	file, err := client.Create(src + "/.cache/cache.json")
	if err != nil {
		panic(err)
	}

	defer file.Close()
	file.Write(bytes)

	return root
}

func CheckCacheFile(src string, client *sftp.Client) (bool, error) {
	_, err := client.Stat(src + "/.cache")
	if err == nil {
		return true, nil
	}

	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	return false, err
}

func LoadCacheFile(src string, client *sftp.Client) *tree.Node {
	var root tree.Node
	file, err := client.Open(src + "/.cache/cache.json")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	bytes, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	json.Unmarshal(bytes, &root)
	return &root
}

func UpdateCache(root *tree.Node, src string, client *sftp.Client) {
	bytes := tree.ExportTree(root)
	file, err := client.Create(src + "/.cache/cache.json")
	if err != nil {
		panic(err)
	}

	defer file.Close()
	file.Write(bytes)
}
