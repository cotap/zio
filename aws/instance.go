package aws

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"github.com/cotap/zio/ssh"
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

func ExecInstance(session *session.Session, filters []*ec2.Filter, command string, concurrency int) {
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

type instanceInfo struct {
	name      string
	ipAddress string
}

func SSHInstance(session *session.Session, filters []*ec2.Filter) {
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

	var instances []instanceInfo
	for _, res := range resp.Reservations {
		for _, inst := range res.Instances {
			if inst.PrivateIpAddress != nil {

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

				instances = append(instances, instanceInfo{name: name, ipAddress: *ipAddress})
			}
		}
	}

	if len(instances) == 1 {
		ssh.SSH(instances[0].ipAddress)
		return
	}

	fmt.Println("\nMultiple instances found:\n")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetRowLine(false)
	table.SetColumnSeparator("")
	for i, instance := range instances {
		table.Append([]string{
			color.New(color.Bold).Sprint(strconv.Itoa(i + 1)),
			color.New(color.FgGreen).Sprint(instance.name),
			color.New(color.FgBlue, color.Bold).Sprint(instance.ipAddress),
		})
	}
	table.Render()
	fmt.Println("")

	var ipAddress string
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Login to [1]: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			ipAddress = instances[0].ipAddress
			break
		}

		index, err := strconv.Atoi(input)
		if err != nil || index > len(instances) || index < 1 {
			continue
		}

		ipAddress = instances[index-1].ipAddress
		break
	}
	ssh.SSH(ipAddress)
}
