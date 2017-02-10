package aws

import (
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/cotap/zio/ssh"
	"github.com/olekukonko/tablewriter"
)

func Filter(tag, stack string) []*ec2.Filter {
	filters := []*ec2.Filter{}

	if tag != "" {
		tagParts := strings.Split(tag, ":")
		filters = append(filters, &ec2.Filter{
			Name: aws.String("tag:" + tagParts[0]),
			Values: []*string{
				aws.String(tagParts[1]),
			},
		})
	}

	if stack != "" {
		filters = append(filters, &ec2.Filter{
			Name: aws.String("tag:aws:cloudformation:stack-name"),
			Values: []*string{
				aws.String(stack),
			},
		})
	}

	return filters
}

func ListInstance(session *session.Session, filters []*ec2.Filter) {
	svc := ec2.New(session)

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
		"Name",
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

			var name string
			for _, t := range inst.Tags {
				switch *t.Key {
				case "Name":
					name = *t.Value
				default:
				}
			}

			table.Append([]string{
				*inst.InstanceId,
				name,
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

func SSHInstance(session *session.Session, filters []*ec2.Filter, command string, concurrency int) {
	svc := ec2.New(session)

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
