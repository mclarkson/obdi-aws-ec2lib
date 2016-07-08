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
	"errors"
	"fmt"
	"github.com/goamz/goamz/aws"
	"github.com/goamz/goamz/ec2"
	"io/ioutil"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"strings"
	"time"
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

func instances(e *ec2.EC2, response *[]byte, args ...string) ([]ec2.Reservation, error) {
	filter := ec2.NewFilter()
	for _, v := range args {
		sl := strings.SplitN(v, "=", 2)
		if len(sl) != 2 {
			t := "Bad key=value pair: " + v
			ReturnError(t, response)
			return nil, errors.New("")
		}
		filter.Add(sl[0], sl[1])
	}

	resp, err := e.DescribeInstances(nil, filter)
	if err != nil {
		t := "instances: " + err.Error()
		ReturnError(t, response)
		return nil, errors.New("")
	}

	return resp.Reservations, nil
}

func (t *Plugin) GetRequest(args *Args, response *[]byte) error {

	// GET requests don't change state, so, don't change state

	// env_id is required, '?env_id=xxx'

	if len(args.QueryString["env_id"]) == 0 {
		ReturnError("'env_id' must be set", response)
		return nil
	}

	env_id_str := args.QueryString["env_id"][0]

	// region is optional, '?region=xxx'

	if len(args.QueryString["region"]) > 0 {
		region = args.QueryString["region"][0]
	}

	r, ok := aws.Regions[region]
	if !ok {
		t := "Unknown region: " + region
		ReturnError(t, response)
		return nil
	}

	// Get aws_access_key_id and aws_secret_access_key from
	// the AWS_ACCESS_KEY_ID_1 capability using sdtoken

	envcaps := []EnvCap{}
	{
		resp, _ := GET("https://127.0.0.1/api/" + args.PathParams["login"] +
			"/" + args.PathParams["GUID"], "/envcaps?code=AWS_ACCESS_KEY_ID_1")
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
		resp, _ := GET("https://127.0.0.1/api/sduser/" + args.SDToken,
			"/jsonobjects?env_id=" + env_id_str + "&env_cap_id=" + env_cap_id_str)
		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			txt := fmt.Sprintf("Error reading Body ('%s').", err.Error())
			ReturnError(txt, response)
			return nil
		} else {
			json.Unmarshal(b, &json_objects)
		}
	}

	type AWSData struct {
		Aws_access_key_id string
		Aws_secret_access_key string
	}

	awsdata := AWSData{}
	if err := json.Unmarshal([]byte(json_objects[0].Json), &awsdata); err != nil {
		txt := fmt.Sprintf("Error decoding JsonObject ('%s').", err.Error())
		ReturnError(txt, response)
		return nil
	}

	auth, err := aws.GetAuth( awsdata.Aws_access_key_id, awsdata.Aws_secret_access_key,
	"",time.Time{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "can not find auth info: %s\n", err)
		t := "can not find auth info: " + err.Error()
		ReturnError(t, response)
		return nil
	}

	e := ec2.New(auth, r)

	instances_json, err := instances(e, response)
	if err != nil {
		return nil
	}

	jsondata, err := json.Marshal(instances_json)
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
