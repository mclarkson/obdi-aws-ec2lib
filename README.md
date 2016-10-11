# obdi-aws-ec2lib
Library of AWS functions to be called via REST.

# Todo

# Screenshot

![](images/obdi-aws-ec2lib-small.png?raw=true)

# What is it?

A collection of REST end points that communicate with the AWS API. It is used
by other plugins.

# Installation

## Installing the plugin

* Log into the admin interface, 'https://ObdiHost/manager/admin'.
* In Plugins -> Manage Repositories add, 'https://github.com/mclarkson/obdi-awstools-repository.git'
* In Plugins -> Add Plugin, choose 'aws-ec2lib' and Install.

# Dev

## REST End Points

[attach-volume](#attach-volume)<br>
[copy-image](#copy-image)<br>
[copy-snapshot](#copy-snapshot)<br>
[create-image](#create-image)<br>
[create-snapshot](#create-snapshot)<br>
[create-volume](#create-volume)<br>
[delete-snapshot](#delete-snapshot)<br>
[delete-volume](#delete-volume)<br>
[describe-availability-zone](#describe-availability-zone)<br>
[describe-instances](#describe-instances)<br>
[describe-regions](#describe-regions)<br>
[describe-snapshots](#describe-snapshots)<br>
[describe-volume-status](#describe-volume-status)<br>
[describe-volumes](#describe-volumes)<br>
[detach-volume](#detach-volume)<br>
[import-image](#import-image)<br>
[import-instance](#import-instance)<br>
[register-image](#register-image)<br>
[run-instances](#run-instances)

### <a name="attach-volume"></a>attach-volume

Attach a volume to a running or stopped instance.

http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.AttachVolume

```
Supported POST data JSON parameters:

    Device     string
    DryRun     bool
    InstanceId string
    VolumeId   string
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Create a 30GB gp2 volume in availability zone us-west-2a

$ curl -k -d '{"Device":"/dev/sdb","InstanceId":"i-xxxxxx","VolumeId":"vol-xxxxx"}' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/attach-volume?env_id=1&region=us-west-2"

```

### <a name="copy-image"></a>copy-image

Copy an AMI to a different region.

[CopyImage (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.CopyImage)

```
Supported POST data JSON parameters:

    Name          string  // The name of the new AMI in the destination region.
    SourceImageId string  // The ID of the AMI to copy.
    SourceRegion  string  // The name of the region that contains the AMI to copy.
    Description   string  // A description for the new AMI in the destination region.
    DryRun        bool    // Checks whether you have the required permissions for the action.
    ClientToken   string  // Unique, case-sensitive identifier you provide to ensure idempotency.
    Encrypted     bool    // Specifies whether the destination snapshots of the copied image should be encrypted.
    KmsKeyId      string  // The full ARN of the AWS Key Management Service (AWS KMS) CMK to use when encrypting the snapshots
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Copy an image from us-west-2 to us-east-1

$ curl -k -d '
{
    "Name":"Fantastic Copied Image",
    "SourceRegion":"us-west-2",
    "SourceImageId":"ami-a9d109c9",
    "Description":"Image copied from us-west-2"
}' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/copy-image?env_id=2&region=us-east-1"
```

### <a name="copy-snapshot"></a>copy-snapshot

Copy a snapshot to the same or different region.

[CopySnapshot (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.CopySnapshot)

```
Supported POST data JSON parameters:

    SourceSnapshotId  string  // The ID of the EBS Snapshot to copy.
    SourceRegion      string  // The name of the region that contains the snapshot to copy.
    Description       string  // A description for the new AMI in the destination region.
    PresignedUrl      string  // The pre-signed URL that facilitates copying an encrypted snapshot.
    DestinationRegion string  // The destination region to use in the PresignedUrl parameter
                              //  of a snapshot copy operation. This parameter is only valid
                              //  for specifying the destination region in a PresignedUrl
                              //  parameter, where it is required.
    DryRun            bool    // Checks whether you have the required permissions for the action.
    Encrypted         bool    // Specifies whether the destination snapshots of the
                              //  copied image should be encrypted.
    KmsKeyId          string  // The full ARN of the AWS Key Management Service (AWS KMS)
                              //  CMK to use when encrypting the snapshots
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Copy a snapshot from us-west-2 to us-east-1

$ curl -k -d '
{
    "SourceRegion":"us-west-2",
    "SourceSnapshotId":"snap-ac56028b",
    "Description":"Snapshot copied from us-west-2"
}' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/copy-snapshot?env_id=2&region=us-east-1"

```

### <a name="create-image"></a>create-image

Create an AMI from an Amazon EBS-backed instance.

[CreateImage (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.CreateImage)

```
Supported POST data JSON parameters:

    Description string  // Description of the snapshot.
    DryRun      bool    // Check permissions.
    InstanceId  string  // The ID of the instance.
    Name        string  // A name for the new image.
    NoReboot    bool    // If true, the instance will not be shut down before creating the image.

    BlockDeviceMappings [{
        DeviceName  string  // The device name exposed to the instance (for example, /dev/sdh or xvdh).
        NoDevice    string  // Suppresses the specified device
        VirtualName string  // The virtual device name (ephemeralN).
        Ebs {
            DeleteOnTermination bool   // Indicates whether the EBS volume is deleted on termination
            Encrypted           bool   // Indicates whether the EBS volume is encrypted.
            Iops                int64  // The number of I/O operations per second (IOPS) that the volume supports.
            SnapshotId          string // The ID of the snapshot.
            VolumeSize          int64  // The size of the volume, in GiB.
            VolumeType          string // The volume type: gp2, io1, st1, sc1, or standard.
        }
    }]
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Create an image of instance i-2aa60a32

$ curl -k -d '{
    "InstanceID":"i-2aa60a32",
    "Description":"Created by obdi-aws-ec2lib from instance i-2aa60a32",
    "Name":"New AMI",
    "NoReboot":true }' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/create-image?env_id=2&region=us-west-2"

```

### <a name="create-snapshot"></a>create-snapshot

Create a snapshot, in S3, of a volume.

[CreateSnapshot (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.CreateSnapshot)

```
Supported POST data JSON parameters:

    Description string  // Description of the snapshot.
    VolumeId    string  // VolumeId to take a snapshot of.
    DryRun      bool
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Create a snapshot of vol-cb5f1166 in S3

$ curl -k -d '{ "DryRun":false,
                "Description":"Created by obdi-aws-p2ec2 for vol-cb5f1166",
                "VolumeId":"vol-cb5f1166"}' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/create-snapshot?env_id=1"

```

### <a name="create-volume"></a>create-volume

Create a volume in an availability zone.

[CreateVolume (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.CreateVolume)

[EBS Volumes](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumes.html)

[Device Naming](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/device_naming.html)

```
Supported POST data JSON parameters:

    Encrypted  bool
    Iops       int64  // 100 to 20000 for io1
    KmsKeyId   string // For encrypted volume
    Size       int64  // In GB
    SnapshotId string
    VolumeType string // gp2, io1, st1, sc1 or standard
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Create a 30GB gp2 volume in availability zone us-west-2a

$ curl -k -d '{"Size":30,"VolumeType":"gp2","Encrypted":false}' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/create-volume?env_id=1&region=us-west-2&availability_zone=us-west-2a"

```

### <a name="delete-snapshot"></a>delete-snapshot

Delete a snapshot from a region.

[DeleteSnapshot (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DeleteSnapshot)

```
Supported POST data JSON parameters:

    DryRun     bool
    SnapshotId   string
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Delete Volume vol-c5e13a4d

$ curl -k -d '{"DryRun":false,"SnapshotId":"snap-c5e13a4d"}' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/delete-snapshot?env_id=1&region=us-east-1"

```

### <a name="delete-volume"></a>delete-volume

Delete a volume from a region.

[DeleteVolume (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DeleteVolume)

```
Supported POST data JSON parameters:

    DryRun     bool
    VolumeId   string
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Delete Volume vol-c5e13a4d

$ curl -k -d '{"DryRun":false,"VolumeId":"vol-c5e13a4d"}' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/delete-volume?env_id=1&region=us-west-2"

```

### <a name="describe-availability-zone"></a>describe-availability-zone

Get the status of an availability zone.

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Get status of availability zone us-east-1c

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-availability-zone?env_id=1&region=us-east-1&availability_zone=us-east-1c"

```

### <a name="describe-instances"></a>describe-instances

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# All instances in the default region (us-east-1)

$ curl -k https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-instances?env_id=1

# All instances in us-west-1

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-instances?env_id=1&region=us-west-1"

# Filter on instance-id (the global filter is also applied if set)

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-instances?env_id=1&filter=instance-id=i-e12hb395"

# Filter on 2 instance-ids

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-instances?env_id=1&filter=instance-id=i-e12hb395&filter=instance-id=i-7gbd59fe"

```

The filter name instance-id was used above. A list of all filter names are at:

  https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#DescribeInstancesInput

### <a name="describe-regions"></a>describe-regions

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Show all available regions. There are no other options.
# This does not call out to AWS, it uses the goamz library.

$ curl -k https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-regions

```

### <a name="describe-snapshots"></a>describe-snapshots

Get the details of EBS snapshot(s).

http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeSnapshots

```
URL parameters:

    dry_run                "true"|"false".
    filter                 List. E.g. snapshot-id=snap-cd90d5ea.
                           Use more than once to specify more.
    owner_id               List. Use more than once to specify more.
    snapshot_id            List. Use more than once to specify more.
                           Omit snapshot_id to get status of all volumes.
    restorable_by_user_id  List. Use more than once to specify more.
    max_results            E.g. 100.
    next_token             E.g. token.
    env_id                 E.g. 1.
    region                 E.g. us-east-1.
```

The filter name snapshot-id was used above. A list of all filter names are at:

  https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#DescribeSnapshotsInput

To specify multiple values for a filter key use a comma, not more filter entries.

Examples:

Filter by status completed or error (but not pending status):
> &filter=status=completed,error

Filter by 3 'volume-id's that the snapshot is for.
> &filter=volume-id=vol-810baafb,vol-cdea3445,vol-800baafa

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Show all snapshots that you're allowed to see

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-snapshots?env_id=2&region=us-west-2"

# Show the details for two snapshot IDs
# It is an error to use a non-existent snapshot ID and AWS will complain

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-snapshots?env_id=2&region=us-west-2&snapshot_id=snap-38cfe109&snapshot_id=snap-7fb97b3b"

# Show the details for the same two snapshot IDs using filters instead
# It is /not/ an error to use a non-existent snapshot ID and AWS will /not/ complain.

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-snapshots?env_id=2&region=us-west-2&filter=snapshot-id=snap-38cfe109,snap-7fb97b3b"

```

### <a name="describe-volumes"></a>describe-volumes

Get the details of EBS volume(s).

http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeVolumes

```
URL parameters:

    dry_run     "true"|"false".
    volume_id   E.g. vol-3af379e. Use more than once to specify more.
                Omit volume_id to get status of all volumes.
    env_id      E.g. 1.
    region      E.g. us-east-1.
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Get details of all volumes

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-volumes?env_id=1&region=us-east-1"

# Get details of volume vol-3af379e

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-volumes?env_id=1&region=us-east-1&volume_id=vol-3af379e"

```

### <a name="describe-volume-status"></a>describe-volume-status

Get the status of an EBS volume.

http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeVolumeStatus

```
URL parameters:

    dry_run     "true"|"false".
    volume_id   E.g. vol-3af379e. Use more than once to specify more.
                Omit volume_id to get status of all volumes.
    env_id      E.g. 1.
    region      E.g. us-east-1.
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Get status of all volumes

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-volume-status?env_id=1&region=us-east-1"

# Get status of volume vol-3af379e

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-volume-status?env_id=1&region=us-east-1&volume_id=vol-3af379e"

```

### <a name="detach-volume"></a>detach-volume

Detach a volume from an instance.

[DetachVolume (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DetachVolume)

```
Supported POST data JSON parameters:

    DryRun     bool
    Device     string // The device name.
    Force      bool   // Last-resort force detachment.
    InstanceId string // The ID of the instance.
    VolumeId   string // The ID of the volume to be detached.
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Detach volume vol-cb5f1166, mounted on /dev/xvdb, from instance i-d0d63149:

$ curl -k -d '{"Device":"/dev/xvdb","InstanceId":"i-d0d63149","VolumeId":"vol-cb5f1166"}' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/detach-volume?env_id=1&region=us-west-1"

```

### <a name="import-image"></a>import-image

UNIMPLEMENTED - not sure how to do this and for it to make sense when using remotely.

### <a name="import-instance"></a>import-instance

UNIMPLEMENTED - not sure how to do this and for it to make sense when using remotely.

### <a name="register-image"></a>register-image

Create an AMI from a snapshot.

[RegisterImage (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.RegisterImage)

```
Supported POST data JSON parameters:

    Name               string  // A name for your AMI.
    Architecture       string
    Description        string  // A description for your AMI.
    DryRun             bool
    EnaSupport         bool    // enhanced networking with ENA.
    ImageLocation      string  // full path to your AMI manifest in Amazon S3.
    KernelId           string  // The ID of the kernel.
    RamdiskId          string  // The ID of the RAM disk.
    RootDeviceName     string  // for example, /dev/sda1, or /dev/xvda.
    SriovNetSupport    string  // enhanced networking with the Intel 82599
    VirtualizationType string  // The type of virtualization. Default: paravirtual

    BlockDeviceMappings [{
        DeviceName  string  // The device name exposed to the instance (for example, /dev/sdh or xvdh).
        NoDevice    string  // Suppresses the specified device
        VirtualName string  // The virtual device name (ephemeralN).
        Ebs {
            DeleteOnTermination bool   // Indicates whether the EBS volume is deleted on termination
            Encrypted           bool   // Indicates whether the EBS volume is encrypted.
            Iops                int64  // The number of I/O operations per second (IOPS) that the volume supports.
            SnapshotId          string // The ID of the snapshot.
            VolumeSize          int64  // The size of the volume, in GiB.
            VolumeType          string // The volume type: gp2, io1, st1, sc1, or standard.
        }
    }]
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Create an AMI from a snapshot

$ curl -k -d '
{
    "Name":"My AMI",
    "Description":"My AMI Description",
    "RootDeviceName":"sda1",
    "VirtualizationType":"hvm",
    "BlockDeviceMappings":[
        {
            "DeviceName":"sda1",
            "Ebs":{
                "DeleteOnTermination":true,
                "SnapshotId":"snap-af21558b",
                "VolumeSize":21,
                "VolumeType":"gp2"
             }
         }
    ]
}' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/register-image?env_id=2&region=us-west-2"

```

### <a name="run-instances"></a>run-instances

Create an Instance from an AMI.

[RunInstances (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.RunInstances)

```
Supported POST data JSON parameters:

    AdditionalInfo string  // Reserved.
    ClientToken    string  // Identifier you provide to ensure the idempotency.
    DisableApiTermination string  // If you set this parameter to true, you can't terminate the instance.
    DryRun         bool
    EbsOptimized   bool    // Indicates whether the instance is optimized for EBS I/O.
    ImageId        string  // The ID of the AMI.
    InstanceInitiatedShutdownBehavior string // Whether an instance stops or terminates.
    InstanceType   string  // The instance type. Default: m1.small.
    KernelId       string  // The ID of the kernel.
    KeyName        string  // The name of the key pair.
    MaxCount       int64   // The maximum number of instances to launch.
    MinCount       int64   // The minimum number of instances to launch.
    PrivateIpAddress string  // [EC2-VPC] The primary IP address.
    RamdiskId      string  // The ID of the RAM disk.
    SubnetId       string  // [EC2-VPC] The ID of the subnet to launch the instance into.
    UserData       string  // The user data to make available to the instance.

    // One or more security group IDs. You can create a security group using CreateSecurityGroup.
    // Default: Amazon EC2 uses the default security group.
    SecurityGroupIds [ string ]

    // [EC2-Classic, default VPC] One or more security group names. For a nondefault
    // VPC, you must use security group IDs instead.
    // Default: Amazon EC2 uses the default security group.
    SecurityGroups [ string ]

    Monitoring {
        Enabled bool  //Indicates whether monitoring is enabled for the instance.
    }

    NetworkInterfaces [{
        AssociatePublicIpAddress bool    // Indicates whether to assign a public IP address to an instance.
        DeleteOnTermination      bool    // Whether the interface is deleted on termination.
        Description              string  // The description of the network interface.
        DeviceIndex              int64   // You must provide the device index.
        Groups                   string  // The IDs of the security groups for the network interface.
        NetworkInterfaceId       string  // The ID of the network interface.
        PrivateIpAddress         string  // The private IP address of the network interface.
        SubnetId                 string  // The ID of the subnet associated with the network string.
        SecondaryPrivateIpAddressCount int64 // The number of secondary private IP addresses.

        PrivateIpAddresses [{
            Primary           bool   // Indicates whether this is the primary private IP address.
            PrivateIpAddress  string // The private IP addresses.
        }]

    }]

    Placement {
        Affinity         string  // The affinity setting for the instance on the Dedicated Host.
        AvailabilityZone string  // The Availability Zone of the instance.
        GroupName        string  // The name of the placement group the instance is in.
        HostId           string  // The ID of the Dedicted host on which the instance resides.
        Tenancy          string  // The tenancy of the instance (if the instance is running in a VPC).
                                 // default, dedicated or host
    }

    IamInstanceProfile IamInstanceProfileSpecification { // The IAM instance profile.
        Arn    string  // The Amazon Resource Name (ARN) of the instance profile.
        Name   string  // The name of the instance profile.
    }

    BlockDeviceMappings [{
        DeviceName  string  // The device name exposed to the instance (for example, /dev/sdh or xvdh).
        NoDevice    string  // Suppresses the specified device
        VirtualName string  // The virtual device name (ephemeralN).
        Ebs {
            DeleteOnTermination bool   // Indicates whether the EBS volume is deleted on termination
            Encrypted           bool   // Indicates whether the EBS volume is encrypted.
            Iops                int64  // The number of I/O operations per second (IOPS) that the volume supports.
            SnapshotId          string // The ID of the snapshot.
            VolumeSize          int64  // The size of the volume, in GiB.
            VolumeType          string // The volume type: gp2, io1, st1, sc1, or standard.
        }
    }]
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Create an Instance from an AMI using mostly default values.

$ curl -k -d '
{
    "ImageId":"ami-e6d00e86",
    "InstanceType":"t2.micro",
    "MaxCount":1,
    "MinCount":1,
    "BlockDeviceMappings":[
        {
            "DeviceName":"sda1",
            "Ebs":{
                "DeleteOnTermination":true,
                "VolumeSize":21,
                "VolumeType":"gp2"
             }
         }
    ]
}' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/run-instances?env_id=2&region=us-west-2"

# Create an Instance from an AMI using mostly default values.
# Additionally specify the Availability zone and security group.

$ curl -k -d '
{
    "ImageId":"ami-e6d00e86",
    "InstanceType":"t2.micro",
    "MaxCount":1,
    "MinCount":1,
    "SecurityGroups":"mainsg"
    "Placement": {
        "AvailabilityZone":"us-west-2c"
    },
    "BlockDeviceMappings":[
        {
            "DeviceName":"sda1",
            "Ebs":{
                "DeleteOnTermination":true,
                "VolumeSize":21,
                "VolumeType":"gp2"
             }
         }
    ]
}' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/run-instances?env_id=2&region=us-west-2"

```

