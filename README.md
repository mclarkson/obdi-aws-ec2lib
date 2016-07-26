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
* In Plugins -> Add Plugin, choose 'aws-p2ec2' and Install.

# Dev

## REST End Points

### attach-volume

Attach a volume to a running or stopped instance.

http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.AttachVolume

```
POST data parameters for curl's '-d' option:

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

### create-volume

Create a volume in an availability zone.

http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumes.html

http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.CreateVolume

```
POST data parameters for curl's '-d' option:

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

### describe-availability-zone

Get the status of an availability zone.

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Get status of availability zone us-east-1c

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-availability-zone?env_id=1&region=us-east-1&availability_zone=us-east-1c"

```

### describe-instances

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

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-instances?env_id=1&&filter=instance-id=i-e12hb395"

# Filter on 2 instance-ids

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-instances?env_id=1&&filter=instance-id=i-e12hb395&filter=instance-id=i-7gbd59fe"

```

The filter name instance-id was used above. A list of all filter names are at:

  https://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#DescribeInstancesInput

### describe-regions

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Show all available regions. There are no other options.
# This does not call out to AWS, it uses the goamz library.

$ curl -k https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-regions

```

### describe-volume-status

Get the status of an EBS volume.

http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DescribeVolumeStatus

```
POST data parameters for curl's '-d' option:

    Device     string
    VolumeIds  list of strings
```

```
# Log in

$ ipport="127.0.0.1:443"

$ guid=`curl -ks -d '{"Login":"nomen.nescio","Password":"password"}' \
  https://$ipport/api/login | grep -o "[a-z0-9][^\"]*"`

# Get status using a GET request with POST data. This is allowed but frowned upon:
# http://stackoverflow.com/questions/978061/http-get-with-request-body#answer-983458

$ curl -k -X GET -d '{"VolumeIds":["vol-010101"]}' \
  "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-volume-status?env_id=1&region=us-east-1"

```

