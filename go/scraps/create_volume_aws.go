package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {

	region := "us-west-1"
	availzone := "us-west-1a"

	svc := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	params := &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(availzone), // Required

		// With DryRun true:
		//  DryRunOperation: Request would have succeeded, but DryRun flag is set.
		//  status code: 412, request id: 6e56ac83-fa8f-4e1c-b4bc-7cb5ab888be2

		DryRun:    aws.Bool(false),
		Encrypted: aws.Bool(false),

		// Constraint: Range is 100 to 20000 for Provisioned IOPS SSD volumes:
		//Iops:             aws.Int64(1),

		// For encrypted volume:
		//KmsKeyId:         aws.String("String"),

		// In GB. Constraints: 1-16384 for gp2, 4-16384 for io1, 500-16384 for st1, 500-16384:
		Size: aws.Int64(30),

		// To create this vol from a snapshot:
		//SnapshotId:       aws.String("String"),

		// This can be gp2 for General Purpose SSD, io1 for Provisioned
		// IOPS SSD, st1 for Throughput Optimized HDD, sc1 for Cold HDD, or standard
		// for Magnetic volumes.
		VolumeType: aws.String("gp2"),
	}
	resp, err := svc.CreateVolume(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}
