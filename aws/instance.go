package aws

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/cotap/zio/ssh"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

func ListInstance(instances []InstanceInfo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Instance ID",
		"Name",
		"Stack",
		"Type",
		"AZ",
		"IP Address",
		"Key Name",
	})

	for _, instance := range instances {
		table.Append([]string{
			instance.InstanceId,
			instance.Name,
			instance.StackName,
			instance.InstanceType,
			instance.AZ,
			instance.IpAddress,
			instance.KeyName,
		})
	}
	table.Render()
}

func ExecInstance(instances []InstanceInfo, command string, concurrency int) error {
	var ipAddresses []string
	for _, instance := range instances {
		ipAddresses = append(ipAddresses, instance.IpAddress)
	}

	return ssh.ExecAll(ipAddresses, command, concurrency)
}

func SSHInstance(instances []InstanceInfo, sshCmd string) {
	if len(instances) == 1 {
		err := ssh.SSH(instances[0].IpAddress, sshCmd)
		if err != nil {
			log.Println(err)
		}
		return
	}

	fmt.Println("\nMultiple instances found:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(false)
	table.SetRowLine(false)
	table.SetColumnSeparator("")
	for i, instance := range instances {
		table.Append([]string{
			color.New(color.Bold).Sprint(strconv.Itoa(i + 1)),
			color.New(color.FgGreen).Sprint(instance.Name),
			color.New(color.FgMagenta, color.Bold).Sprint(instance.StackName),
			color.New(color.FgBlue, color.Bold).Sprint(instance.IpAddress),
		})
	}
	table.Render()
	fmt.Println("")

	var ipAddress string
	for {
		fmt.Print("Login to [1]: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			ipAddress = instances[0].IpAddress
			break
		}

		index, err := strconv.Atoi(input)
		if err != nil || index > len(instances) || index < 1 {
			continue
		}

		ipAddress = instances[index-1].IpAddress
		break
	}

	err := ssh.SSH(ipAddress, sshCmd)
	if err != nil {
		log.Fatal(err)
	}
}
