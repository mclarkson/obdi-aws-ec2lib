package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func main() {

	region := "us-west-1"
	//availzone := "us-west-1a"

	svc := ec2.New(session.New(), &aws.Config{Region: aws.String(region)})

	var GB8 int64
	GB8 = 8 * 1024 * 1024 * 1024

	disk_image := ec2.DiskImage{
		Description: aws.String("A disk image"),
		Image: &ec2.DiskImageDetail{
			Bytes: &GB8,
		},
	}

	var disk_images []*ec2.DiskImage
	disk_images = append(disk_images, &disk_image)

	params := &ec2.ImportInstanceInput{
		// A description for the instance being imported.
		Description: aws.String("A Description"),

		// The disk image.
		DiskImages: disk_images,
		/*
			type DiskImage struct {

				// A description of the disk image.
				Description *string `type:"string"`

				// Information about the disk image.
				Image *DiskImageDetail `type:"structure"`

				type DiskImageDetail struct {

					// The size of the disk image, in GiB.
					Bytes *int64 `locationName:"bytes" type:"long" required:"true"`

					// The disk image format.
					Format *string `locationName:"format" type:"string" required:"true" enum:"DiskImageFormat"`

					// A presigned URL for the import manifest stored in Amazon S3 and presented
					// here as an Amazon S3 presigned URL. For information about creating a presigned
					// URL for an Amazon S3 object, read the "Query String Request Authentication
					// Alternative" section of the Authenticating REST Requests (http://docs.aws.amazon.com/AmazonS3/latest/dev/RESTAuthentication.html)
					// topic in the Amazon Simple Storage Service Developer Guide.
					//
					// For information about the import manifest referenced by this API action,
					// see VM Import Manifest (http://docs.aws.amazon.com/AWSEC2/latest/APIReference/manifest.html).
					ImportManifestUrl *string `locationName:"importManifestUrl" type:"string" required:"true"`
					// contains filtered or unexported fields
				}


				// Information about the volume.
				Volume *VolumeDetail `type:"structure"`
				// contains filtered or unexported fields
			}
		*/

		// Checks whether you have the required permissions for the action, without
		// actually making the request, and provides an error response. If you have
		// the required permissions, the error response is DryRunOperation. Otherwise,
		// it is UnauthorizedOperation.
		//DryRun *bool `locationName:"dryRun" type:"boolean"`

		// The launch specification.
		//LaunchSpecification *ImportInstanceLaunchSpecification `locationName:"launchSpecification" type:"structure"`

		// The instance operating system.
		Platform: aws.String("Linux"),
	}

	resp, err := svc.ImportInstance(params)

	if err != nil {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		fmt.Println(err.Error())
		return
	}

	// Pretty-print the response data.
	fmt.Println(resp)
}
