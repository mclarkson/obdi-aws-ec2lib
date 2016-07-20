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

## Configuration

Set the AWS_ACCESS_KEY_ID_1 json object in the environment, using the Admin interface.
```
{

    "aws_access_key_id":"ALIENX2KD6OINVA510NQ",
    "aws_secret_access_key":"wHdlwoigU637fgnjAu+IRNVHfT-EXnIU5C2MbiQd",
    "aws_obdi_worker_instance_id":"i-e29eg362",
    "aws_obdi_worker_region":"us-east-1",
    "aws_obdi_worker_url":"https://1.2.3.4:4443/",
    "aws_obdi_worker_key":"secretkey",
    "aws_filter":"key-name=groupkey"

}
```

*aws_access_key_id*
> AWS API access key ID

*aws_secret_access_key*
> AWS API password

*aws_obdi_worker_instance_id*
> The AWS instance that contains the Obdi worker.

*aws_obdi_worker_region*
> The region the Obdi worker resides in.

*aws_obdi_worker_url*
> The URL used to access the Obdi worker.

*aws_obdi_worker_key*
> The password for the Obdi worker.

*aws_filter*
> The filter to apply, if any.

# Dev

