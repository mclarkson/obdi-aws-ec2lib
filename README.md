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

### create-snapshot

Create a snapshot, in S3, of a volume.

[CreateSnapshot (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.CreateSnapshot)

```
POST data parameters for curl's '-d' option:

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

### create-volume

Create a volume in an availability zone.

[CreateVolume (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.CreateVolume)

[EBS Volumes](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/EBSVolumes.html)

[Device Naming](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/device_naming.html)

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

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-instances?env_id=1&filter=instance-id=i-e12hb395"

# Filter on 2 instance-ids

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-instances?env_id=1&filter=instance-id=i-e12hb395&filter=instance-id=i-7gbd59fe"

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

### describe-snapshots

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

$ curl -k "https://$ipport/api/nomen.nescio/$guid/aws-ec2lib/describe-snapshots?env_id=2&region=us-west-2&snapshot_id=snap-38cfe109&snapshot_id=snap-7fb97b3b"

```

### describe-volumes

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

### describe-volume-status

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

### detach-volume

Detach a volume from an instance.

[DetachVolume (go aws sdk)](http://docs.aws.amazon.com/sdk-for-go/api/service/ec2/#EC2.DetachVolume)

```
POST data parameters for curl's '-d' option:

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

