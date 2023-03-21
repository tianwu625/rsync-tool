package snapshot

import (
	"encoding/json"
	"net/http"
	"net/url"
	"fmt"
	"strings"
	"path/filepath"
	"errors"

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
	Next int `json:"next,omitempty"`
}

const (
	defaultCount = 1000
)

func diffSnapshotsDir(c *xclient.Client, sid, did int, dir string) ([]EntryDir, error) {
	res := make([]EntryDir, 0, 0)
	urlStr := fmt.Sprintf("https://%s/api/v1/snapshots/%d/diff/%d/%s",c.Ip, sid, did, strings.TrimPrefix(dir, "/"))
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
	Length int `json:"length"`
	Type string `json:"type"`
	Offset int `json:"offset"`
}

type ResEntryFile struct {
	Content []EntryFile `json:"content"`
	Next int `json:"next,omitempty"`
}

var ErrInvalidFile = errors.New("file path should be '/'")

func diffSnapshotsFile(c *xclient.Client, sid, did int, file string) ([]EntryFile, error) {
	res := make([]EntryFile, 0, 0)
	if file == "" || file == "/" {
		return res, ErrInvalidFile
	}
	urlStr := fmt.Sprintf("https://%s/api/v1/snapshots/%d/diff/%d/%s",c.Ip, sid, did, strings.TrimPrefix(file, "/"))
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
	Name string
	Type string
}

func DiffSnapshots(c *xclient.Client, sid, did int, path string, recursive bool) ([]EntryDiff, error) {
	res := make([]EntryDiff, 0, 0)
	s, err := xfs.Stat(c, path)
	if err != nil {
		return res, err
	}
	if s.IsDir {
		xres, err := diffSnapshotsDir(c, sid, did, path)
		if err != nil {
			return res, err
		}
		for _, entry := range xres {
			diffEntry := EntryDiff {
				Name:entry.Name,
				Type:entry.Type,
			}
			res = append(res, diffEntry)
		}
		if recursive {
			for _, e := range xres {
				childres, err := DiffSnapshots(c, sid, did, filepath.Join(path, e.Name), recursive)
				if err != nil {
					return res, err
				}
				res = append(res, childres...)
			}
		}
	} else {
		xres, err := diffSnapshotsFile(c, sid, did, path)
		if err != nil {
			return res, err
		}
		if len(xres) != 0 {
			diffEntry := EntryDiff {
				Name:path,
				Type:"MODIFY",
			}
			res = append(res, diffEntry)
		}
	}

	return res, nil
}
