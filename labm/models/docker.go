package models

import (
	"bytes"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/wy3148/lab-manager/labm/db"
	"github.com/wy3148/lab-manager/labm/util"
	"golang.org/x/net/context"
	"log"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var tokenMatch *regexp.Regexp

type Jupyter struct {
	Url string `json:"url"`
}

type jupyManager struct {
	sync.Mutex
	port int
}

var redisDb *db.RedisCli
var jp *jupyManager

func init() {
	tokenMatch, _ = regexp.Compile(`.+token=(\w+)`)
	redisDb = db.NewRedisClient()
	jp = &jupyManager{}
	jp.port = 10000
	v, err := redisDb.Do("GET", "PORT")
	if err == nil && v != nil {
		port, err := strconv.Atoi(string(v.([]byte)))
		if err == nil {
			jp.port = port
		}
	}
}

func (j *jupyManager) getPort() string {
	j.Lock()
	defer j.Unlock()
	j.port++
	redisDb.Do("SET", "PORT", strconv.Itoa(j.port))
	return strconv.Itoa(j.port)
}

// NewJupyterDocker
func NewJupyterDocker(uid string) (*Jupyter, error) {
	//we don't very this uid exist or not in aiqinet.cn

	user, err := redisDb.GetMap("jp:" + uid)
	if err == nil && len(user) > 0 {
		util.Log.Warning("user %s is already having a running instance %s", uid, user["docker"])
		return nil, fmt.Errorf("user %s is having a running instance %s", uid, user["docker"])
	}

	//create a real docker instance, also get token
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	newPort := jp.getPort()
	util.Log.Notice("using new port on host:%s", newPort)
	config := &container.Config{
		Image: "jupyter/scipy-notebook:2c80cf3537ca",
		ExposedPorts: nat.PortSet{
			nat.Port("8888/tcp"): struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port("8888/tcp"): []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: newPort,
				},
			},
		},
		PublishAllPorts: true,
		Privileged:      false,
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, nil, "jplabtesting_"+uid)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	//wait until something happens, or 5 second timeout
	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	tc := time.Tick(1 * time.Second)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	case <-tc:
	}

	i, err := cli.ContainerLogs(context.Background(), resp.ID, types.ContainerLogsOptions{
		ShowStderr: true,
		ShowStdout: true,
		Timestamps: false,
		Follow:     true,
	})
	if err != nil {
		log.Fatal(err)
	}

	for {
		dat := make([]byte, 4096)
		_, err = i.Read(dat)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		res := tokenMatch.FindStringSubmatch(string(dat))
		if len(res) == 2 {
			util.Log.Debug("New docker conainer %s running successfully", "jplabtesting_"+uid)
			dockerInst := map[string]string{
				"docker": "jplabtesting_" + uid,
				"since":  strconv.FormatInt(time.Now().Unix(), 10),
			}
			redisDb.StoreMap("jp:"+uid, dockerInst)
			util.Log.Debug("got container token value: %s", res[1])
			return &Jupyter{
				Url: "http://138.197.221.253:" + newPort + "/?token=" + res[1],
			}, nil
		}
	}
	return nil, fmt.Errorf("Failed to get lab resource")
}
