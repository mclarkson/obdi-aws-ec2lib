package main

import (
	"fmt"

	"github.com/goamz/goamz/aws"
)

func main() {

	for i, j := range aws.Regions {
		fmt.Printf("%s - %s\n", i, j.EC2Endpoint)
	}
}
