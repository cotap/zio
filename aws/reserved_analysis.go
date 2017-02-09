package aws

import (
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/olekukonko/tablewriter"
)

type utilization struct {
	reserved  int
	allocated int
}

type reservedAnalysis map[string]map[string]*utilization

func (r reservedAnalysis) reserve(az string, instanceType string, count int) *utilization {
	i, ok := r[instanceType]
	if !ok {
		i = make(map[string]*utilization)
		r[instanceType] = i
	}

	u, ok := i[az]
	if !ok {
		u = &utilization{reserved: 0, allocated: 0}
		i[az] = u
	}

	u.reserved += count
	return u
}

func (r reservedAnalysis) allocate(az string, instanceType string) {
	u, ok := r[instanceType][az]
	if !ok {
		u = r.reserve(az, instanceType, 0)
	}
	u.allocated += 1
}

func ReservedAnalysis(session *session.Session) {
	svc := ec2.New(session)

	instances, err := svc.DescribeInstances(nil)
	if err != nil {
		log.Fatal(err)
	}

	reservations, err := svc.DescribeReservedInstances(nil)
	if err != nil {
		log.Fatal(err)
	}

	analysis := make(reservedAnalysis)

	for _, r := range reservations.ReservedInstances {
		analysis.reserve(*r.AvailabilityZone, *r.InstanceType, int(*r.InstanceCount))
	}

	for _, res := range instances.Reservations {
		for _, inst := range res.Instances {
			analysis.allocate(*inst.Placement.AvailabilityZone, *inst.InstanceType)
		}
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{
		"Type",
		"AZ",
		"Allocated",
		"Reserved",
		"Delta",
	})
	for instanceType, azs := range analysis {
		for az, u := range azs {
			table.Append([]string{
				instanceType,
				az,
				strconv.Itoa(u.allocated),
				strconv.Itoa(u.reserved),
				strconv.Itoa(u.reserved - u.allocated),
			})
		}
	}
	table.Render()
}
