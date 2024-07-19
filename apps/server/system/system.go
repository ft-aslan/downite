package system

import (
	"os"
	"path/filepath"
)

type SystemEngine struct {
}
type FileSystemNode struct {
	Type string `json:"type" enum:"dir,file"`
	Size int64  `json:"size"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func (engine *SystemEngine) GetFileSystemNodes(targetPath string) ([]FileSystemNode, error) {
	files := make([]FileSystemNode, 0)
	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			files = append(files, FileSystemNode{
				Type: "dir",
				Size: 0,
				Name: info.Name(),
				Path: path,
			})
		} else {
			files = append(files, FileSystemNode{
				Type: "file",
				Size: info.Size(),
				Name: info.Name(),
				Path: path,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
