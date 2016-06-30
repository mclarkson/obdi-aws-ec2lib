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
	"net"
	"net/rpc"
	"os"
)

// The format of the json sent by the client in a POST request
type PostedData struct {
	Grain string
	Text  string
}

// Name of the sqlite3 database file
const DBFILE = "rsyncbackup.db"

// The 'tasks' table
type Task struct {
	Id       int64
	TaskDesc string
	CapTag   string
	DcId     int64 // Data centre name
	EnvId    int64 // Environment name
}

// The 'includes' table
type Include struct {
	Id     int64
	TaskId int64
	Host   string
	Base   string // Data centre name
}

// The 'excludes' table
type Exclude struct {
	Id        int64
	IncludeId int64
	Path      string
}

// The 'settings' table
type Setting struct {
	Id         int64
	TaskId     int64
	Protocol   string
	Pre        string
	RsyncOpts  string
	BaseDir    string
	KnownHosts string
	NumPeriods int64
	Timeout    int64
	Verbose    bool
}

// Create tables and indexes in InitDB
func (gormInst *GormDB) InitDB(dbname string) error {

	// We just read the database so empty

	return nil
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

func (t *Plugin) GetRequest(args *Args, response *[]byte) error {

	// GET requests don't change state, so, don't change state

	// env_id is required, '?env_id=xxx'

	if len(args.QueryString["env_id"]) == 0 {
		ReturnError("'env_id' must be set", response)
		return nil
	}

	env_id_str := args.QueryString["env_id"][0]

	// task_id is required, '?task_id=xxx'

	if len(args.QueryString["task_id"]) == 0 {
		ReturnError("'task_id' must be set", response)
		return nil
	}

	task_id_str := args.QueryString["task_id"][0]

	// path is optional, '?path=xxx'

	path_str := ""
	if len(args.QueryString["path"]) > 0 {
		path_str = args.QueryString["path"][0]
	}

	// snapshot is optional, '?snapshot=', '?snapshot=20160901.1'

	snapshot_dir := ""
	if len(args.QueryString["snapshot"]) > 0 {
		if len(args.QueryString["snapshot"][0]) > 0 {
			snapshot_dir = ".zfs/snapshot/"
			snapshot_dir += args.QueryString["snapshot"][0] + "/"
		}
	}

	// Check if the user is allowed to access the environment
	var err error
	if _, err = t.GetAllowedEnv(args, env_id_str, response); err != nil {
		// GetAllowedEnv wrote the error
		return nil
	}

	// Setup/Open the local database

	var gormdb *GormDB
	if gormdb, err = NewDB(args, DBFILE); err != nil {
		txt := "NewDB open error for '" + config.DBPath() + DBFILE + "'. " +
			err.Error()
		ReturnError(txt, response)
		return nil
	}

	db := gormdb.DB() // for convenience

	// We got this far so access is allowed.

	// Get the task to get the CapabilityTag
	task := Task{}
	Lock()
	if err = db.First(&task, "id = ?", task_id_str).Error; err != nil {
		Unlock()
		ReturnError("Query error. "+err.Error(), response)
		return nil
	}
	Unlock()

	// Get settings to get the BaseDir

	var setting Setting

	Lock()
	if err = db.First(&setting, "task_id = ?", task_id_str).Error; err != nil {
		Unlock()
		ReturnError("Query error. "+err.Error(), response)
		return nil
	}
	Unlock()

	sa := ScriptArgs{
		// The name of the script to send an run
		ScriptName: "rsyncbackup-ls.sh",
		// The arguments to use when running the script
		CmdArgs: setting.BaseDir + "/" + snapshot_dir + path_str,
		// Environment variables to pass to the script
		EnvVars: "", //`A=1 B=2 C='a b c' D=44`,
		// Name of an environment capability (where isworkerdef == true)
		// that can point to a worker other than the default.
		EnvCapDesc: task.CapTag,
		// Type 1 - User Job - Output is
		//     sent back as it's created
		// Type 2 - System Job - All output
		//     is saved in one single line.
		//     Good for json etc.
		Type: 2,
	}

	var jobid int64
	if jobid, err = t.RunScript(args, sa, response); err != nil {
		// RunScript wrote the error so just return
		return nil
	}

	reply := Reply{jobid, "", SUCCESS, ""}
	jsondata, err := json.Marshal(reply)
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

	// Decode the post data into struct

	var postedData PostedData

	if err := json.Unmarshal(args.PostData, &postedData); err != nil {
		txt := fmt.Sprintf("Error decoding JSON ('%s')"+".", err.Error())
		ReturnError("Error decoding the POST data ("+
			fmt.Sprintf("%s", args.PostData)+"). "+txt, response)
		return nil
	}

	// Use salt to change the version, if it's changed

	if len(args.QueryString["var_a"]) == 0 {
		ReturnError("'var_a' must be set", response)
		return nil
	}

	sa := ScriptArgs{
		// The name of the script to send an run
		ScriptName: "helloworldrunscript-sets.sh",
		// The arguments to use when running the script
		CmdArgs: args.QueryString["var_a"][0] + " " +
			postedData.Grain + "," + postedData.Text,
		// Environment variables to pass to the script
		EnvVars: "",
		// Name of an environment capability (where isworkerdef == true)
		// that can point to a worker other than the default.
		EnvCapDesc: "HELLOWORLD_RUNSCRIPT_WORKER",
		// Type 1 - User Job - Output is
		//     sent back as it's created
		// Type 2 - System Job - All output
		//     is saved in one single line.
		//     Good for json etc.
		Type: 2,
	}

	var jobid int64
	if len(postedData.Grain) > 0 && len(postedData.Text) > 0 {
		var err error
		if jobid, err = t.RunScript(args, sa, response); err != nil {
			// RunScript wrote the error so just return
			return nil
		}
	} else {
		ReturnError("No POST data received. Nothing to do.", response)
		return nil
	}

	reply := Reply{jobid, "", SUCCESS, ""}
	jsondata, err := json.Marshal(reply)
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
	config.Port = 49995
	config.Portlock = NewPortLock(config.Port)

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
