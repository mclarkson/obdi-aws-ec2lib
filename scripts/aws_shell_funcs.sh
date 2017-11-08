#!/bin/bash

# ---------------------------------------------------------------------------
# GLOBAL DEFINITIONS (CHANGE AS NECESSARY)
# ---------------------------------------------------------------------------

# For priming ssh
KNOWNHOSTS=~/.ssh/known_hosts
TMPKEYFILE=~/.tmpkeyfile.key.$$

# The following variables should be set externally, by the golang
# REST end-point, as in obdi-aws-p2ec2/osedits-centos6.go.

# User to log in as
REMOTEUSER="ec2-user"
# Put a base64 encoded key in PRIVKEYB64
#PRIVKEYB64=""

# ---------------------------------------------------------------------------
# GLOBAL DECLARATIONS
# ---------------------------------------------------------------------------

# Temporary files:
t= #stdout
e= #stderr

# Globals set by AWS functions:
VOLUMEID=
VOLSIZE=
SNAPSHOTID=
SNAPSHOTIDS=
SNAPVOL=
PRIMARY_IP_ADDRESS=
DEVICE_ATTACHMENT=
NEWINSTANCEID=
NEWIP=
AMIID=

# ---------------------------------------------------------------------------
# ENVIRONMENT GLOBALS (COPIED TO THIS SCRIPT FROM THE CALLER)
# ---------------------------------------------------------------------------

# Expected environment vars:
SIZE=$SIZE
ENVID=$ENVID
REGION=$REGION
USERID=$USERID
AVAILZONECHAR=$AVAILZONECHAR
GUID=$GUID

# ---------------------------------------------------------------------------
main() {
# ---------------------------------------------------------------------------
# It all starts here

    export PATH

    local instanceid=$1

    [[ -z $instanceid ]] && {
        echo "Usage: $(basename $0) <instance id>"
        exit 1
    }

    PROTO="https"
    OPTS="-k -s"    # don't check ssl cert, silent
    OPTSD="-k -v"   # don't check ssl cert (shows errors for debugging)
    IPPORT="127.0.0.1:443"

    sanity_checks
    create_temp_file # files are $t and $e (stdout and stderr)

    # vvvvvvvvvv YOUR COMMANDS GO HERE vvvvvvvvvv

    # EXAMPLE. START A NEW INSTANCE BASED ON AMAZON LINUX AMI
    #          THEN LOG INTO IT AND LIST THE ROOT DIR FOR FUN

    # The official Amazon Linux AMI
    AMAZON_LINUX="ami-f173cc91"
    # AWS Subnet
    SUBNET_ID="subnet-1edaf568"
    # List of AWS security groups
    SECURITY_GROUP_IDS="[\"sg-4692f73f\"]"
    # AWS key
    AWS_KEYNAME="pret"
    # Placement group
    GROUPNAME=""

    create_instance $AMAZON_LINUX     # (sets $NEWINSTANCEID & $NEWIP)

    ssh_cmd $NEWIP savestdout \
        "ls -lh /"

    echo "Output from command is:"
    echo "$LAST_STDOUT"

    # ^^^^^^^^^^ YOUR COMMANDS GO HERE ^^^^^^^^^^

    echo "Finished!"
}

# ===========================================================================
# AWS FUNCTIONS
# ===========================================================================

# attach_volume
# create_instance
# create_placement_group
# create_snapshot
# create_tag
# create_volume
# create_volume_from_snapshot
# delete_snapshot
# delete_volume
# deregister_image
# detach_volume
# get_device_volumeid
# get_snapshotids_from_image
# get_vol_size
# register_image
# start_instance
# stop_instance
# terminate_instance
# wait_for_instance
# wait_for_snapshot
# wait_for_volstatus
# wait_for_volume

# ---------------------------------------------------------------------------
get_device_volumeid() {
# ---------------------------------------------------------------------------
# Get volumeid of device
# Arg1 - instance id
# Arg2 - device, eg. sda, sda1, xvda, xvdb
# Sets global var: VOLUMEID

    local instanceid=$1 device=$2

    echo "`date` Querying instance $instanceid for device /dev/$device..."

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/describe-instances"
    
    curl $OPTSD \
        "$url?env_id=$ENVID&region=$REGION&filter=instance-id=$instanceid" \
        >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    VOLUMEID=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
        | python -mjson.tool \
        | sed -n '/"DeviceName": "\/dev\/'"$device"'"/,/ *}/ { s/^ *"VolumeId": "\(.*\)"/\1/p }'
    )

    [[ -z $VOLUMEID ]] && {
        echo
        echo "ERROR: VOLUMEID is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
| python -mjson.tool \
| sed -n '/"DeviceName": "\/dev\/'"$device"'"/,/ *}/ { s/^ *"VolumeId": "\(.*\)"/\1/p }'
EnD
        cat $e
        exit 1
    }

    echo "`date` Device /dev/$device: $VOLUMEID"
}

# ---------------------------------------------------------------------------
get_vol_size() {
# ---------------------------------------------------------------------------
# Get size of boot volume /dev/sda1
# Arg1 - volume id to query
# Sets global var: VOLSIZE

    local volumeid=$1

    echo "`date` Querying size of volume $volumeid..."

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/describe-volumes"

    curl $OPTSD \
        "$url?env_id=$ENVID&region=$REGION&volume_id=$volumeid" \
        >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    VOLSIZE=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
        | python -mjson.tool \
        | sed -n 's/^ *"Size": \(.*\),/\1/p'
    )

    [[ -z $VOLSIZE ]] && {
        echo
        echo "ERROR: VOLSIZE is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
| python -mjson.tool \
| sed -n 's/^ *"Size": \(.*\),/\1/p'
EnD
        cat $e
        exit 1
    }

    echo "`date` Volume size GiB: $VOLSIZE"
}

# ---------------------------------------------------------------------------
create_snapshot() {
# ---------------------------------------------------------------------------
# Create a snapshot from the volume
# Arg1 - volume id to snapshot
# Sets global var: SNAPSHOTID

    local volumeid=$1

    echo "`date` Creating snapshot from $volumeid"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/create-snapshot"

    curl $OPTSD -d "
    {
        \"DryRun\":false,
        \"Description\":\"Created by obdi-hcom-dse-pcs-cache from $volumeid\",
        \"VolumeId\":\"$volumeid\"
    }" \
    "$url?env_id=$ENVID&region=$REGION" \
    >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    SNAPSHOTID=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
        | python -mjson.tool \
        | sed -n 's/^ *"SnapshotId": "\(.*\)".*/\1/p' 
    )

    [[ -z $SNAPSHOTID ]] && {
        echo
        echo "ERROR: SNAPSHOTID is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
| python -mjson.tool \
| sed -n 's/^ *"SnapshotId": "\(.*\)".*/\1/p' 
EnD
        cat $e
        exit 1
    }

    echo "`date` Snapshot: $SNAPSHOTID"
}

# ---------------------------------------------------------------------------
wait_for_snapshot() {
# ---------------------------------------------------------------------------
# Wait for snapshot to complete
# Arg1 - snapshot id to snapshot
# Sets global var: none

    local snapstatus snapshotid=$1

    declare -i i=0

    while true; do

        echo "`date` Waiting for snapshot to complete"

        sleep 10

        url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/describe-snapshots"

        snapstatus=$(
            curl $OPTS "$url?env_id=$ENVID&region=$REGION&snapshot_id=$snapshotid" \
            | python -mjson.tool \
            | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
            | python -mjson.tool \
            | sed -n 's/^ *"State": "\(.*\)".*/\1/p'
        )

        [[ $snapstatus == "completed" ]] && break
        #[[ $snapstatus == "pending" ]] && break

        [[ $i -ge 120 ]] && {
            echo "Gave up waiting for snapshot to complete. Aborting"
            exit 1
        }

        i=i+1
    done

    echo "`date` Snapshot is 'completed'"
}

# ---------------------------------------------------------------------------
create_tag() {
# ---------------------------------------------------------------------------
# Create a new tag for a resource
# Arg1 - resource id
# Arg2 - Tag Key
# Arg3 - Tag Value
# Sets global var: none

    local resourceid=$1 key="$2" val="$3"

    echo "`date` Tagging resource, $resourceid, with '$key:$val'"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/create-tags"

    curl $OPTSD -d "
    {
        \"Resources\": [ \"$resourceid\" ],
        \"Tags\": [
            {
                \"Key\":\"$key\",
                \"Value\":\"$val\"
            }
        ]
    }" "$url?env_id=$ENVID&region=$REGION" \
    >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    dummyval=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//g;p}'
    )

    [[ -z $dummyval ]] && {
        echo
        echo "ERROR: dummyval is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\\\//g;p}'
EnD
        cat $e
        exit 1
    }

    echo "`date` Created tag: '$key:$val' for resource $resourceid"
}

# ---------------------------------------------------------------------------
create_volume() {
# ---------------------------------------------------------------------------
# Create a Volume
# Arg1 - Volume size in GiB
# Sets global var: VOLUMEID

    local volsize=$1

    echo "`date` Creating new volume"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/create-volume"

    curl $OPTSD -d "
    {
        \"VolumeType\":\"gp2\",
        \"Size\": $volsize
    }" \
    "$url?env_id=$ENVID&region=$REGION&availability_zone=${REGION}$AVAILZONECHAR" \
    >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    VOLUMEID=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
        | python -mjson.tool \
        | sed -n 's/^ *"VolumeId": "\(.*\)".*/\1/p' 
    )

    [[ -z $VOLUMEID ]] && {
        echo
        echo "ERROR: VOLUMEID is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
| python -mjson.tool \
| sed -n 's/^ *"VolumeId": "\(.*\)".*/\1/p' 
EnD
        cat $e
        exit 1
    }

    echo "`date` Created volume: $VOLUMEID"
}

# ---------------------------------------------------------------------------
create_volume_from_snapshot() {
# ---------------------------------------------------------------------------
# Create a Volume from the snapshot
# Arg1 - Snapshot ID
# Arg2 - Volume size in GiB
# Sets global var: SNAPVOL

    local snapshotid=$1 volsize=$2

    echo "`date` Creating volume from $snapshotid"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/create-volume"

    curl $OPTSD -d "
    {
        \"VolumeType\":\"gp2\",
        \"SnapshotId\":\"$snapshotid\",
        \"Size\": $volsize
    }" \
    "$url?env_id=$ENVID&region=$REGION&availability_zone=${REGION}$AVAILZONECHAR" \
    >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    SNAPVOL=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
        | python -mjson.tool \
        | sed -n 's/^ *"VolumeId": "\(.*\)".*/\1/p' 
    )

    [[ -z $SNAPVOL ]] && {
        echo
        echo "ERROR: SNAPVOL is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
| python -mjson.tool \
| sed -n 's/^ *"VolumeId": "\(.*\)".*/\1/p' 
EnD
        cat $e
        exit 1
    }

    echo "`date` Volume: $SNAPVOL"
}

# ---------------------------------------------------------------------------
wait_for_instance() {
# ---------------------------------------------------------------------------
# Wait for instance to be started
# Sets global var: PRIMARY_IP_ADDRESS

    local inststatus instid=$1 wantedstatus=$2

    declare -i i=0
    
    while true; do

        echo "`date` Waiting for instance, $instid, to be $wantedstatus"

        sleep 10

        url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/describe-instances"

        curl $OPTSD \
            "$url?env_id=$ENVID&region=$REGION&filter=instance-id=$instid" \
            >$t 2>$e

        inststatus=$( cat $t | python -mjson.tool \
            | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
            | python -mjson.tool \
            | sed -n '/"State": /,/ *}/ { s/^ *"Name": "\(.*\)"[, ]*/\1/p }'
        )

        [[ $inststatus == "$wantedstatus" ]] && break

        [[ $i -ge 120 ]] && {
            echo "Gave up waiting for instance to be $wantedstatus Aborting"
            exit 1
        }

        i=i+1
    done

    PRIMARY_IP_ADDRESS=$(cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
        | python -mjson.tool \
        | sed '/"NetworkInterfaces": \[/,/^ *\]/ {d}' \
        | sed -n 's/^ *"PrivateIPAddress": "\(.*\)"[, ]*/\1/p'
    )

    echo "`date` Instance, $instid, is '$wantedstatus' ($PRIMARY_IP_ADDRESS)"
}

# ---------------------------------------------------------------------------
wait_for_volume() {
# ---------------------------------------------------------------------------
# Wait for volume to complete
# Sets global var: none

    local volstatus volid=$1

    declare -i i=0
    
    while true; do

        echo "`date` Waiting for volume, $volid, to become available"

        sleep 10

        url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/describe-volumes"

        volstatus=$(
            curl $OPTS "$url?env_id=$ENVID&region=$REGION&volume_id=$volid" \
            | python -mjson.tool \
            | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
            | python -mjson.tool \
            | sed -n 's/^ *"State": "\(.*\)".*/\1/p'
        )

        [[ $volstatus == "available" ]] && break

        [[ $i -ge 120 ]] && {
            echo "Gave up waiting for volume to become available. Aborting"
            exit 1
        }

        i=i+1
    done

    echo "`date` Volume, $volid, is 'available'"
}

# ---------------------------------------------------------------------------
attach_volume() {
# ---------------------------------------------------------------------------
# Attach a Volume to an AWS instance (the obdi master)
# Sets global var: DEVICE_ATTACHMENT

    local newvol=$1 instance=$2 discardedvalue
    echo "`date` Attaching volume, $newvol, to $instance"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/attach-volume"

    local letters="fghijklmnopqrstuvwxyz"

    for i in {0..20}; do

        discardedvalue=$(
            curl $OPTS -d "
            {
                \"Device\":\"/dev/sd${letters:i:1}\",
                \"InstanceId\": \"$instance\",
                \"VolumeId\": \"$newvol\"
            }" \
            "$url?env_id=$ENVID&region=$REGION&availability_zone=${REGION}$AVAILZONECHAR" \
            | python -mjson.tool \
            | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}'
        )

        [[ -z $discardedvalue ]] && {
            # Assume the mount failed because it's in use - A bad assumption.
            sleep 1 # to avoid api limits
            echo "`date` Device /dev/sd${letters:i:1} is in use. Trying next device."
            continue
        }
        
        echo "`date` Volume attachment request sent for device /dev/sd${letters:i:1}"
        DEVICE_ATTACHMENT="/dev/sd${letters:i:1}"
        return

    done

    echo "`date` Could not find free device, or some other error!"
    exit 1
}

# ---------------------------------------------------------------------------
detach_volume() {
# ---------------------------------------------------------------------------
# Detach a Volume from an AWS instance (the obdi master)
# Arg1 - volume id
# Arg2 - instance id
# Arg3 - device, eg. /dev/sdf
# Sets global var: none

    local vol=$1 instance=$2 dev=$3 discardedvalue
    echo "`date` Detaching volume, $vol, from $instance"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/detach-volume"

    curl $OPTS -d "
    {
        \"Device\":\"$dev\",
        \"InstanceId\": \"$instance\",
        \"VolumeId\": \"$vol\"
    }" \
    "$url?env_id=$ENVID&region=$REGION&availability_zone=${REGION}$AVAILZONECHAR" \
        >$t 2>$e

    #[[ $? -ne 0 ]] && show_err_quit

    discardedvalue=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}'
    )

    [[ -z $discardedvalue ]] && {
        # Just let the error go!
        echo "`date` ERROR: Detach FAILED!"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}'
EnD
        cat $e
    }
    
    echo "`date` Volume detachment request sent for device $dev"
}

# ---------------------------------------------------------------------------
delete_volume() {
# ---------------------------------------------------------------------------
# Delete a Volume
# Arg1 - volumeid
# Sets global var: none

    local vol=$1 discardedvalue
    echo "`date` Deleting volume, $vol"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/delete-volume"

    curl $OPTS -d "
    {
        \"VolumeId\": \"$vol\"
    }" \
    "$url?env_id=$ENVID&region=$REGION" \
        >$t 2>$e

    #[[ $? -ne 0 ]] && show_err_quit

    discardedvalue=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//g;p}'
    )

    [[ -z $discardedvalue ]] && {
        # Just let the error go!
        echo "`date` ERROR: Delete FAILED!"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\\\//g;p}'
EnD
        cat $e
    }
    
    echo "`date` Volume delete request sent for volume $vol"
}

# ---------------------------------------------------------------------------
delete_snapshot() {
# ---------------------------------------------------------------------------
# Delete a Volume
# Arg1 - snapshot id
# Sets global var: none

    local snapid=$1 discardedvalue

    echo "`date` Deleting snapshot, $snapid"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/delete-snapshot"

    curl $OPTS -d "
    {
        \"SnapshotId\": \"$snapid\"
    }" \
    "$url?env_id=$ENVID&region=$REGION" \
        >$t 2>$e

    #[[ $? -ne 0 ]] && show_err_quit

    discardedvalue=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//g;p}'
    )

    [[ -z $discardedvalue ]] && {
        # Just let the error go!
        echo "`date` ERROR: Delete snapshot FAILED!"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\\\//g;p}'
EnD
        cat $e
    }
    
    echo "`date` Snapshot delete request sent for snapshot $snapid"
}

# ---------------------------------------------------------------------------
wait_for_volstatus() {
# ---------------------------------------------------------------------------
# Wait for volume to be attached
# Arg1 - volume id to check
# Arg2 - volume status (State) to wait for
# Sets global var: none

    local volstatus volid=$1 wantedstatus=$2

    declare -i i=0
    
    while true; do

        echo "`date` Waiting for volume, $volid, to be attached"

        sleep 10

        url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/describe-volumes"

        volstatus=$(
            curl $OPTS "$url?env_id=$ENVID&region=$REGION&volume_id=$volid" \
            | python -mjson.tool \
            | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
            | python -mjson.tool \
            | sed -n '/"Attachments": /,/ *}/ { s/^ *"State": "\(.*\)"[, ]*/\1/p }'
        )

        [[ $volstatus == "$wantedstatus" ]] && break

        [[ $i -ge 120 ]] && {
            echo "Gave up waiting for volume, $volid, to be attached. Aborting"
            echo "Try rebooting the instance, $instanceid."
            exit 1
        }

        i=i+1
    done

    echo "`date` Volume, $volid, is '$wantedstatus'"
}

# ---------------------------------------------------------------------------
create_instance() {
# ---------------------------------------------------------------------------
# Create a new instance from an ami
# Arg1 - AMI ID
# Sets global var: NEWINSTANCEID, NEWIP

    local amiid=$1

    echo "`date` Creating new instance"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/run-instances"

    curl $OPTSD -d "
    {
        \"ImageId\":\"$amiid\",
        \"InstanceType\":\"i2.2xlarge\",
        \"MaxCount\":1,
        \"MinCount\":1,
        \"SubnetId\":\"$SUBNET_ID\",
        \"SecurityGroupIds\":$SECURITY_GROUP_IDS,
        \"KeyName\":\"$AWS_KEYNAME\",
        \"Placement\": {
            \"AvailabilityZone\":\"${REGION}$AVAILZONECHAR\",
            \"GroupName\":\"$GROUPNAME\"
        }
    }" "$url?env_id=$ENVID&region=$REGION" \
    >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    NEWINSTANCEID=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
        | python -mjson.tool \
        | sed -n 's/^ *"InstanceId": "\(.*\)".*/\1/p' 
    )

    NEWIP=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
        | python -mjson.tool \
        | grep -m1 PrivateIpAddress \
        | sed -n 's/^ *"PrivateIpAddress": "\(.*\)"[, ]*/\1/p'
    )

    [[ -z $NEWINSTANCEID ]] && {
        echo
        echo "ERROR: NEWINSTANCEID is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
| python -mjson.tool \
| sed -n 's/^ *"InstanceId": "\(.*\)".*/\1/p' 
EnD
        cat $e
        exit 1
    }

    echo "`date` Created instance(s): $NEWINSTANCEID"
}

# ---------------------------------------------------------------------------
start_instance() {
# ---------------------------------------------------------------------------
# Create a new instance from an ami
# Sets global var: none

    local instanceid=$1 dummyval

    echo "`date` Starting instance $instanceid"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/start-instances"

    curl $OPTSD -d "
    {
        \"InstanceIds\":[\"$instanceid\"]
    }" "$url?env_id=$ENVID&region=$REGION" \
    >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    dummyval=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}'
    )

    [[ -z $dummyval ]] && {
        echo
        echo "ERROR: dummyval is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}'
EnD
        cat $e
        exit 1
    }
}

# ---------------------------------------------------------------------------
stop_instance() {
# ---------------------------------------------------------------------------
# Create a new instance from an ami
# Sets global var: none

    local instanceid=$1 dummyval

    echo "`date` Stopping instance $instanceid"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/stop-instances"

    curl $OPTSD -d "
    {
        \"InstanceIds\":[\"$instanceid\"]
    }" "$url?env_id=$ENVID&region=$REGION" \
    >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    dummyval=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}'
    )

    [[ -z $dummyval ]] && {
        echo
        echo "ERROR: dummyval is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}'
EnD
        cat $e
        exit 1
    }
}

# ---------------------------------------------------------------------------
terminate_instance() {
# ---------------------------------------------------------------------------
# Create a new instance from an ami
# Arg1 - instance id
# Sets global var: none

    local instanceid=$1 dummyval

    echo "`date` Terminate instance $instanceid"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/terminate-instances"

    curl $OPTSD -d "
    {
        \"InstanceIds\":[\"$instanceid\"]
    }" "$url?env_id=$ENVID&region=$REGION" \
    >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    dummyval=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}'
    )

    [[ -z $dummyval ]] && {
        echo
        echo "ERROR: dummyval is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}'
EnD
        cat $e
        exit 1
    }
}

# ---------------------------------------------------------------------------
register_image() {
# ---------------------------------------------------------------------------
# Create a new instance from an ami
# Arg1 - snapshot id of boot vol
# Arg2 - snapshot id of xvdb vol
# Sets global var: AMIID

    local snapshotid1=$1 snapshotid2=$2 dummyval

    echo "`date` Creating AMI, $AMINAME, from $snapshotid1 and $snapshotid2"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/register-image"

    curl $OPTSD -d "
    {
        \"Name\": \"$AMINAME\",
        \"Description\": \"Created by obdi-hcom-dse-pcs-cache from $snapshotid1\",
        \"RootDeviceName\": \"sda1\",
        \"VirtualizationType\": \"hvm\",
        \"Architecture\": \"x86_64\",
        \"BlockDeviceMappings\": [
        {
           \"DeviceName\": \"sda1\",
           \"Ebs\": {
               \"DeleteOnTermination\": true,
               \"SnapshotId\": \"$snapshotid1\",
               \"VolumeSize\": $bootvolsize,
               \"VolumeType\": \"gp2\"
           }
        },{
           \"DeviceName\": \"xvdb\",
           \"Ebs\": {
               \"DeleteOnTermination\": true,
               \"SnapshotId\": \"$snapshotid2\",
               \"VolumeSize\": $xvdbvolsize,
               \"VolumeType\": \"gp2\"
           }
        },{
           \"DeviceName\": \"xvdc\",
           \"VirtualName\": \"ephemeral0\"
        },{
           \"DeviceName\": \"xvdd\",
           \"VirtualName\": \"ephemeral1\"
        }]
    }" "$url?env_id=$ENVID&region=$REGION" \
    >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    dummyval=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}'
    )

    [[ -z $dummyval ]] && {
        echo
        echo "ERROR: dummyval is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}'
EnD
        cat $e
        exit 1
    }

    AMIID=$(echo "$dummyval" | grep -m1 -Eo 'ami-[[:alnum:]]+')

    echo "`date` AMIID=$AMIID"
}

# ---------------------------------------------------------------------------
get_snapshotids_from_image() {
# ---------------------------------------------------------------------------
# Get all snapshotids from an AMI
# Arg1 - image id
# Sets global var: SNAPSHOTIDS

    local imageid=$1

    echo "`date` Querying instance $instanceid for device /dev/$device..."

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/describe-images"
    
    curl $OPTSD \
        "$url?env_id=$ENVID&region=$REGION&filter=image-id=$imageid" \
        >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    SNAPSHOTIDS=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
        | python -mjson.tool \
        | grep -i '"snapshotid":' | grep -o 'snap-[[:alnum:]]*'
    )

    [[ -z $SNAPSHOTIDS ]] && {
        echo
        echo "ERROR: SNAPSHOTIDS is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//gp}' \
| python -mjson.tool \
| grep -i '"snapshotid":' | grep -o 'snap-[[:alnum:]]*'
EnD
        cat $e
        exit 1
    }

    echo "`date` Snapshot IDs for $imageid: " $SNAPSHOTIDS
}

# ---------------------------------------------------------------------------
deregister_image() {
# ---------------------------------------------------------------------------
# Delete a Volume
# Arg1 - image id
# Sets global var: none

    local imageid=$1 discardedvalue

    echo "`date` Deregistering image, $imageid"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/deregister-image"

    curl $OPTS -d "
    {
        \"ImageId\": \"$imageid\"
    }" \
    "$url?env_id=$ENVID&region=$REGION" \
        >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    discardedvalue=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//g;p}'
    )

    [[ -z $discardedvalue ]] && {
        # Just let the error go!
        echo "`date` ERROR: Delete snapshot FAILED!"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\\\//g;p}'
EnD
        cat $e
        exit 1
    }
    
    echo "`date` AMI, $imageid, deregistered"
}

# ---------------------------------------------------------------------------
create_placement_group() {
# ---------------------------------------------------------------------------
# Create a new placement group
# Arg1 - placement group name
# Sets global var: none

    local groupname=$1

    echo "`date` Creating a new placement group"

    url="$PROTO://$IPPORT/api/$USERID/$GUID/aws-ec2lib/create-placement-group"

    curl $OPTSD -d "
    {
        \"GroupName\":\"$groupname\",
        \"Strategy\":\"cluster\"
    }" "$url?env_id=$ENVID&region=$REGION" \
    >$t 2>$e

    [[ $? -ne 0 ]] && show_err_quit

    dummyval=$( cat $t | python -mjson.tool \
        | sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\//g;p}'
    )

    [[ -z $dummyval ]] && {
        echo
        echo "ERROR: dummyval is empty."
        echo
        echo -n "echo '"
        cat $t
        echo -n "' "
        cat <<EnD
| python -mjson.tool \
| sed -n '/"Text":/ {s/^ *"Text": "//;s/"$//;s/\\\\//g;p}'
EnD
        cat $e
        exit 1
    }

    echo "`date` Created placement group: $groupname"
}

# ===========================================================================
# HELPER FUNCTIONS
# ===========================================================================

# ---------------------------------------------------------------------------
ssh_cmd() {
# ---------------------------------------------------------------------------
# Run a remote command and exit 1 on failure.
# Arguments: arg1 - server to connect to
#            arg2 - "savestdout" then save output in LAST_STDOUT
#            arg3 - Command and arguments to run
# Returns: Nothing

    [[ $DEBUG -eq 1 ]] && echo "In ssh_cmd()"

    local -i retval=0 t=0 n=0
    local tmpout="/tmp/tmprsyncbackupout_$$.out"
    local tmperr="/tmp/tmprsyncbackuperr_$$.out"
    local tmpret="/tmp/tmprsyncbackupret_$$.out"
    local tmpechod="/tmp/tmprsyncbackupechod_$$.out"

    if [[ $2 == "savestdout" ]]; then
        trap : INT
        echo "> Running remotely: $3"
        ( ssh -i "$TMPKEYFILE" "$REMOTEUSER@$1" "$3" \
          >$tmpout 2>$tmperr;
          echo $? >$tmpret) & waiton=$!;
        ( t=0;n=0
          while true; do
            [[ $n -eq 2 ]] && {
                echo -n "> Waiting ";
                touch $tmpechod;
            }
            [[ $n -gt 2 ]] && echo -n ".";
            sleep $t;
            t=1; n+=1
          done 2>/dev/null;
        ) & killme=$!
        wait $waiton &>/dev/null
        kill $killme &>/dev/null
        wait $killme 2>/dev/null
        [[ -e $tmpechod ]] && {
            rm -f $tmpechod &>/dev/null
            echo
        }
        retval=`cat $tmpret`
        LAST_STDOUT=`cat $tmpout`
        trap - INT
    else
        LAST_STDOUT=
        trap : INT
        echo "> Running remotely: $3"
        ( ssh -i "$TMPKEYFILE" "$REMOTEUSER@$1" "$3" \
          >$tmpout 2>$tmperr;
          echo $? >$tmpret) & waiton=$!;
        ( t=0;n=0
          while true; do
            [[ $n -eq 2 ]] && {
                echo -n "> Waiting ";
                touch $tmpechod;
            }
            [[ $n -gt 2 ]] && echo -n ".";
            sleep $t;
            t=1; n+=1
          done 2>/dev/null;
        ) & killme=$!
        wait $waiton &>/dev/null
        kill $killme &>/dev/null
        wait $killme 2>/dev/null
        [[ -e $tmpechod ]] && {
            rm -f $tmpechod &>/dev/null
            echo
        }
        retval=`cat $tmpret`
        trap - INT
    fi

    [[ $retval -ne 0 ]] && {
        echo
        echo -e "${R}ERROR$N: Command failed on '$1'. Command was:"
        echo
        echo "  $3"
        echo
        echo "OUTPUT WAS:"
        echo "  $LAST_STDOUT"
        echo "  $(cat $tmperr | sed 's/^/  /')"
        echo
        echo "Cannot continue. Aborting."
        echo
        exit 1
    }
}

# ---------------------------------------------------------------------------
cleanup() {
# ---------------------------------------------------------------------------
    rm -f $TMPKEYFILE &>/dev/null
}

# ---------------------------------------------------------------------------
write_tmpkeyfile() {
# ---------------------------------------------------------------------------

    touch $KNOWNHOSTS

    trap cleanup EXIT

    echo "$PRIVKEYB64" | base64 -d >$TMPKEYFILE
    chmod 0600 $TMPKEYFILE
}

# ---------------------------------------------------------------------------
prime_ssh() {
# ---------------------------------------------------------------------------
# Add server, Arg1, to known_hosts
# Notes: ssh-keyscan will fail until the server is up and ssh is accepting
# connections, hence the 'for' loop.

    local i

    if grep -qs $1 $KNOWNHOSTS; then
        sed -i "/^$1/ {d}" $KNOWNHOSTS
    fi

    for i in 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0; do

        # Add the server to known_hosts
        echo "`date` ssh-keyscan $1 >> $KNOWNHOSTS"
        ssh-keyscan $1 >> $KNOWNHOSTS 2>$e
        if grep -qs "^$1" $KNOWNHOSTS; then
            break
        fi

        echo "`date` Retrying..."

        sleep 5

    done

    if ! grep -qs $1 $KNOWNHOSTS; then
        echo "ERROR: ssh-keyscan failed:"
        cat $e
    fi

    echo "`date` Instance is ready for commands (ssh is up)"
}

# ---------------------------------------------------------------------------
sanity_checks() {
# ---------------------------------------------------------------------------

    for binary in sed date curl python true ssh-keyscan base64; do
        if ! which $binary >& /dev/null; then
            echo "ERROR: $binary binary not found in path. Aborting."
            exit 1
        fi
    done

    [[ -z $ENVID ]] && {
      echo "ERROR: environment variable 'ENVID' must be set."
      exit 1
    }

    [[ -z $REGION ]] && {
      echo "ERROR: environment variable 'REGION' must be set."
      exit 1
    }

    [[ -z $USERID ]] && {
      echo "ERROR: environment variable 'USERID' must be set."
      exit 1
    }

    [[ -z $AVAILZONECHAR ]] && {
      echo "ERROR: environment variable 'AVAILZONECHAR' must be set."
      exit 1
    }

    [[ -z $GUID ]] && {
      echo "ERROR: environment variable 'GUID' must be set."
      exit 1
    }

    if [[ -z $AMINAME ]]; then
      #echo "ERROR: environment variable 'AMINAME' must be set."
      AMINAME="ami-pcs_cache_$(date +%Y%m%d.%H%M%S)"
      #exit 1
    else
      AMINAME="${AMINAME}_$(date +%Y%m%d.%H%M%S)"
    fi
}

# ---------------------------------------------------------------------------
create_temp_file() {
# ---------------------------------------------------------------------------
# Creates two temporary files and a trap to delete them
# Sets global var: t - for stdout, e for stderr

    declare -i i=0

    t="/var/tmp/run_hcom-dse-pcs-cache_$$_$i"
    while [[ -e $t ]]; do
        i=i+1
        t="/var/tmp/run_hcom-dse-pcs-cache_$$_$i"
    done
    touch $t
    [[ $? -ne 0 ]] && {
        echo "Could not create temporary file. Aborting."
        exit 1
    }

    i=0

    e="/var/tmp/run_hcom-dse-pcs-cache_err_$$_$i"
    while [[ -e $e ]]; do
        i=i+1
        e="/var/tmp/run_hcom-dse-pcs-cache_err_$$_$i"
    done
    touch $e
    [[ $? -ne 0 ]] && {
        echo "Could not create temporary file. Aborting."
        exit 1
    }

    trap "rm -f -- '$t'; rm -f -- '$e'" EXIT
}

# ---------------------------------------------------------------------------
show_err_quit() {
# ---------------------------------------------------------------------------
    echo
    echo "ERROR: curl command exited with non-zero exit value."
    echo
    cat $t
    cat $e
    exit 1
}

