package aws

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type InstanceInfo struct {
	InstanceId   string
	Name         string
	IpAddress    string
	InstanceType string
	AZ           string
	State        string
	KeyName      string
	StackName    string
}

func GetInstances(session *session.Session, query, stack, tag string, ips []string) ([]InstanceInfo, error) {
	svc := ec2.New(session)

	params := &ec2.DescribeInstancesInput{}
	if filters := filter(query, stack, tag, ips); len(filters) > 0 {
		params.Filters = filters
	}

	resp, err := svc.DescribeInstances(params)
	if err != nil {
		return nil, err
	}

	instances := make([]InstanceInfo, 0)
	for _, res := range resp.Reservations {
		for _, inst := range res.Instances {
			ipAddress := inst.PrivateIpAddress
			if ipAddress == nil {
				ipAddress = aws.String("")
			}

			var name string
			var stackName string
			for _, t := range inst.Tags {
				switch *t.Key {
				case "Name":
					name = *t.Value
				case "aws:cloudformation:stack-name":
					stackName = *t.Value
				default:
				}
			}

			instances = append(instances, InstanceInfo{
				InstanceId:   *inst.InstanceId,
				Name:         name,
				IpAddress:    *ipAddress,
				InstanceType: *inst.InstanceType,
				AZ:           *inst.Placement.AvailabilityZone,
				State:        *inst.State.Name,
				KeyName:      *inst.KeyName,
				StackName:    stackName,
			})
		}
	}

	return instances, nil
}

func filter(query, stack, tag string, ips []string) []*ec2.Filter {
	filters := []*ec2.Filter{}

	if query != "" {
		filters = append(filters, &ec2.Filter{
			Name: aws.String("tag-value"),
			Values: []*string{
				aws.String("*" + query + "*"),
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

	if tag != "" {
		tagParts := strings.Split(tag, ":")
		filters = append(filters, &ec2.Filter{
			Name: aws.String("tag:" + tagParts[0]),
			Values: []*string{
				aws.String(tagParts[1]),
			},
		})
	}

	if len(ips) > 0 {
		awsIps := make([]*string, len(ips))
		for i, ip := range ips {
			awsIps[i] = aws.String(ip)
		}
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("private-ip-address"),
			Values: awsIps,
		})
	}

	return filters
}
