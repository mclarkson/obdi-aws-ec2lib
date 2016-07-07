package main

import (
	"fmt"
	"os"
	"strings"
)

import (
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/ec2"
)

var region = "us-east-1"

func instances(e *ec2.EC2, args ...string) {
	filter := ec2.NewFilter()
	for _, v := range args {
		sl := strings.SplitN(v, "=", 2)
		if len(sl) != 2 {
			fmt.Fprintf(os.Stderr, "instances: bad key=value pair \"%s\",
			skipping\n", v)
			continue
		}
		filter.Add(sl[0], sl[1])
	}

	resp, err := e.DescribeInstances(nil, filter)
	if err != nil {
		fmt.Fprintf(os.Stderr, "instances: %s\n", err)
		os.Exit(1)
	}

	for _, r := range resp.Reservations {
		fmt.Println("reservation:", r.ReservationId)
		for _, i := range r.Instances {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\n", i.InstanceId, i.State.Name,
			i.PrivateIPAddress, i.IPAddress, i.Tags[0], i.DNSName, i.ImageId)
			//fmt.Printf("%#v\n", i)
		}
	}
}

func main() {
	r, ok := aws.Regions[region]
	if !ok {
		fmt.Fprintf(os.Stderr,
			"unknown region: %s (aws regions to list all available)\n", region)
		os.Exit(1)
	}

	env := os.Getenv("AWS_ACCESS_KEY_ID")
	if env == "" {
		fmt.Fprintf(os.Stderr, "set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY\n")
		os.Exit(1)
	}

	auth, err := aws.EnvAuth()
	if err != nil {
		fmt.Fprintf(os.Stderr, "can not find auth info: %s\n", err)
		os.Exit(1)
	}

	e := ec2.New(auth, r)

	instances(e)
}
