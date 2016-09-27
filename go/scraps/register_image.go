package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {

	region := "us-west-2"
	//availzone := "us-west-1a"

	svc := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	bdm := ec2.BlockDeviceMapping{
		DeviceName: aws.String("sda1"),
		Ebs: &ec2.EbsBlockDevice{
			DeleteOnTermination: aws.Bool(true),
			//Encrypted:           aws.Bool(false),
			//Iops:                aws.Int64(1),
			SnapshotId: aws.String("snap-eb8c4617"),
			VolumeSize: aws.Int64(21),
			VolumeType: aws.String("gp2"),
		},
		//NoDevice:    aws.String("String"),
		//VirtualName: aws.String("String"),
	}

	var bdms []*ec2.BlockDeviceMapping
	bdms = append(bdms, &bdm)

	params := &ec2.RegisterImageInput{
		Name: aws.String("PRET-TEST-NAME"), // Required
		//Architecture: aws.String("ArchitectureValues"),
		BlockDeviceMappings: bdms,
		Description:         aws.String("PRET-TEST-DESC"),
		DryRun:              aws.Bool(false),
		//EnaSupport:         aws.Bool(true),
		//ImageLocation:      aws.String("String"),
		//KernelId:           aws.String("String"),
		//RamdiskId:          aws.String("String"),
		RootDeviceName: aws.String("sda1"),
		//SriovNetSupport:    aws.String("String"),
		VirtualizationType: aws.String("hvm"),
	}

	resp, err := svc.RegisterImage(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}
