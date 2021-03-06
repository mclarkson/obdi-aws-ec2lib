package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {
	// Create an EC2 service object in the "us-west-2" region
	// Note that you can also configure your region globally by
	// exporting the AWS_REGION environment variable
	creds := credentials.NewStaticCredentials("ID", "KEY", "")
	svc := ec2.New(session.New(), &aws.Config{Region: aws.String("us-east-1"), Credentials: creds})

	zone := "us-east-1c"
	params := &ec2.DescribeAvailabilityZonesInput{
		ZoneNames: []*string{&zone},
	}

	resp, err := svc.DescribeAvailabilityZones(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}
