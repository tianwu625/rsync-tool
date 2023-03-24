package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"

	xclient "github.com/openfs/rsync-tools/client"
	xsnap "github.com/openfs/rsync-tools/snapshot"
)

const (
	usageHost     = "hostname of server include ip"
	usageUsername = "username for login"
	usagePassword = "password for login"
	usageSid      = "source snapshot id"
	usageDid      = "destination snapshot id"
	usagePath     = "compare file path"
)

const (
	defaultPath = "/"
	defaultUser = "admin"
	defaultPass = "admin"
)

const (
	keyHost        = "HOST"
	keyUsername    = "USERNAME"
	keyPassword    = "PASSWORD"
	keySid         = "SID"
	keyDid         = "DID"
	keyComparePath = "COMPAREPATH"
)

func osEnvSet() []string {
	res := make([]string, 6, 6)
	res[0] = os.Getenv(keyHost)
	res[1] = os.Getenv(keyUsername)
	res[2] = os.Getenv(keyPassword)
	res[3] = os.Getenv(keySid)
	res[4] = os.Getenv(keyDid)
	res[5] = os.Getenv(keyComparePath)

	return res
}

func setArg(e string, a string, i int) (string, int) {
	if e != "" {
		return e, i
	}
	return a, i + 1
}

func main() {
	var host string
	flag.StringVar(&host, "h", "", usageHost)
	flag.StringVar(&host, "host", "", usageHost)
	var username string
	flag.StringVar(&username, "u", "", usageUsername)
	flag.StringVar(&username, "username", "", usageUsername)
	var password string
	flag.StringVar(&password, "p", "", usagePassword)
	flag.StringVar(&password, "password", "", usagePassword)
	var sid string
	flag.StringVar(&sid, "f", "", usageSid)
	flag.StringVar(&sid, "sourceid", "", usageSid)
	var did string
	flag.StringVar(&did, "t", "", usageDid)
	flag.StringVar(&did, "destid", "", usageDid)
	var path string
	flag.StringVar(&path, "c", "", usagePath)
	flag.StringVar(&path, "comparepath", "", usagePath)
	flag.Parse()

	/*
		fmt.Printf("param gets service ip %s, sid %d, did %d, compare %s, output %s, style %s, recursive %t\n",
			ip, sid, did, comparePath, outputPath, formatStyle, r)
	*/

	envlist := osEnvSet()
	i := 0
	if host == "" {
		host, i = setArg(envlist[0], flag.Arg(i), i)
	}
	if username == "" {
		username, i = setArg(envlist[1], flag.Arg(i), i)
	}
	if password == "" {
		password, i = setArg(envlist[2], flag.Arg(i), i)
	}
	if sid == "" {
		sid, i = setArg(envlist[3], flag.Arg(i), i)
	}
	if did == "" {
		did, i = setArg(envlist[4], flag.Arg(i), i)
	}
	if path == "" {
		path, i = setArg(envlist[5], flag.Arg(i), i)
	}
	r := regexp.MustCompile(`\d`)
	if host == "" ||
		r.MatchString(sid) == false ||
		r.MatchString(did) == false {
		log.Fatal("invalid argument")
	}

	if username == "" && password == "" {
		username = defaultUser
		password = defaultPass
	}

	if path == "" {
		path = defaultPath
	}

	c, err := xclient.NewClient(host, username, password)
	if err != nil {
		log.Fatal("new client fail")
	}

	diff, err := xsnap.DiffSnapshots(c, sid, did, path)
	if err != nil {
		log.Fatal("diff fail ", err)
	}
	fmt.Printf("%v\n", diff)
}
