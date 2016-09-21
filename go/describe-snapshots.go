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

// GOPATH=$PWD go build describe-snapshots.go obdi_clientlib.go

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
	"strings"
)

var region = "us-east-1"

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

	// Describe the volume

	creds := credentials.NewStaticCredentials(
		awsdata.Aws_access_key_id,
		awsdata.Aws_secret_access_key,
		"")
	config := aws.Config{
		Region:      aws.String(region),
		Credentials: creds,
	}

	svc := ec2.New(session.New(), &config)

	params := &ec2.DescribeSnapshotsInput{}

	// dry_run

	if len(args.QueryString["dry_run"]) > 0 {
		if args.QueryString["dry_run"][0] == "true" {
			params.DryRun = aws.Bool(true)
		} else {
			params.DryRun = aws.Bool(false)
		}
	}

	// filter

	if len(args.QueryString["filter"]) > 0 {
		for i := range args.QueryString["filter"] {
			keyval := strings.SplitN(args.QueryString["filter"][i], "=", 2)
			if len(keyval) != 2 {
				txt := fmt.Sprintf("Error: Invalid filter expression ('%s'). No '=' found.",
					args.QueryString["filter"][i])
				ReturnError(txt, response)
				return nil
			}
			var vals []*string
			items := strings.Split(keyval[1], ",")
			for j := range items {
				vals = append(vals, &items[j])
			}
			flt := ec2.Filter{Name: &keyval[0], Values: vals}
			params.Filters = append(params.Filters, &flt)
		}
	}

	// owner_id

	if len(args.QueryString["owner_id"]) > 0 {
		for i := range args.QueryString["owner_id"] {
			params.OwnerIds = append(params.OwnerIds,
				&args.QueryString["owner_id"][i])
		}
	}

	// snapshot_id

	if len(args.QueryString["snapshot_id"]) > 0 {
		for i := range args.QueryString["snapshot_id"] {
			params.SnapshotIds = append(params.SnapshotIds,
				&args.QueryString["snapshot_id"][i])
		}
	}

	// restorable_by_user_id

	if len(args.QueryString["restorable_by_user_id"]) > 0 {
		for i := range args.QueryString["restorable_by_user_id"] {
			params.RestorableByUserIds = append(params.RestorableByUserIds,
				&args.QueryString["restorable_by_user_id"][i])
		}
	}

	// max_results

	if len(args.QueryString["max_results"]) > 0 {
		MaxResults, _ := strconv.ParseInt(args.QueryString["max_results"][0], 10, 64)
		params.MaxResults = &MaxResults
	}

	// next_token

	if len(args.QueryString["max_results"]) > 0 {
		params.NextToken = &args.QueryString["next_token"][0]
	}

	resp, err := svc.DescribeSnapshots(params)

	if err != nil {
		t := "Error running DescribeSnapshots: " + err.Error()
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

func (t *Plugin) PostRequest(args *Args, response *[]byte) error {

	// POST requests can change state

	ReturnError("Internal error: Unimplemented HTTP POST", response)
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
