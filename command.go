package main

import (
	"fmt"

	"minidocker/cgroups/subsystems"
	"minidocker/container"
	"minidocker/network"

	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var runCommand = cli.Command{
	Name: "run",
	Usage: `Create a container with namespace and cgroups limit.
			minidocker run -it [command]`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "it",
			Usage: "enable tty",
		},
		cli.StringFlag{
			Name:  "mem",
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "cpu",
			Usage: "cpu quota",
		},
		cli.StringFlag{
			Name:  "cpuset",
			Usage: "cpuset limit",
		},
		cli.StringFlag{
			Name:  "net",
			Usage: "container network",
		},
		cli.StringSliceFlag{
			Name:  "p",
			Usage: "port mapping",
		},
	},
	/*
		这里是run命令执行的真正函数。
		1.判断参数是否包含command
		2.获取用户指定的command
		3.调用Run function去准备启动容器:
	*/
	Action: func(context *cli.Context) error {
		if len(context.Args()) < 1 {
			return fmt.Errorf("Missing container command.")
		}

		var cmdArray []string
		for _, arg := range context.Args() {
			cmdArray = append(cmdArray, arg)
		}

		tty := context.Bool("it")
		resConf := &subsystems.ResourceConfig{
			MemoryLimit: context.String("mem"),
			CpuSet:      context.String("cpuset"),
			CpuCfsQuota: context.Int("cpu"),
		}
		containerName := context.String("name")
		network := context.String("net")
		portMapping := context.StringSlice("p")
		Run(tty, cmdArray, resConf, containerName, network, portMapping)
		return nil
	},
}

var initCommand = cli.Command{
	Name:  "init",
	Usage: "Init container process run user's process in container. Do not call it outside.",
	Action: func(context *cli.Context) error {
		log.Infof("Initiating.")
		err := container.RunContainerInitProcess()
		return err
	},
}

var networkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		{
			Name:  "create",
			Usage: "create a container network",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "driver",
					Usage: "network driver",
				},
				cli.StringFlag{
					Name:  "subnet",
					Usage: "subnet cidr",
				},
			},
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("missing network name")
				}
				err := network.Init()
				if err != nil {
					return fmt.Errorf("init network error: %+v", err)
				}
				err = network.CreateNetwork(context.String("driver"), context.String("subnet"), context.Args()[0])
				if err != nil {
					return fmt.Errorf("create network error: %+v", err)
				}
				return nil
			},
		},
		{
			Name:  "list",
			Usage: "list container network",
			Action: func(context *cli.Context) error {
				err := network.Init()
				if err != nil {
					return err
				}
				network.ListNetwork()
				return nil
			},
		},
		{
			Name:  "remove",
			Usage: "remove container network",
			Action: func(context *cli.Context) error {
				if len(context.Args()) < 1 {
					return fmt.Errorf("missing network name")
				}
				network.Init()
				err := network.DeleteNetwork(context.Args()[0])
				if err != nil {
					return fmt.Errorf("remove network error: %+v", err)
				}
				return nil
			},
		},
	},
}
