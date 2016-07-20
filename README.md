# obdi-aws-ec2lib
Library of AWS functions to be called via REST.

# Todo

# Screenshot

![](images/obdi-aws-ec2lib-small.png?raw=true)

# What is it?

# Installation

## Installing the plugin

* Log into the admin interface, 'https://ObdiHost/manager/admin'.
* In Plugins -> Manage Repositories add, 'https://github.com/mclarkson/obdi-awstools-repository.git'
* In Plugins -> Add Plugin, choose 'aws-p2ec2' and Install.

# Dev

**describe-instances**

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

