package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jawher/mow.cli"
	"github.com/mattn/go-shellwords"
	"github.com/vaughan0/go-ini"

	zaws "github.com/cotap/zio/aws"
)

var AwsConfig *aws.Config
var AwsSession *session.Session

func main() {
	zio := cli.App("zio", "Manage zinc.io infrastructure")
	zio.Version("v version", "zio 1.0.0")

	config, _ := loadConfig()
	if config != nil {
		applyConfig(config)
	}

	region := zio.String(cli.StringArg{
		Name:   "REGION",
		Value:  "us-east-1",
		Desc:   "AWS region",
		EnvVar: "AWS_REGION",
	})

	zio.Spec = "[REGION]"

	zio.Before = func() {
		var err error
		AwsConfig = &aws.Config{Region: aws.String(*region)}
		AwsSession, err = session.NewSession(AwsConfig)
		if err != nil {
			log.Fatal(err)
		}
	}

	zio.Command("instance i", "EC2 Instances", func(cmd *cli.Cmd) {
		var (
			instances []zaws.InstanceInfo
			query     = cmd.StringArg("QUERY", "", "Fuzzy search query")
			stack     = cmd.StringOpt("s stack", "", "Stack")
			tag       = cmd.StringOpt("t tag", "", "Tag")
			ids       = cmd.StringsOpt("id", make([]string, 0), "Instance ID")
			ips       = cmd.StringsOpt("ip", make([]string, 0), "IP")
		)

		cmd.Before = func() {
			var err error
			instances, err = zaws.GetInstances(AwsSession, &zaws.InstanceQuery{*query, *stack, *tag, *ids, *ips})
			if err != nil {
				log.Fatal(err)
			}

			if len(instances) == 0 {
				fmt.Println("No instances found for query")
				cli.Exit(0)
			}
		}

		cmd.Spec = "[QUERY] [--stack=<stack name>] [--tag=<Name:Value>] [--id=<id>]... [--ip=<ip>]..."
		cmd.Action = func() {
			zaws.ListInstance(instances)
			cli.Exit(0)
		}

		cmd.Command("exec e", "Execute command on instance", func(cmd *cli.Cmd) {
			var (
				command     = cmd.StringArg("CMD", "", "Command to execute")
				concurrency = cmd.IntOpt("c concurrency", 2, "Concurrency")
			)
			cmd.Spec = "CMD [-c]"
			cmd.Action = func() {
				zaws.ExecInstance(instances, *command, *concurrency)
				cli.Exit(0)
			}
		})

		cmd.Command("ssh", "SSH into an instance", func(cmd *cli.Cmd) {
			var (
				command = cmd.StringArg("EXEC", "", "Command to attach to the session")
			)
			cmd.Spec = "[EXEC]"
			cmd.Action = func() {
				zaws.SSHInstance(instances, *command)
				cli.Exit(0)
			}
		})
	})

	zio.Command("reserved", "Enumerate EC2 reserved instance status", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			zaws.ReservedAnalysis(AwsSession)
			cli.Exit(0)
		}
	})

	zio.Run(os.Args)
}

func applyConfig(config ini.File) {
	replaceAliases(config.Section("alias"))
}

func loadConfig() (ini.File, error) {
	config := make(ini.File)

	if err := config.LoadFile(".ziorc"); err == nil {
		return config, nil
	}

	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	if err = config.LoadFile(usr.HomeDir + "/.ziorc"); err == nil {
		return config, nil
	}

	return nil, err
}

func replaceAliases(aliases map[string]string) {
	if len(aliases) == 0 {
		return
	}

	aliased := false
	args := make([]string, 0)
	for _, arg := range os.Args {
		replacement, ok := aliases[arg]
		if !ok {
			args = append(args, arg)
			continue
		}

		newArgs, err := shellwords.Parse(replacement)
		if err != nil {
			log.Fatal(err)
		}

		aliased = true
		args = append(args, newArgs...)
	}

	if !aliased {
		return
	}

	os.Args = args
	fmt.Printf("zio %v\n", strings.Join(os.Args[1:], " "))

}
