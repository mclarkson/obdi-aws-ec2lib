package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"log/syslog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	SUCCESS = 0
	ERROR   = 1
)

type ApiError struct {
	details string
}

func (e ApiError) Error() string {
	return fmt.Sprintf("%s", e.details)
}

// For sending a job to the Manager
type Job struct {
	Id            int64
	ScriptId      int64
	Args          string // E.g. `-a -f "bob 1" name`
	EnvVars       string // E.g. `A:1 B:"Hi there" C:3`
	Status        int64
	StatusReason  string
	StatusPercent int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     time.Time
	UserLogin     string
	Errors        int64
	EnvId         int64  // For WorkerUrl and WorkerKey
	EnvCapDesc    string // For WorkerUrl and WorkerKey, e.g. "SALT_WORKER"
	Type          int64  // 1 - user job, 2 - system job
}

// For retrieving details from the Manager
type Env struct {
	Id       int64
	DispName string // Display name
	SysName  string // System name (Salt name)
	/*Dc          Dc*/ // only for creating Env and substruct
	DcId               int64
	DcSysName          string
	//WorkerIp    string      // Hostname or IP address of worker
	//WorkerPort  string      // Port the worker listens on
	WorkerUrl string // Worker URL Prefix
	WorkerKey string // Key (password) for worker
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

// For retrieving details from the Manager
type Script struct {
	Id        int64
	Name      string
	Desc      string
	Source    []byte
	Type      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt time.Time
}

// Args are send over RPC from the Manager
type Args struct {
	PathParams  map[string]string
	QueryString map[string][]string
	PostData    []byte
	QueryType   string
}

type Plugin struct{}

// The reply will be sent and output by the master
type Reply struct {
	// Add more if required
	JobId int64
	Text  string
	// Must have the following
	PluginReturn int64 // 0 - success, 1 - error
	PluginError  string
}

type ScriptArgs struct {
	ScriptName string
	CmdArgs    string
	EnvVars    string
	EnvCapDesc string
	Type       int64
}

func ReturnError(text string, response *[]byte) {

	errtext := Reply{0, "", ERROR, text}
	logit(text)
	jsondata, _ := json.Marshal(errtext)
	*response = jsondata
}

// ***************************************************************************
// SUPPORT FUNCS
// ***************************************************************************

func logit(msg string) {

	// Log to syslog

	log.Println(msg)
	l, err := syslog.New(syslog.LOG_ERR, "obdi")
	defer l.Close()
	if err != nil {
		log.Fatal("error writing syslog!")
	}

	l.Err(msg)
}

func GET(url, endpoint string) (r *http.Response, e error) {

	// Send HTTP GET request

	// accept bad certs
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// Not available in Go<1.3
	//client.Timeout = 8 * 1e9

	//fmt.Printf("\n%s/api/%s\n",url,endpoint)
	for strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	//fmt.Printf( "%s\n", url+"/"+endpoint )
	resp, err := client.Get(url + "/" + endpoint)
	if err != nil {
		txt := fmt.Sprintf("Could not send REST request ('%s').", err.Error())
		return resp, ApiError{txt}
	}

	if resp.StatusCode != 200 {
		var body []byte
		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			txt := fmt.Sprintf("Error reading Body ('%s').", err.Error())
			return resp, ApiError{txt}
		} else {
			body = b
		}
		type myErr struct {
			Error string
		}
		errstr := myErr{}
		if err := json.Unmarshal(body, &errstr); err != nil {
			txt := fmt.Sprintf("Error decoding JSON "+
				"returned from worker - (%s). Check the Worker URL.",
				err.Error())
			return resp, ApiError{txt}
		}

		//txt := fmt.Sprintf("%s", resp.StatusCode)
		return resp, ApiError{errstr.Error}
	}

	return resp, nil
}

func POST(jsondata []byte, url, endpoint string) (r *http.Response, e error) {

	// Send HTTP POST request

	buf := bytes.NewBuffer(jsondata)

	// accept bad certs
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// Not available in Go<1.3
	//client.Timeout = 8 * 1e9

	//fmt.Printf("\n%s/api/%s\n",url,endpoint)
	for strings.HasSuffix(url, "/") {
		url = strings.TrimSuffix(url, "/")
	}
	//fmt.Printf( "%s\n", url+"/"+endpoint )
	resp, err := client.Post(url+"/"+endpoint, "application/json", buf)
	if err != nil {
		txt := fmt.Sprintf("Could not send REST request ('%s').", err.Error())
		return resp, ApiError{txt}
	}

	if resp.StatusCode != 200 {
		var body []byte
		if b, err := ioutil.ReadAll(resp.Body); err != nil {
			txt := fmt.Sprintf("Error reading Body ('%s').", err.Error())
			return resp, ApiError{txt}
		} else {
			body = b
		}
		type myErr struct {
			Error string
		}
		errstr := myErr{}
		if err := json.Unmarshal(body, &errstr); err != nil {
			txt := fmt.Sprintf("Error decoding JSON "+
				"returned from worker - (%s). Check the Worker URL.",
				err.Error())
			return resp, ApiError{txt}
		}

		//txt := fmt.Sprintf("%s", resp.StatusCode)
		return resp, ApiError{errstr.Error}
	}

	return resp, nil
}

func (t *Plugin) GetAllowedEnv(args *Args, env_id_str string, response *[]byte) (Env, error) {

	// Get the Env (SysName) for this env_id using REST.
	// The Environment name (e.g. dev) is stored in:
	//   envs[0].SysName
	// GET queries always return an array of items, even for 1 item.
	envs := []Env{}
	resp, _ := GET("https://127.0.0.1/api/"+
		args.PathParams["login"]+"/"+args.PathParams["GUID"], "envs"+
		"?env_id="+env_id_str)
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		txt := fmt.Sprintf("Error reading Body ('%s').", err.Error())
		ReturnError(txt, response)
		return Env{}, ApiError{"Error"}
	} else {
		json.Unmarshal(b, &envs)
	}
	// If envs is empty then we don't have permission to see it
	// or the env does not exist so bug out.
	if len(envs) == 0 {
		txt := "The requested environment id does not exist" +
			" or the permissions to access it are insufficient."
		ReturnError(txt, response)
		return Env{}, ApiError{"Error"}
	}

	return envs[0], nil
}

func (t *Plugin) RunScript(args *Args, sa ScriptArgs, response *[]byte) (int64, error) {

	// Check for required query string entries

	if len(args.QueryString["env_id"]) == 0 {
		ReturnError("'env_id' must be set", response)
		return 0, ApiError{"Error"}
	}

	env_id, _ := strconv.ParseInt(args.QueryString["env_id"][0], 10, 64)

	// Get the ScriptId from the scripts table for:
	scriptName := sa.ScriptName
	scripts := []Script{}
	resp, err := GET("https://127.0.0.1/api/"+
		args.PathParams["login"]+"/"+args.PathParams["GUID"], "scripts"+
		"?nosource=1&name="+scriptName)
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		txt := fmt.Sprintf("Error reading Body ('%s').", err.Error())
		ReturnError(txt, response)
		return 0, ApiError{"Error"}
	} else {
		json.Unmarshal(b, &scripts)
	}
	// If scripts is empty then we don't have permission to see it
	// or the script does not exist (well, scripts don't have permissions
	// but lets say the same thing anyway)
	if len(scripts) == 0 {
		txt := "The requested script, '" + scriptName + "', does not exist" +
			" or the permissions to access it are insufficient."
		ReturnError(txt, response)
		return 0, ApiError{"Error"}
	}

	// Set up some fields for the Job struct we'll send to the master
	job := Job{
		ScriptId:   scripts[0].Id,
		EnvId:      env_id,
		EnvCapDesc: sa.EnvCapDesc,
		EnvVars:    sa.EnvVars,
		Args:       sa.CmdArgs,

		// Type 1 - User Job - Output is
		//     sent back as it's created
		// Type 2 - System Job - All output
		//     is saved in one single line.
		//     Good for json etc.
		Type: sa.Type,
	}

	// Send the job POST request to the master
	jsonjob, err := json.Marshal(job)
	resp, err = POST(jsonjob, "https://127.0.0.1/api/"+
		args.PathParams["login"]+"/"+args.PathParams["GUID"], "jobs")
	if err != nil {
		txt := "Could not send job to worker. ('" + err.Error() + "')"
		ReturnError(txt, response)
		return 0, ApiError{"Error"}
	}
	defer resp.Body.Close()
	// Read the worker's response from the master
	if b, err := ioutil.ReadAll(resp.Body); err != nil {
		txt := fmt.Sprintf("Error reading Body ('%s').", err.Error())
		ReturnError(txt, response)
		return 0, ApiError{"Error"}
	} else {
		json.Unmarshal(b, &job)
	}

	// Send the Job ID as the RPC reply (back to the master)

	return job.Id, nil
}

// --

var config *Config

type Config struct {
	Dbname   string
	Portlock *PortLock
	Port     int
}

func (c *Config) DBPath() string {

	return c.Dbname
}

func (c *Config) SetDBPath(path string) {

	c.Dbname = path
}

func NewConfig() {

	config = &Config{}
}

// --

type GormDB struct {
	db *gorm.DB
}

func (gormInst *GormDB) DB() *gorm.DB {

	return gormInst.db
}

func NewDB(args *Args, dbname string) (*GormDB, error) {

	var err error
	var gormInst GormDB

	// PluginDatabasePath is required to open our private db
	if len(args.PathParams["PluginDatabasePath"]) == 0 {
		return &gormInst,
			ApiError{"Internal Error: 'PluginDatabasePath' must be set."}
	}

	config.SetDBPath(args.PathParams["PluginDatabasePath"])
	dbpath := config.DBPath()

	// Open the database file
	gormInst.db, err = gorm.Open("sqlite3", dbpath+dbname)
	if err != nil {
		return &gormInst,
			ApiError{"Open " + dbpath + dbname + " failed. " + err.Error()}
	}

	if err = gormInst.InitDB(dbname); err != nil {
		return &gormInst, err
	}

	return &gormInst, nil
}

// ***************************************************************************
// PORT LOCKING
// ***************************************************************************

// PortLock is a locker which locks by binding to a port on the loopback IPv4
// interface
type PortLock struct {
	hostport string
	ln       net.Listener
}

func NewPortLock(port int) *PortLock {

	// NewFLock creates new Flock-based lock (unlocked first)
	return &PortLock{hostport: net.JoinHostPort("127.0.0.1", strconv.Itoa(port))}
}

func (p *PortLock) Lock() {

	// Lock acquires the lock, blocking
	t := 50 * time.Millisecond
	for {
		if l, err := net.Listen("tcp", p.hostport); err == nil {
			p.ln = l // thanks to zhangpy
			return
		}
		//log.Printf("spinning lock on %s (%s)", p.hostport, err)
		time.Sleep(t)
		//t = time.Duration(
		//  math.Min( float64(time.Duration(float32(t) * 1.5)), 2000 ))
	}
}

func (p *PortLock) Unlock() {

	// Unlock releases the lock
	if p.ln != nil {
		p.ln.Close()
	}
}

func Unlock() {

	config.Portlock.Unlock()
}

func Lock() {

	config.Portlock.Lock()
}
