package aws

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cotap/zio/ssh"
	"github.com/olekukonko/tablewriter"
)

func ListInstance(session *session.Session, environment, role string) {
	svc := ec2.New(session)

	filters := []*ec2.Filter{}

	if role != "" {
		filters = append(filters, &ec2.Filter{
			Name: aws.String("tag:Role"),
			Values: []*string{
				aws.String(role),
			},
		})
	}

	if environment != "" {
		filters = append(filters, &ec2.Filter{
			Name: aws.String("tag:Environment"),
			Values: []*string{
				aws.String(environment),
			},
		})
	}

	var params *ec2.DescribeInstancesInput
	if len(filters) > 0 {
		params = &ec2.DescribeInstancesInput{
			Filters: filters,
		}
	}

	resp, err := svc.DescribeInstances(params)
	if err != nil {
		log.Fatal(err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Instance ID",
		"Environment",
		"Role",
		"Type",
		"AZ",
		"State",
		"IP Address",
		"Key Name",
	})
	for _, res := range resp.Reservations {
		for _, inst := range res.Instances {
			ipAddress := inst.PrivateIpAddress
			if ipAddress == nil {
				ipAddress = aws.String("")
			}

			var environment, role string
			for _, t := range inst.Tags {
				switch *t.Key {
				case "Environment":
					environment = *t.Value
				case "Role":
					role = *t.Value
				default:
				}
			}

			table.Append([]string{
				*inst.InstanceId,
				environment,
				role,
				*inst.InstanceType,
				*inst.Placement.AvailabilityZone,
				*inst.State.Name,
				*ipAddress,
				*inst.KeyName,
			})
		}
	}
	table.Render()
}

func SSHInstance(session *session.Session, environment, command string, concurrency int) {
	svc := ec2.New(session)
	params := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("tag:Environment"),
				Values: []*string{
					aws.String(environment),
				},
			},
		},
	}

	resp, err := svc.DescribeInstances(params)
	if err != nil {
		log.Fatal(err)
	}

	var ipAddresses []string
	for _, res := range resp.Reservations {
		for _, inst := range res.Instances {
			if inst.PrivateIpAddress != nil {
				ipAddresses = append(ipAddresses, *inst.PrivateIpAddress)
			}
		}
	}
	if len(ipAddresses) == 0 {
		log.Fatal("No instances found")
	}

	if command != "" {
		ssh.ExecAll(ipAddresses, command, concurrency)
	} else {
		ssh.SSH(ipAddresses[0])
	}
}
