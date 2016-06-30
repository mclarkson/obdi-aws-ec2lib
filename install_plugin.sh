#!/bin/bash
#
# Obdi - a REST interface and GUI for deploying software
# Copyright (C) 2014  Mark Clarkson
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#
# helloworld-runscript plugin

[[ -z $guid ]] && {
  echo "ERROR: environment variable 'guid' must be set."
  exit 1
}

proto="https"
opts="-k -s" # don't check ssl cert, silent
ipport="127.0.0.1:443"

#
# Create a temporary file and a trap to delete it
#

t="/tmp/install_aws-p2ec2_$$"
touch $t
[[ $? -ne 0 ]] && {
    echo "Could not create temporary file. Aborting."
    exit 1
}
trap "rm -f -- '$t'" EXIT

#
# Create a custom environment capability so other plugins
# know we're available.
#

curl -k -d '{
    "Code":"HAS_AWS_P2EC2",
    "Desc":"Has Physical files to EC2 plugin (AWS-P2EC2).",
    "IsWorkerDef":false
}' $proto://$ipport/api/admin/$guid/envcaps

#
# Create environment capabilities for AWS credentials
#

curl -k -d '{
    "Code":"AWS_ACCESS_KEY_ID_1",
    "Desc":"AWS credential. For example: ALIBJX4KM6OPZEEW6B2Q",
    "IsJsonObjectDef":true
}' $proto://$ipport/api/admin/$guid/envcaps

curl -k -d '{
    "Code":"AWS_SECRET_ACCESS_KEY_1",
    "Desc":"AWS credential. For example: MY-SECRET-KEY.",
    "IsJsonObjectDef":true
}' $proto://$ipport/api/admin/$guid/envcaps

#
# Create the plugin entry in obdi, but it's not shown in the sidebar.
#

curl -k -d '{
    "Name":"aws-p2ec2",
    "Desc":"AWS Physical to EC2 plugin.",
    "HasView":0,
    "Parent":""
}' $proto://$ipport/api/admin/$guid/plugins | tee $t

# Grab the id of the last insert
id=`grep Id $t | grep -Eo "[0-9]+"`

#
# Add the AJS controller files
#
# These need to be loaded when the application starts
#

curl -k -d '{
    "Name":"aws-p2ec2.js",
    "Desc":"Controller for AWS-P2EC2.",
    "Type":1,
    "PluginId":'"$id"',
    "Url":"aws-p2ec2/js/controllers/aws-p2ec2.js"
}' $proto://$ipport/api/admin/$guid/files

#
# Add the scripts, removing comment lines (#) and empty lines
#

#script="backup.sh"
#
#source=`sed '1n;/^\s*#/d;/^$/d;' scripts/$script | base64 -w 0`
#
#curl -k $proto://$ipport/api/admin/$guid/scripts?name=$script | tee $t
#
## Grab the id of the last insert
#id=`grep Id $t | grep -Eo "[0-9]+"`
#
#if [[ -z $id ]]; then
#	curl -k -d '{
#		"Desc": "Rsync backup script. Only the verbose option, -V, is used here.",
#		"Name": "'"$script"'",
#		"Source": "'"$source"'"
#	}' $proto://$ipport/api/admin/$guid/scripts
#else
#	curl -k -X PUT -d '{ "Source": "'"$source"'" }' \
#	$proto://$ipport/api/admin/$guid/scripts/$id
#fi

# --

# Delete the temporary file and delete the trap
rm -f -- "$t"
trap - EXIT

