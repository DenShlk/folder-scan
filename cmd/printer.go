package main

import (
	"bytes"
	"fmt"
	"folder-scan/internal"
	"folder-scan/internal/differ"
	"folder-scan/util"
	"github.com/fatih/color"
	"io"
	"math"
	"text/tabwriter"
)

func printFolderTree(info *internal.FolderInfo, depth int) string {
	var buf bytes.Buffer
	table := tabwriter.NewWriter(&buf, 0, 0, 4, ' ', 0)

	_printFolderTree(table, info, "", depth)
	err := table.Flush()
	if err != nil {
		panic(err)
	}

	return buf.String()
}

func _printFolderTree(table *tabwriter.Writer, info *internal.FolderInfo, offset string, depth int) {
	folderInfo2Row(&offset, info, table)
	_, err := fmt.Fprint(table)
	if err != nil {
		panic(err)
	}

	offset += "    "
	if depth > 0 || depth <= -1 {
		for _, folder := range info.Subs {
			_printFolderTree(table, folder, offset, depth-1)
		}
	}

	if depth > 0 || depth <= -1 {
		for _, file := range info.Files {
			fileInfo2Row(&offset, file, table)
		}
	}
}

func fileInfo2Row(offset *string, file *internal.FileInfo, table *tabwriter.Writer) {
	_, err := fmt.Fprintf(table,
		"%s%.1f%%    =%dMb\t%s%s\n",
		*offset,
		float32(file.Size)/float32(file.Parent.Data.Size)*100,
		file.Size/1_000_000,
		*offset,
		file.Name,
	)
	if err != nil {
		panic(err)
	}
}

func folderInfo2Row(offset *string, folder *internal.FolderInfo, table *tabwriter.Writer) {
	if folder.Data.Parent == nil {
		_, err := fmt.Fprintf(table,
			"%s100%%    =%dMb\t%s%s\\\n",
			*offset,
			folder.Data.Size/1_000_000,
			*offset,
			folder.Data.Name,
		)
		if err != nil {
			panic(err)
		}
		return
	}
	_, err := fmt.Fprintf(table,
		"%s%.1f%%    =%dMb\t%s%s\\\n",
		*offset,
		float32(folder.Data.Size)/float32(folder.Data.Parent.Data.Size)*100,
		folder.Data.Size/1_000_000,
		*offset,
		folder.Data.Name,
	)
	if err != nil {
		panic(err)
	}
}

func dataChange2color(data *differ.DataChangeInfo) func(w io.Writer, format string, a ...interface{}) {
	if data.Cur.Size > data.Old.Size {
		return color.New(color.FgGreen).FprintfFunc()
	}
	if data.Cur.Size < data.Old.Size {
		return color.New(color.FgRed).FprintfFunc()
	}
	return color.New(color.FgWhite).FprintfFunc()
}

func dataChange2PercentSizeChange(data *differ.DataChangeInfo) float64 {
	if data.ChangeMode == differ.ChangeModeNew || data.ChangeMode == differ.ChangeModeDeleted {
		return 100
	}
	if util.MinInt64(data.Old.Size, data.Cur.Size) == 0 {
		if util.MaxInt64(data.Old.Size, data.Cur.Size) == 0 {
			return 0
		}
		return math.Inf(1)
	}
	return math.Abs(1-float64(data.Old.Size)/float64(data.Cur.Size)) * 100
}

func dataChange2Row(data, parentData *differ.DataChangeInfo, isFolder bool) string {
	sign := "="
	if data.Cur.Size > data.Old.Size {
		sign = "+"
	}
	if data.Cur.Size < data.Old.Size {
		sign = "-"
	}
	folderMark := ""
	if isFolder {
		folderMark = "\\"
	}
	//C:\tmp\	changed		+49.6%		+555Kb		20%		> 25%    12mb > 13mb
	return fmt.Sprintf("%s%s\t%s\t%s%.1f%%\t%s%s\t%.1f%%\t->\t%.1f%%\t%s\t->\t%s",
		data.Cur.Name,
		folderMark,
		data.ChangeMode,
		sign,
		dataChange2PercentSizeChange(data),
		sign,
		autoDataSize(util.AbsInt64(data.Cur.Size-data.Old.Size)),
		percentFromParent(data.Old, parentData.Old),
		percentFromParent(data.Cur, parentData.Cur),
		autoDataSize(data.Old.Size),
		autoDataSize(data.Cur.Size),
	)
}

func percentFromParent(child, parent *internal.FSData) float64 {
	if parent == nil {
		return 100
	}
	return float64(child.Size) / float64(parent.Size) * 100
}

func printFolderChange(info *differ.FolderChangeInfo) string {
	var buf bytes.Buffer
	table := tabwriter.NewWriter(&buf, 2, 4, 2, ' ', 0)

	color.New(color.FgWhite).Fprintln(table, "Analysis for path: ", info.Data.Cur.Path)
	color.New(color.FgWhite).Fprintln(table, "")
	color.New(color.FgWhite).Fprintln(table, "Name\tChange type\tPercent diff\tAbsolute diff\tOld fraction\t\tNew fraction\tOld size\t\tNew size")
	dataChange2color(info.Data)(table, "%s\n", dataChange2Row(info.Data, &differ.DataChangeInfo{}, true))

	for _, sub := range info.Subs {
		dataChange2color(sub.Data)(table, "|----%s\n", dataChange2Row(sub.Data, info.Data, true))
	}

	for _, file := range info.Files {
		data := (*differ.DataChangeInfo)(file)
		dataChange2color(data)(table, "|----%s\n", dataChange2Row(data, info.Data, false))
	}

	table.Flush()
	return buf.String()
}

func autoDataSize(b int64) string {
	if b < 10000 {
		return fmt.Sprintf("%db", b)
	}
	kb := b / 1000
	if kb < 10000 {
		return fmt.Sprintf("%dKb", kb)
	}
	mb := kb / 1000
	if mb < 10000 {
		return fmt.Sprintf("%dMb", mb)
	}
	gb := mb / 1000
	return fmt.Sprintf("%dGb", gb)
}
