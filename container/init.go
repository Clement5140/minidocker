package container

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// RunContainerInitProcess 启动容器的init进程
/*
这里的init函数是在容器内部执行的，也就是说，代码执行到这里后，容器所在的进程其实就已经创建出来了，
这是本容器执行的第一一个进程。
使用mount先去挂载proc文件系统，以便后面通过ps等系统命令去查看当前进程资源的情况。
*/
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if len(cmdArray) == 0 {
		return errors.New("Run container get user command error, cmdArray is nil.")
	}
	setUpMount()
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("Exec loop path error %v.", err)
		return err
	}
	log.Infof("Find path %s", path)
	if err = syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf("RunContainerInitProcess exec :" + err.Error())
	}
	return nil
}

const fdIndex = 3

func readUserCommand() []string {
	pipe := os.NewFile(uintptr(fdIndex), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("Init read pipe error %v.", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func setUpMount() {
	// mount proc
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	err := syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	if err != nil {
		log.Errorf("mount proc error %v", err)
	}
}
