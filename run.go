package main

import (
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"minidocker/cgroups"
	"minidocker/cgroups/subsystems"
	"minidocker/container"
	"minidocker/network"

	log "github.com/sirupsen/logrus"
)

// Run 执行具体 command
/*
这里的Start方法是真正开始执行由NewParentProcess构建好的command的调用，它首先会clone出来一个namespace隔离的
进程，然后在子进程中，调用/proc/self/exe,也就是调用自己，发送init参数，调用我们写的init方法，
去初始化容器的一些资源。
*/
func Run(tty bool, comArray []string, res *subsystems.ResourceConfig, containerName, nw string, portMapping []string) {
	containerID := randStringBytes(container.IDLength)
	if containerName == "" {
		containerName = containerID
	}
	parent, writePipe := container.NewParentProcess(tty)
	if parent == nil {
		log.Errorf("New parent process error.")
		return
	}
	if err := parent.Start(); err != nil {
		log.Errorf("Run parent.Start err:%v.", err)
	}
	// 创建cgroup manager, 并通过调用set和apply设置资源限制并使限制在容器上生效
	cgroupManager := cgroups.NewCgroupManager("minidocker-cgroup")
	defer cgroupManager.Destroy()
	_ = cgroupManager.Set(res)
	_ = cgroupManager.Apply(parent.Process.Pid, res)

	if nw != "" {
		// config container network
		network.Init()
		containerInfo := &container.Info{
			Id:          containerID,
			Pid:         strconv.Itoa(parent.Process.Pid),
			Name:        containerName,
			PortMapping: portMapping,
		}
		if err := network.Connect(nw, containerInfo); err != nil {
			log.Errorf("Error Connect Network %v", err)
			return
		}
	}

	// 子进程创建后才能通过管道来发送参数
	sendInitCommand(comArray, writePipe)
	_ = parent.Wait()
}

// sendInitCommand 通过writePipe将指令发送给子进程
func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("Command : %s.", command)
	_, _ = writePipe.WriteString(command)
	_ = writePipe.Close()
}

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
