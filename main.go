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

	zio.Before = func() {
		var err error
		AwsConfig = &aws.Config{Region: aws.String(*region)}
		AwsSession, err = session.NewSession(AwsConfig)
		if err != nil {
			log.Fatal(err)
		}
	}

	zio.Command("instance instances i", "EC2 Instances", func(cmd *cli.Cmd) {
		cmd.Command("list ls", "list instances", func(cmd *cli.Cmd) {
			environment := cmd.StringOpt("e env", "", "Filter by environment")
			role := cmd.StringOpt("r role", "", "Filter by role")

			cmd.Action = func() {
				zaws.ListInstance(AwsSession, *environment, *role)
				cli.Exit(0)
			}
		})

		cmd.Command("ssh", "Start SSH session", func(cmd *cli.Cmd) {
			environment := cmd.StringOpt("e env", "", "Filter by environment")

			cmd.Action = func() {
				zaws.SSHInstance(AwsSession, *environment)
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
