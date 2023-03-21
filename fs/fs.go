package fs

import (
	"fmt"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"path/filepath"

	xclient "github.com/openfs/rsync-tools/client"
)

type State struct {
	IsDir bool
	Name string
}

type resStat struct {
	OwnerGroup string `json:"owner-group"`
	Access string `json:"access"`
	Blocks int `json:"blocks"`
	Atime string `json:"atime"`
	Mtime string `json:"mtime"`
	Ctime string `json:"ctime"`
	Gfid string `json:"gfid"`
	Nlink int `json:"nlink"`
	Ino int `json:"ino"`
	Blksize int `json:"blksize"`
	Sid int `json:"sid"`
	OwnerName string `json:"owner-name"`
	OwnerGid int `json:"owner-gid"`
	Type string `json:"type"`
	Size int `json:"size"`
}

func Stat(c *xclient.Client, path string) (State, error) {
	s := State{
		IsDir: true,
		Name:path,
	}
	urlStr := fmt.Sprintf("https://%s/api/v1/namespace/%s", c.Ip, strings.TrimPrefix(path, "/"))
	u, err := url.Parse(urlStr)
	if err != nil {
		return s, err
	}
	q := u.Query()
	q.Set("metadata", "true")
	u.RawQuery = q.Encode()
	//fmt.Println(u.String())
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return s, err
	}
	r, err := c.Do(req)
	if err != nil {
		return s, err
	}
	defer r.Close()
	xres := resStat{}
	err = json.NewDecoder(r).Decode(&xres)
	if err != nil {
		return s, err
	}
	if xres.Type != "directory" {
		s.IsDir = false
	}
	//fmt.Printf("%v\n", s)

	return s, nil
}

func Readdir(c *xclient.Client, path string, offset, count int) ([]State, error) {
	s := make([]State, 0, 0)
	urlStr := fmt.Sprintf("https://%s/api/v1/namespace/%s", c.Ip, strings.TrimPrefix(path, "/"))
	u, err := url.Parse(urlStr)
	if err != nil {
		return s, err
	}
	q := u.Query()
	q.Set("count", fmt.Sprintf("%d", count))
	q.Set("offset", fmt.Sprintf("%d", offset))
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return s, err
	}

	r, err := c.Do(req)
	if err != nil {
		return s, err
	}
	defer r.Close()
	xres := make(map[string]resStat, 0)
	err = json.NewDecoder(r).Decode(&xres)
	if err != nil {
		fmt.Printf("fail to decode %v\n", err)
	}
	for k, v := range xres {
		e := State {
			Name: filepath.Join(path, k),
			IsDir: v.Type == "directory",
		}
		s = append(s, e)
	}
	//fmt.Printf("res %v\n", s)

	return s, nil
}
