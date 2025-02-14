package tree

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
)

type Files struct {
	Name     string `json:"name"`
	ByteSize int64  `json:"size"`
	Path     string `json:"path"`
}

type Node struct {
	Name  string  `json:"name"`
	Path  string  `json:"path"`
	Files []Files `json:"files"`
	Size  int64   `json:"size"`
	Child []Node  `json:"children"`
	Root  bool    `json:"root"`
}

func BuildFileTree(filePath string, name string, structure ...any) *Node {
	root := NewNode(filePath, name, true)

	files, _ := os.ReadDir(filePath)

	for _, file := range files {
		if file.IsDir() {
			Insert(filePath+file.Name(), uuid.NewString(), file, root)
		} else {
			info, _ := file.Info()
			root.Files = append(root.Files, Files{
				Name:     file.Name(),
				ByteSize: info.Size(),
				Path:     filePath + info.Name(),
			})
		}
	}

	return root
}

func NewNode(filePath string, name string, structure ...any) *Node {
	isRoot := structure[0]
	node := &Node{
		Name:  name,
		Path:  filePath,
		Files: make([]Files, 0),
		Child: make([]Node, 0),
		Root:  isRoot.(bool),
	}

	return node
}

func Insert(path string, id string, file os.DirEntry, node *Node) {
	newNode := NewNode(path, file.Name(), false)
	BuildStructure(path, newNode)
	node.Child = append(node.Child, *newNode)
}

func BuildStructure(path string, node *Node) {
	dir, _ := os.ReadDir(path)

	for _, item := range dir {
		if item.IsDir() {
			Insert(path+"/"+item.Name(), uuid.NewString(), item, node)
		} else {
			info, _ := item.Info()
			fileData := &Files{
				Name:     info.Name(),
				ByteSize: info.Size(),
				Path:     path + "/" + info.Name(),
			}
			node.Files = append(node.Files, *fileData)
		}
	}
}

func PrintTree(node *Node, indent int) {
	prefix := strings.Repeat(" ", indent)
	fmt.Println(prefix + node.Name + "/")

	for _, file := range node.Files {
		fmt.Println(prefix + "  " + file.Name)
	}

	for _, child := range node.Child {
		PrintTree(&child, indent+2)
	}
}

func ExportTree(root *Node) []byte {
	jsonBytes, err := json.Marshal(root)
	if err != nil {
		panic(err)
	}

	return jsonBytes
}
