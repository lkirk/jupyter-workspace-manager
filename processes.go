package main

import (
	"fmt"
	"os/user"
	"strconv"

	"github.com/go-errors/errors"
	"github.com/shirou/gopsutil/v3/process"
)

func getRunningJupyterURLsForCurrentUser(notebookIP string) (pidURLMap map[int32]string, err error) {
	pidURLMap = make(map[int32]string)
	currentUser, err := user.Current()
	if err != nil {
		err = errors.Wrap(err, 0)
		return
	}

	currentUserID, err := strconv.Atoi(currentUser.Uid)
	if err != nil {
		err = errors.Wrap(err, 0)
		return
	}

	pids, err := process.Processes()
	if err != nil {
		err = errors.Wrap(err, 0)
		return
	}

	for _, p := range pids {
		// get the process name, verify that it's a jupyter process (by the name)
		procName, err := p.Name()
		if err != nil {
			return pidURLMap, errors.Wrap(err, 0)
		}
		if procName != "jupyter-lab" {
			continue
		}

		// get the username, verify that it's the current user
		uids, err := p.Uids()
		if err != nil {
			return pidURLMap, errors.Wrap(err, 0)
		}

		if int(uids[0]) != currentUserID {
			continue
		}

		// get all tcp connections that are listening and construct their URLs
		conns, err := p.Connections()
		if err != nil {
			return pidURLMap, errors.Wrap(err, 0)
		}
		for _, conn := range conns {
			if conn.Status == "LISTEN" && conn.Family == 2 { // TODO: Family == 2 is linux specific
				var address string
				if notebookIP != "" {
					address = notebookIP
				} else {
					address = conn.Laddr.IP
				}

				pidURLMap[p.Pid] = fmt.Sprintf("%s:%d", address, conn.Laddr.Port)
			}
			// TODO: error if more than one found
		}
	}

	if err != nil {
		err = errors.Wrap(err, 0)
	}
	return
}
