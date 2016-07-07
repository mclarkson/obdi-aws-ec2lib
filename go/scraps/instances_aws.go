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

	// Call the DescribeInstances Operation
	resp, err := svc.DescribeInstances(nil)
	if err != nil {
		panic(err)
	}

	// resp has all of the response data, pull out instance IDs:
	fmt.Println("> Number of reservation sets: ", len(resp.Reservations))
	for idx, res := range resp.Reservations {
		fmt.Println("  > Number of instances: ", len(res.Instances))
		for _, inst := range resp.Reservations[idx].Instances {
			fmt.Println("    - Instance ID: ", *inst.InstanceId)
			fmt.Println(*inst)
		}
	}
}

/* resp.Reservations.Instances
{
  AmiLaunchIndex: 0,
  Architecture: "x86_64",
  BlockDeviceMappings: [{
      DeviceName: "/dev/sda1",
      Ebs: {
        AttachTime: 2016-05-12 12:00:57 +0000 UTC,
        DeleteOnTermination: false,
        Status: "attached",
        VolumeId: "vol-07c876ac"
      }
    }],
  ClientToken: "14630544459935275",
  EbsOptimized: true,
  Hypervisor: "xen",
  ImageId: "ami-6d1c2007",
  InstanceId: "i-044b1583",
  InstanceType: "m4.10xlarge",
  KeyName: "hcompret",
  LaunchTime: 2016-05-12 12:00:56 +0000 UTC,
  Monitoring: {
    State: "enabled"
  },
  NetworkInterfaces: [{
      Attachment: {
        AttachTime: 2016-05-12 12:00:56 +0000 UTC,
        AttachmentId: "eni-attach-96792d6b",
        DeleteOnTermination: true,
        DeviceIndex: 0,
        Status: "attached"
      },
      Description: "Primary network interface",
      Groups: [{
          GroupId: "sg-25c7355e",
          GroupName: "pret-security-group"
        }],
      MacAddress: "0a:06:70:a8:11:cb",
      NetworkInterfaceId: "eni-3b66767b",
      OwnerId: "355169842987",
      PrivateIpAddress: "10.17.9.54",
      PrivateIpAddresses: [{
          Primary: true,
          PrivateIpAddress: "10.17.9.54"
        }],
      SourceDestCheck: true,
      Status: "in-use",
      SubnetId: "subnet-483ed83e",
      VpcId: "vpc-50d07134"
    }],
  Placement: {
    AvailabilityZone: "us-east-1c",
    GroupName: "",
    Tenancy: "default"
  },
  PrivateDnsName: "ip-10-17-9-54.us-east-1.compute.internal",
  PrivateIpAddress: "10.17.9.54",
  ProductCodes: [{
      ProductCodeId: "aw0evgkw8e5c1q413zgy5pjce",
      ProductCodeType: "marketplace"
    }],
  PublicDnsName: "",
  RootDeviceName: "/dev/sda1",
  RootDeviceType: "ebs",
  SecurityGroups: [{
      GroupId: "sg-25c7355e",
      GroupName: "pret-security-group"
    }],
  SourceDestCheck: true,
  State: {
    Code: 16,
    Name: "running"
  },
  StateTransitionReason: "",
  SubnetId: "subnet-483ed83e",
  Tags: [{
      Key: "Name",
      Value: "pret-jmeter01"
    }],
  VirtualizationType: "hvm",
  VpcId: "vpc-50d07134"
}
*/
