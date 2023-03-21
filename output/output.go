package output


import (
	"fmt"
	"os"
	"io"

	xsnap "github.com/openfs/rsync-tools/snapshot"
)

const (
	StyleStandard      = "standard"
	StyleRsyncAll      = "rsync-all"
	StyleRsyncIncrease = "rsync"
	StyleDiffer        = "differ"
)

var StyleList = []string{
	StyleStandard,
	StyleRsyncAll,
	StyleRsyncIncrease,
	StyleDiffer,
}

func InStyleList(style string) bool {
	res := false
	for _, i := range StyleList {
		if i == style {
			res = true
			break
		}
	}

	return res
}

func outStandard(o io.Writer, diffs []xsnap.EntryDiff) error {
	for _, e := range diffs {
		fmt.Fprintf(o, "%v %v\n", e.Name, e.Type)
	}

	return nil
}

func outRsyncAll(o io.Writer, diffs []xsnap.EntryDiff) error {
	for _, e := range diffs {
		fmt.Fprintf(o, "%v\n", e.Name)
	}

	return nil
}

func outRsyncIncrease(o io.Writer, diffs []xsnap.EntryDiff) error {
	for _, e := range diffs {
		if e.Type == "DELETE" {
			continue
		}
		fmt.Fprintf(o, "%v\n", e.Name)
	}

	return nil
}

func outDiffer(o io.Writer, diffs []xsnap.EntryDiff, extra ...string) error {

	return nil
}

func OutputFile(outputPath, formatStr string, diffs []xsnap.EntryDiff, extra ...string) (err error){
	ofd := os.Stdout
	if outputPath != "" {
		ofd, err = os.OpenFile(outputPath, os.O_RDWR|os.O_CREATE, 0777)
		if err != nil {
			return
		}
		defer ofd.Close()
	}
	switch formatStr {
	case StyleStandard:
		err = outStandard(ofd, diffs)
	case StyleRsyncAll:
		err = outRsyncAll(ofd, diffs)
	case StyleRsyncIncrease:
		err = outRsyncIncrease(ofd, diffs)
	case StyleDiffer:
		err = outDiffer(ofd, diffs, extra...)
	default:
		err = fmt.Errorf("not support %v style", formatStr)
	}

	return
}
