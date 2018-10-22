package models

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"lab-manager/labm/db"
	"lab-manager/labm/util"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"

	"golang.org/x/net/context"
	"strconv"
	"sync"
)

type Jupyter struct {
	Url string `json:"url"`
}

type jupyManager struct {
	sync.Mutex
	port int
}

var dbInst db.Store
var jp *jupyManager

func init() {
	dbInst = db.NewDb()
	jp = &jupyManager{}
	jp.port = 10000
}

func (j *jupyManager) getPort() string {
	j.Lock()
	defer j.Unlock()
	j.port++
	return strconv.Itoa(j.port)
}

// NewJupyterDocker
func NewJupyterDocker(uid string) (*Jupyter, error) {
	//we don't very this uid exist or not in aiqinet.cn
	pyter, err := dbInst.Get(uid)
	if err == nil && len(pyter) > 0 {
		util.Log.Warning("user %s is having a running instance %s", uid, pyter)
		return nil, fmt.Errorf("user %s is having a running instance %s", uid, pyter)
	}

	//create a real docker instance, also get token
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	newPort := jp.getPort() + "/tcp"
	util.Log.Notice("using new port %s", newPort)

	config := &container.Config{
		Image: "jupyter/scipy-notebook:2c80cf3537ca",
		ExposedPorts: nat.PortSet{
			nat.Port(newPort): struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			"4140/tcp": []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: "4140",
				},
			},
		},
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return nil, nil
}
