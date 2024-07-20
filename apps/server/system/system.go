package system

import (
	"os"
	"path/filepath"
)

type SystemEngine struct {
}
type FileSystemNode struct {
	Type string `json:"type" enum:"dir,file,parent"`
	Size int64  `json:"size"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func (engine *SystemEngine) GetFileSystemNodes(targetPath string) ([]FileSystemNode, error) {
	files := make([]FileSystemNode, 0)
	if targetPath != "/" {
		files = append(files, FileSystemNode{
			Type: "parent",
			Size: 0,
			Name: ".. [Back]",
			Path: filepath.Dir(targetPath),
		})
	}
	nodes, err := os.ReadDir(targetPath)

	if err != nil {
		return nil, err
	}
	for _, node := range nodes {
		nodeInfo, err := node.Info()
		if err != nil {
			return nil, err
		}
		if node.IsDir() {
			files = append(files, FileSystemNode{
				Type: "dir",
				Size: 0,
				Name: node.Name(),
				Path: filepath.Join(targetPath, nodeInfo.Name()),
			})
		} else {
			files = append(files, FileSystemNode{
				Type: "file",
				Size: nodeInfo.Size(),
				Name: node.Name(),
				Path: filepath.Join(targetPath, nodeInfo.Name()),
			})
		}
	}
	return files, nil
}
