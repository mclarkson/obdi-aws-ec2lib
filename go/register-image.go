// Obdi - a REST interface and GUI for deploying software
// Copyright (C) 2014  Mark Clarkson
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"strconv"
)

var region = "us-east-1"

// The format of the json sent by the client in a POST request
type PostedData struct {
	// The architecture of the AMI.
	//
	// Default: For Amazon EBS-backed AMIs, i386. For instance store-backed AMIs,
	// the architecture specified in the manifest file.
	Architecture string

	// One or more block device mapping entries.
	BlockDeviceMappings []*BlockDeviceMapping

	// A description for your AMI.
	Description string

	// Checks whether you have the required permissions for the action, without
	// actually making the request, and provides an error response. If you have
	// the required permissions, the error response is DryRunOperation. Otherwise,
	// it is UnauthorizedOperation.
	DryRun bool

	// Set to true to enable enhanced networking with ENA for the AMI and any instances
	// that you launch from the AMI.
	//
	// This option is supported only for HVM AMIs. Specifying this option with
	// a PV AMI can make instances launched from the AMI unreachable.
	//EnaSupport *bool `locationName:"enaSupport" type:"boolean"`

	// The full path to your AMI manifest in Amazon S3 storage.
	//ImageLocation *string `type:"string"`

	// The ID of the kernel.
	//KernelId *string `locationName:"kernelId" type:"string"`

	// A name for your AMI.
	//
	// Constraints: 3-128 alphanumeric characters, parentheses (()), square brackets
	// ([]), spaces ( ), periods (.), slashes (/), dashes (-), single quotes ('),
	// at-signs (@), or underscores(_)
	//Name *string `locationName:"name" type:"string" required:"true"`

	// The ID of the RAM disk.
	//RamdiskId *string `locationName:"ramdiskId" type:"string"`

	// The name of the root device (for example, /dev/sda1, or /dev/xvda).
	//RootDeviceName *string `locationName:"rootDeviceName" type:"string"`

	// Set to simple to enable enhanced networking with the Intel 82599 Virtual
	// Function interface for the AMI and any instances that you launch from the
	// AMI.
	//
	// There is no way to disable sriovNetSupport at this time.
	//
	// This option is supported only for HVM AMIs. Specifying this option with
	// a PV AMI can make instances launched from the AMI unreachable.
	//SriovNetSupport *string `locationName:"sriovNetSupport" type:"string"`

	// The type of virtualization.
	//
	// Default: paravirtual
	//VirtualizationType *string `locationName:"virtualizationType" type:"string"`
}

// ***************************************************************************
// GO RPC PLUGIN
// ***************************************************************************

/*
 * For running scripts, the script and arguments etc. are sent to the target
 * Obdi worker as an Obdi job. The job ID is returned to the client straight
 * away and then the job is run. The client should poll the jobs table to
 * see when the job finishes. When the job is finished, the output will be
 * available to the client in the outputlines table.
 */

// Create tables and indexes in InitDB
func (gormInst *GormDB) InitDB(dbname string) error {

	// We just read the database so empty

	return nil
}

func (t *Plugin) GetRequest(args *Args, response *[]byte) error {

	// GET requests don't change state, so, don't change state

	ReturnError("Internal error: Unimplemented HTTP GET", response)
	return nil
}

func (t *Plugin) PostRequest(args *Args, response *[]byte) error {

	// POST requests can change state

	// env_id is required, '?env_id=xxx'

	if len(args.QueryString["env_id"]) == 0 {
		ReturnError("'env_id' must be set", response)
		return nil
	}

	env_id_str := args.QueryString["env_id"][0]

	// region is required, '?region=xxx'

	if len(args.QueryString["region"]) == 0 {
		ReturnError("'region' must be set", response)
		return nil
	}

	region := args.QueryString["region"][0]

	// availability_zone is required, '?availability_zone=xxx'

	if len(args.QueryString["availability_zone"]) == 0 {
		ReturnError("'availability_zone' must be set", response)
		return nil
	}

	availzone := args.QueryString["availability_zone"][0]

	// Decode the post data into struct

	var postedData PostedData

	if err := json.Unmarshal(args.PostData, &postedData); err != nil {
		txt := fmt.Sprintf("Error decoding JSON ('%s')"+".", err.Error())
		ReturnError("Error decoding the POST data ("+
			fmt.Sprintf("%s", args.PostData)+"). "+txt, response)
		return nil
	}

	// Get aws_access_key_id and aws_secret_access_key from
	// the AWS_ACCESS_KEY_ID_1 capability using sdtoken

	envcaps := []EnvCap{}
	{
		resp, _ := GET("https://127.0.0.1/api/"+args.PathParams["login"]+
			"/"+args.PathParams["GUID"], "/envcaps?code=AWS_ACCESS_KEY_ID_1")
		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			txt := fmt.Sprintf("Error reading Body ('%s').", err.Error())
			ReturnError(txt, response)
			return nil
		} else {
			json.Unmarshal(b, &envcaps)
		}
	}

	json_objects := []JsonObject{}
	{
		env_cap_id_str := strconv.FormatInt(envcaps[0].Id, 10)
		resp, _ := GET("https://127.0.0.1/api/sduser/"+args.SDToken,
			"/jsonobjects?env_id="+env_id_str+"&env_cap_id="+env_cap_id_str)
		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			txt := fmt.Sprintf("Error reading Body ('%s').", err.Error())
			ReturnError(txt, response)
			return nil
		} else {
			json.Unmarshal(b, &json_objects)
		}
	}

	type AWSData struct {
		Aws_access_key_id           string // E.g. OKLAJX2KN6OXZEFV4B1Q
		Aws_secret_access_key       string // E.g. oiwjeg^laGDIUsg@jfa
		Aws_obdi_worker_instance_id string // E.g. i-e19ec362
		Aws_obdi_worker_region      string // E.g. us-east-1
		Aws_obdi_worker_url         string // E.g. https://1.2.3.4:4443/
		Aws_obdi_worker_key         string // E.g. secretkey
		Aws_filter                  string // E.g. key-name=itsupkey
	}

	awsdata := AWSData{}
	if err := json.Unmarshal([]byte(json_objects[0].Json), &awsdata); err != nil {
		txt := fmt.Sprintf("Error decoding JsonObject ('%s').", err.Error())
		ReturnError(txt, response)
		return nil
	}

	// Create the volume

	creds := credentials.NewStaticCredentials(
		awsdata.Aws_access_key_id,
		awsdata.Aws_secret_access_key,
		"")
	config := aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	}

	svc := ec2.New(session.New(), &config)

	params := &ec2.RegisterImageInput{
		Name:         aws.String("String"), // Required
		Architecture: aws.String("ArchitectureValues"),
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{ // Required
				DeviceName: aws.String("String"),
				Ebs: &ec2.EbsBlockDevice{
					DeleteOnTermination: aws.Bool(true),
					Encrypted:           aws.Bool(true),
					Iops:                aws.Int64(1),
					SnapshotId:          aws.String("String"),
					VolumeSize:          aws.Int64(1),
					VolumeType:          aws.String("VolumeType"),
				},
				NoDevice:    aws.String("String"),
				VirtualName: aws.String("String"),
			},
			// More values...
		},
		Description:        aws.String("String"),
		DryRun:             aws.Bool(true),
		EnaSupport:         aws.Bool(true),
		ImageLocation:      aws.String("String"),
		KernelId:           aws.String("String"),
		RamdiskId:          aws.String("String"),
		RootDeviceName:     aws.String("String"),
		SriovNetSupport:    aws.String("String"),
		VirtualizationType: aws.String("String"),
	}

	resp, err := svc.RegisterImage(params)

	if err != nil {
		t := "Error running RegisterImage: " + err.Error()
		ReturnError(t, response)
		return nil
	}

	jsondata, err := json.Marshal(resp)
	if err != nil {
		ReturnError("Marshal error: "+err.Error(), response)
		return nil
	}
	reply := Reply{0, string(jsondata), SUCCESS, ""}
	jsondata, err = json.Marshal(reply)
	if err != nil {
		ReturnError("Marshal error: "+err.Error(), response)
		return nil
	}

	*response = jsondata

	return nil
}

func (t *Plugin) HandleRequest(args *Args, response *[]byte) error {

	// All plugins must have this.

	if len(args.QueryType) > 0 {
		switch args.QueryType {
		case "GET":
			t.GetRequest(args, response)
			return nil
		case "POST":
			t.PostRequest(args, response)
			return nil
		}
		ReturnError("Internal error: Invalid HTTP request type for this plugin "+
			args.QueryType, response)
		return nil
	} else {
		ReturnError("Internal error: HTTP request type was not set", response)
		return nil
	}
}

func main() {

	//logit("Plugin starting")

	// Sets the global config var, needed for PluginDatabasePath
	NewConfig()

	// Create a lock file to use for synchronisation
	//config.Port = 49995
	//config.Portlock = NewPortLock(config.Port)

	plugin := new(Plugin)
	rpc.Register(plugin)

	listener, err := net.Listen("tcp", ":"+os.Args[1])
	if err != nil {
		txt := fmt.Sprintf("Listen error. %s", err)
		logit(txt)
	}

	//logit("Plugin listening on port " + os.Args[1])

	if conn, err := listener.Accept(); err != nil {
		txt := fmt.Sprintf("Accept error. %s", err)
		logit(txt)
	} else {
		//logit("New connection established")
		rpc.ServeConn(conn)
	}
}
