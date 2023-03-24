package fs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	xclient "github.com/openfs/rsync-tools/client"
)

type State struct {
	FileType string
	Name     string
}

type resStat struct {
	OwnerGroup string `json:"owner-group"`
	Access     string `json:"access"`
	Blocks     int    `json:"blocks"`
	Atime      string `json:"atime"`
	Mtime      string `json:"mtime"`
	Ctime      string `json:"ctime"`
	Gfid       string `json:"gfid"`
	Nlink      int    `json:"nlink"`
	Ino        int    `json:"ino"`
	Blksize    int    `json:"blksize"`
	Sid        int    `json:"sid"`
	OwnerName  string `json:"owner-name"`
	OwnerGid   int    `json:"owner-gid"`
	Type       string `json:"type"`
	Size       int    `json:"size"`
}

func Stat(c *xclient.Client, path, sid string) (State, error) {
	s := State{
		Name: path,
	}
	urlStr := fmt.Sprintf("https://%s/api/v1/namespace/%s", c.Ip, strings.TrimPrefix(path, "/"))
	u, err := url.Parse(urlStr)
	if err != nil {
		return s, err
	}
	q := u.Query()
	q.Set("metadata", "true")
	q.Set("sid", sid)
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
	s.FileType = xres.Type
	//fmt.Printf("%v\n", s)

	return s, nil
}

func Readdir(c *xclient.Client, path, sid string, offset, count int) ([]State, error) {
	s := make([]State, 0, 0)
	urlStr := fmt.Sprintf("https://%s/api/v1/namespace/%s", c.Ip, strings.TrimPrefix(path, "/"))
	u, err := url.Parse(urlStr)
	if err != nil {
		return s, err
	}
	q := u.Query()
	q.Set("sid", sid)
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
		e := State{
			Name:     filepath.Join(path, k),
			FileType: v.Type,
		}
		s = append(s, e)
	}
	//fmt.Printf("res %v\n", s)

	return s, nil
}

const (
	defaultCount = 1000
)

func ReaddirRecursive(c *xclient.Client, path, sid string) ([]State, error) {
	s := make([]State, 0, 0)

	offset := 0
	count := defaultCount
	for count == defaultCount {
		res, err := Readdir(c, path, sid, offset, count)
		if err != nil {
			return s, err
		}
		s = append(s, res...)
		for _, e := range res {
			if e.FileType == "directory" {
				childres, err := ReaddirRecursive(c, filepath.Join(path, e.Name), sid)
				if err != nil {
					return s, err
				}
				s = append(s, childres...)
			}
		}
		count = len(res)
		offset += count
	}

	return s, nil
}
