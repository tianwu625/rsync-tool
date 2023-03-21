package main

import (
	"flag"
//	"fmt"
	"log"
	"net"
	"strconv"
//	"time"

	xclient "github.com/openfs/rsync-tools/client"
	xsnap "github.com/openfs/rsync-tools/snapshot"
//	xfs "github.com/openfs/rsync-tools/fs"
	xout "github.com/openfs/rsync-tools/output"
)

const (
	//defaultIp = "127.0.0.1"
	usageIp        = "ip address of server"
	usageSid       = "source snapshot id"
	usageDid       = "destination snapshot id"
	defaultPath    = "/"
	usagePath      = "compare file path"
	usageOutput    = "output file path"
	usageStyle     = "format style of output"
	usageRecursive = "recursive directory"
)


func main() {
	var ip string
	flag.StringVar(&ip, "s", "", usageIp)
	flag.StringVar(&ip, "serviceip", "", usageIp)
	var sid int
	flag.IntVar(&sid, "f", 0, usageSid)
	flag.IntVar(&sid, "sourceid", 0, usageSid)
	var did int
	flag.IntVar(&did, "t", 0, usageDid)
	flag.IntVar(&did, "destid", 0, usageDid)
	var comparePath string
	flag.StringVar(&comparePath, "p", defaultPath, usagePath)
	flag.StringVar(&comparePath, "comparepath", defaultPath, usagePath)
	var outputPath string
	flag.StringVar(&outputPath, "o", "", usageOutput)
	flag.StringVar(&outputPath, "outputpath", "", usageOutput)
	var formatStyle string
	flag.StringVar(&formatStyle, "c", "standard", usageStyle)
	flag.StringVar(&formatStyle, "comparestyle", "standard", usageStyle)
	var r bool
	flag.BoolVar(&r, "r", true, usageRecursive)
	flag.BoolVar(&r, "recursive", true, usageRecursive)
	flag.Parse()
	i := 0
	if ip == "" {
		ip = flag.Arg(i)
		i++
	}
	var err error
	if sid == 0 && flag.Arg(i) != "" {
		sid, err = strconv.Atoi(flag.Arg(i))
		if err != nil {
			log.Fatal("set invalid source id")
		}
		i++
	}
	if did == 0 && flag.Arg(i) != "" {
		did, err = strconv.Atoi(flag.Arg(i))
		if err != nil {
			log.Fatal("set invalid source id")
		}
		i++
	}
	if flag.Arg(i) != "" {
		comparePath = flag.Arg(i)
		i++
	}
	if flag.Arg(i) != "" {
		outputPath = flag.Arg(i)
		i++
	}
	if flag.Arg(i) != "" {
		formatStyle = flag.Arg(i)
		i++
	}

	//fmt.Printf("args %v, ncount %v, count %v, ip %v\n", flag.Args(), flag.NArg(), flag.NFlag(), ip)
	if nil == net.ParseIP(ip) || nil == net.ParseIP(ip).To4() {
		log.Fatal("service ip only support IPv4")
	}
	if sid == 0 || did == 0 {
		log.Fatal("source id and destination id should be set id of snapshot")
	}
	if !xout.InStyleList(formatStyle) {
		log.Fatal("style only one of %v", xout.StyleList)
	}
	ip = net.ParseIP(ip).To4().String()

	/*
	fmt.Printf("param gets service ip %s, sid %d, did %d, compare %s, output %s, style %s, recursive %t\n",
		ip, sid, did, comparePath, outputPath, formatStyle, r)
	*/
	c, err := xclient.NewClient(ip)
	if err != nil {
		log.Fatal("new client fail")
	}

	diffs, err:= xsnap.DiffSnapshots(c, sid, did, comparePath, r)
	if err != nil {
		log.Fatal("diff fail ", err)
	}

	err = xout.OutputFile(outputPath, formatStyle, diffs)
	if err != nil {
		log.Fatal("output fail ", err)
	}
}
