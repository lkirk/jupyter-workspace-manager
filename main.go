package main

import (
	_ "embed"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/urfave/cli/v2"
)

func getWorkspaceDir() string {
	homeDir := os.Getenv("HOME")
	return path.Join(homeDir, ".jupyter", "lab", "workspaces")
}

func getWorkspaceTrashDir() string {
	return path.Join(getWorkspaceDir(), "trash")
}

//go:embed "VERSION"
// Version gets incremented when we tag with the release script
var Version string

func main() {
	app := &cli.App{
		Name:    "jupyter-workspace-manager",
		Usage:   "A tool for managing jupyter workspaces",
		Version: "0.1",
		Action: func(c *cli.Context) error {
			mainHandler := func(w http.ResponseWriter, r *http.Request) {
				mainHandlerWrapper(c.String("workspace-dir"), c.String("notebook-ip"), w, r)
			}
			removeWorkspaceHandler := func(w http.ResponseWriter, r *http.Request) {
				removeWorkspaceHandlerWrapper(c.String("workspace-dir"), c.String("workspace-trash-dir"), w, r)
			}
			http.HandleFunc("/", mainHandler)
			http.HandleFunc("/remove_workspace", removeWorkspaceHandler)
			address := strings.Join([]string{c.String("ip"), c.String("port")}, ":")
			log.Printf("running at %s", address)
			return http.ListenAndServe(address, nil)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "ip",
				Value: "localhost",
				Usage: "IP address to listen on",
			},
			// TODO: https://stackoverflow.com/questions/56336168/golang-check-tcp-port-open
			&cli.StringFlag{
				Name:  "port",
				Value: "8080",
				Usage: "Port to listen on",
			},
			&cli.StringFlag{
				Name:  "workspace-dir",
				Value: getWorkspaceDir(),
				Usage: "Dir to search for workspaces in",
			},
			// TODO: make default relative to other arg?
			&cli.StringFlag{
				Name:  "workspace-trash-dir",
				Value: getWorkspaceTrashDir(),
				Usage: "Trash dir to send workspaces to",
			},
			&cli.StringFlag{
				Name:  "notebook-ip",
				Value: "localhost",
				Usage: "IP address to use for notebook links",
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
