package main

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/go-errors/errors"
)

type meta struct {
	ID *string `json:"id"`
}

type workspace struct {
	Data map[string]json.RawMessage `json:"data"`
	Meta meta                       `json:"metadata"`
}

// Workspace holds information about the workspace and all of the json data
type Workspace struct {
	Data     workspace
	Path     string
	FileInfo fs.FileInfo
}

type pathData struct {
	Path string `json:"path"`
}

// type notebookData struct {
// 	Data pathData `json:"data"`
// }

func readWorkspaceFile(file string) (ws workspace, err error) {
	jsonFile, err := os.Open(file)

	if err != nil {
		err = errors.Wrap(err, 0)
		return
	}
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		err = errors.Wrap(err, 0)
		return
	}

	json.Unmarshal(bytes, &ws)

	return
}

func getWorkspaces(workspaceDir string) (workspaces []Workspace, err error) {
	files, err := ioutil.ReadDir(workspaceDir)
	if err != nil {
		err = errors.Wrap(err, 0)
		return
	}

	var ws workspace
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".jupyterlab-workspace") {
			filePath := path.Join(workspaceDir, file.Name())
			ws, err = readWorkspaceFile(filePath)
			if err != nil {
				err = errors.Wrap(err, 0)
				return
			}
			workspaces = append(workspaces, Workspace{
				Data:     ws,
				Path:     filePath,
				FileInfo: file,
			})
		}
	}

	return
}

func getWorkspaceInfo(data map[string]json.RawMessage) (numOpen int, workingDir string) {
	for key, value := range data {
		if strings.HasPrefix(key, "notebook:") {
			numOpen++
		}
		if key == "file-browser-filebrowser:cwd" {
			p := &pathData{}
			json.Unmarshal(value, p)
			workingDir = p.Path
		}
	}
	return
}
