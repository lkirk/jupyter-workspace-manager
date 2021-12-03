package main

import (
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/go-errors/errors"
)

//go:embed "index.html"
var indexTemplateString string
var indexTemplate = template.Must(template.New("index").Parse(indexTemplateString))

//go:embed "error.html"
var errorTemplateString string
var errorTemplate = template.Must(template.New("error").Parse(errorTemplateString))

func respondTemplate(w http.ResponseWriter, t *template.Template, data interface{}) error {
	return t.Execute(w, data)
}

func respondErrorTemplate(w http.ResponseWriter, err error) {
	// TODO: rethink this interface a bit
	respondTemplate(w, errorTemplate, err.(*errors.Error).ErrorStack())
}

func gatherTableData(workspaces []Workspace) (td TableData) {
	for _, workspace := range workspaces {
		numOpen, workingDir := getWorkspaceInfo(workspace.Data.Data)
		td = append(td, tableRow{
			ID:         workspace.Data.Meta.ID,
			WorkingDir: workingDir,
			NumOpenNb:  numOpen,
		})
	}
	return
}

type tableRow struct {
	ID         *string
	WorkingDir string
	NumOpenNb  int
}

// TableData stores data for the table displayed on the main page
type TableData []tableRow

func mainHandlerWrapper(workspaceDir string, notebookIP string, w http.ResponseWriter, r *http.Request) {
	workspaces, err := getWorkspaces(workspaceDir)
	if err != nil {
		respondErrorTemplate(w, errors.Wrap(err, 0))
		return
	}
	instances, err := getRunningJupyterURLsForCurrentUser(notebookIP)
	if err != nil {
		respondErrorTemplate(w, errors.Wrap(err, 0))
		return
	}
	respondTemplate(w, indexTemplate, struct {
		TableData TableData
		Instances map[int32]string
	}{TableData: gatherTableData(workspaces), Instances: instances})
}

func getWorkspacesToRemove(formData string) (workspacesToRemove map[string]struct{}) {
	// make map w/ empty values, set-like object
	workspacesToRemove = make(map[string]struct{})
	if formData == "" {
		return // TODO: log this?
	}
	for _, workspace := range strings.Split(formData, ",") {
		workspacesToRemove[workspace] = struct{}{}
	}
	return
}

func removeWorkspaceHandlerWrapper(workspaceDir string, workspaceTrashDir string, w http.ResponseWriter, r *http.Request) {
	workspacesToRemove := getWorkspacesToRemove(r.FormValue("workspaces"))
	workspaces, err := getWorkspaces(workspaceDir)
	if err != nil {
		// respondErrorTemplate(w, errors.Wrap(err, 0))  respond json
		return
	}

	err = os.MkdirAll(workspaceTrashDir, os.ModePerm)
	if err != nil {
		writeErrorJSON(w, err)
		return
	}

	removed := []string{}
	for _, workspace := range workspaces {
		if _, ok := workspacesToRemove[*workspace.Data.Meta.ID]; ok {
			workspaceTrashPath := path.Join(workspaceTrashDir, workspace.FileInfo.Name())
			log.Printf("moving %s => %s\n", workspace.Path, workspaceTrashPath)
			err = os.Rename(workspace.Path, workspaceTrashPath)
			if err != nil {
				writeErrorJSON(w, err)
			}
			removed = append(removed, *workspace.Data.Meta.ID)
		}
	}

	respondJSON(w, apiResponse{Result: removed})
}

type apiResponse struct {
	Result interface{} `json:"result"`
	Error  *string     `json:"error"`
}

// Write json payload to response writer, setting the proper header
func writeJSON(w http.ResponseWriter, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(payload)
}

// Write json error payload to response writer
func writeErrorJSON(w http.ResponseWriter, err error) {
	response := apiResponse{}
	errMsg := err.Error()
	response.Error = &errMsg
	writeJSON(w, response)
}

// Create json response, where an error is written if there is trouble
// serializing the json response. Otherwise, write the json response
func respondJSON(w http.ResponseWriter, payload interface{}) {
	if err := writeJSON(w, payload); err != nil {
		writeErrorJSON(w, err)
	}
}
