package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {
	// Create an EC2 service object in the "us-west-2" region
	// Note that you can also configure your region globally by
	// exporting the AWS_REGION environment variable
	svc := ec2.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})

	/*
		params := &ec2.DescribeRegionsInput{
			DryRun: aws.Bool(false),
			RegionNames: []*string{
				aws.String("ap-northeast-1"),
				aws.String("us-east-1"),
				aws.String("us-west-1"),
				aws.String("us-west-2"),
			},
		}

		resp, err := svc.DescribeRegions(params)
	*/
	resp, err := svc.DescribeRegions(nil)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}
