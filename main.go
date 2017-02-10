package main

import (
	"log"
	"os"

	cli "github.com/jawher/mow.cli"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	zaws "github.com/cotap/zio/aws"
)

var AwsConfig *aws.Config
var AwsSession *session.Session

func main() {
	zio := cli.App("zio", "Manage zinc.io infrastructure")
	zio.Version("v version", "zio 1.0.0")

	region := zio.String(cli.StringOpt{
		Name:   "r region",
		Value:  "us-east-1",
		Desc:   "AWS region",
		EnvVar: "AWS_REGION",
	})

	zio.Spec = "[-r=<region>]"

	zio.Before = func() {
		var err error
		AwsConfig = &aws.Config{Region: aws.String(*region)}
		AwsSession, err = session.NewSession(AwsConfig)
		if err != nil {
			log.Fatal(err)
		}
	}

	zio.Command("instance i", "EC2 Instances", func(cmd *cli.Cmd) {
		tag := cmd.StringOpt("t tag", "", "Filter by tag")
		stack := cmd.StringOpt("s stack", "", "Filter by Cloudformation stack")

		cmd.Spec = "[-t=<Tag:Value> | -s=<stack>]"

		cmd.Command("list ls", "list instances", func(cmd *cli.Cmd) {
			cmd.Action = func() {
				zaws.ListInstance(AwsSession, zaws.Filter(*tag, *stack))
				cli.Exit(0)
			}
		})

		cmd.Command("ssh", "Start SSH session", func(cmd *cli.Cmd) {
			concurrency := cmd.IntOpt("c concurrency", 5, "Concurrency")
			command := cmd.StringArg("CMD", "", "Command to execute")

			cmd.Spec = "[-c] [CMD]"
			cmd.Action = func() {
				zaws.SSHInstance(AwsSession, zaws.Filter(*tag, *stack), *command, *concurrency)
				cli.Exit(0)
			}
		})

		cmd.Command("reserved", "Enumerate EC2 reserved instance status", func(cmd *cli.Cmd) {
			cmd.Action = func() {
				zaws.ReservedAnalysis(AwsSession)
				cli.Exit(0)
			}
		})
	})

	zio.Run(os.Args)
}
