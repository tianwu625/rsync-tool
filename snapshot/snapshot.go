package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	xclient "github.com/openfs/rsync-tools/client"
	xfs "github.com/openfs/rsync-tools/fs"
)

type Snapshot struct {
	Policy string `json:"policy"`
	State  string `json:"state"`
	Stime  int    `json:"stime"`
	Spid   int    `json:"spid"`
	Sid    int    `json:"sid"`
}

func ListSnapshots(c *xclient.Client) ([]Snapshot, error) {
	res := make([]Snapshot, 0, 1000)
	req, err := http.NewRequest("GET", "https://"+c.Ip+"/api/v1/snapshots", nil)
	if err != nil {
		return res, err
	}
	r, err := c.Do(req)
	if err != nil {
		return res, err
	}
	defer r.Close()
	if err := json.NewDecoder(r).Decode(&res); err != nil {
		return res, err
	}

	return res, nil
}

type EntryDir struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type ResEntryDir struct {
	Content []EntryDir `json:"content"`
	Next    int        `json:"next,omitempty"`
}

const (
	defaultCount = 1000
)

func diffSnapshotsDir(c *xclient.Client, sid, did, dir string) ([]EntryDir, error) {
	res := make([]EntryDir, 0, 0)
	urlStr := fmt.Sprintf("https://%s/api/v1/snapshots/%s/diff/%s/%s", c.Ip, sid, did, strings.TrimPrefix(dir, "/"))
	u, err := url.Parse(urlStr)
	if err != nil {
		return res, err
	}
	offset := 0
	xres := ResEntryDir{}
	for {
		q := u.Query()
		q.Set("type", "directory")
		q.Set("offset", fmt.Sprintf("%d", offset))
		q.Set("count", fmt.Sprintf("%d", defaultCount))
		u.RawQuery = q.Encode()
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return res, err
		}
		r, err := c.Do(req)
		if err != nil {
			return res, err
		}
		defer r.Close()
		if err := json.NewDecoder(r).Decode(&xres); err != nil {
			return res, err
		}
		res = append(res, xres.Content...)
		if xres.Next == 0 {
			break
		}
		offset = xres.Next
	}

	return res, nil
}

type EntryFile struct {
	Length int    `json:"length"`
	Type   string `json:"type"`
	Offset int    `json:"offset"`
}

type ResEntryFile struct {
	Content []EntryFile `json:"content"`
	Next    int         `json:"next,omitempty"`
}

var ErrInvalidFile = errors.New("file path should be '/'")

func diffSnapshotsFile(c *xclient.Client, sid, did, file string) ([]EntryFile, error) {
	res := make([]EntryFile, 0, 0)
	if file == "" || file == "/" {
		return res, ErrInvalidFile
	}
	urlStr := fmt.Sprintf("https://%s/api/v1/snapshots/%s/diff/%s/%s", c.Ip, sid, did, strings.TrimPrefix(file, "/"))
	u, err := url.Parse(urlStr)
	if err != nil {
		return res, err
	}
	offset := 0
	xres := ResEntryFile{}
	for {
		q := u.Query()
		q.Set("type", "file")
		q.Set("offset", fmt.Sprintf("%d", offset))
		u.RawQuery = q.Encode()
		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return res, err
		}
		r, err := c.Do(req)
		if err != nil {
			return res, err
		}
		defer r.Close()
		if err := json.NewDecoder(r).Decode(&xres); err != nil {
			return res, err
		}
		res = append(res, xres.Content...)
		if xres.Next == 0 {
			break
		}
		offset = xres.Next
	}

	return res, nil
}

type EntryDiff struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func deleteDiff(c *xclient.Client, path, sid string) ([]EntryDiff, error) {
	res := make([]EntryDiff, 0, 0)
	res = append(res, EntryDiff{
		Name: path,
		Type: "DELETE",
	})
	s, err := xfs.Stat(c, path, sid)
	if err != nil {
		return res, err
	}
	if s.FileType == "directory" {
		readdirRes, err := xfs.ReaddirRecursive(c, path, sid)
		if err != nil {
			return res, err
		}
		for _, re := range readdirRes {
			childpath := filepath.Join(path, re.Name)
			if re.FileType == "directory" {
				childpath += "/"
			}
			res = append(res, EntryDiff{
				Name: childpath,
				Type: "DELETE",
			})
		}
	}

	return res, nil
}

func createDiff(c *xclient.Client, path, did string) ([]EntryDiff, error) {
	res := make([]EntryDiff, 0, 0)
	res = append(res, EntryDiff{
		Name: path,
		Type: "CREATE",
	})
	s, err := xfs.Stat(c, path, did)
	if err != nil {
		return res, err
	}
	if s.FileType == "directory" {
		readdirRes, err := xfs.ReaddirRecursive(c, path, did)
		if err != nil {
			return res, err
		}
		for _, re := range readdirRes {
			childpath := filepath.Join(path, re.Name)
			if re.FileType == "directory" {
				childpath += "/"
			}
			res = append(res, EntryDiff{
				Name: childpath,
				Type: "CREATE",
			})
		}
	}
	return res, nil
}

func diffSnapshots(c *xclient.Client, sid, did, path string, checkType bool) ([]EntryDiff, error) {
	res := make([]EntryDiff, 0, 0)
	s, err := xfs.Stat(c, path, did)
	if err != nil {
		return res, err
	}
	if checkType {
		olds, err := xfs.Stat(c, path, sid)
		if err != nil {
			return res, err
		}
		if olds.FileType != s.FileType {
			return res, errors.New("diff file type can't be diff")
		}
	}

	if s.FileType == "directory" {
		xres, err := diffSnapshotsDir(c, sid, did, path)
		if err != nil {
			return res, err
		}
		for _, e := range xres {
			if e.Type == "DELETE" {
				dres, err := deleteDiff(c, filepath.Join(path, e.Name), sid)
				if err != nil {
					return res, err
				}
				res = append(res, dres...)
			} else if e.Type == "CREATE" {
				cres, err := createDiff(c, filepath.Join(path, e.Name), did)
				if err != nil {
					return res, err
				}
				res = append(res, cres...)
			} else {
				olds, err := xfs.Stat(c, filepath.Join(path, e.Name), sid)
				if err != nil {
					return res, err
				}
				news, err := xfs.Stat(c, filepath.Join(path, e.Name), did)
				if err != nil {
					return res, err
				}
				if olds.FileType == news.FileType {
					res = append(res, EntryDiff{
						Name: filepath.Join(path, e.Name),
						Type: e.Type,
					})
					if olds.FileType == "directory" {
						childres, err := diffSnapshots(c, sid, did, filepath.Join(path, e.Name), false)
						if err != nil {
							return res, err
						}
						res = append(res, childres...)
					}
				} else {
					dres, err := deleteDiff(c, filepath.Join(path, e.Name), sid)
					if err != nil {
						return res, err
					}
					res = append(res, dres...)
					cres, err := createDiff(c, filepath.Join(path, e.Name), did)
					if err != nil {
						return res, err
					}
					res = append(res, cres...)
				}
			}
		}
	} else {
		xres, err := diffSnapshotsFile(c, sid, did, path)
		if err != nil {
			return res, err
		}
		if len(xres) != 0 {
			diffEntry := EntryDiff{
				Name: path,
				Type: "MODIFY",
			}
			res = append(res, diffEntry)
		}
	}
	return res, nil
}

func DiffSnapshots(c *xclient.Client, sid, did, path string) (string, error) {
	res, err := diffSnapshots(c, sid, did, path, true)
	if err != nil {
		return "", err
	}
	/*
	for i, e := range res {
		fmt.Printf("No.%d name: %s, type %s\n", i, e.Name, e.Type)
	}
	*/
	jstr, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(jstr), nil
}
