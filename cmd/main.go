package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"folder-scan/internal"
	"folder-scan/internal/differ"
	"folder-scan/internal/scanner"
	"log"
	"os"
	"runtime"
	"time"
)

var mode = flag.String("m", "scan", "Goal 'scan' to generate new report, 'diff' to compare two reports")
var rootDir = flag.String("r", "C:/", "Root folder to start scanning")
var workers = flag.Int("w", runtime.NumCPU()*8, "amount of workers, default=runtime.NumCPU()*8")
var saveTo = flag.String("s", "./report.json", "path to output json")

var oldReportPath = flag.String("o", "./report2.json", "Path to report to track changes from")
var newReportPath = flag.String("n", "./report.json", "Path to report to track changes to")

func main() {
	flag.Parse()

	if *mode == "scan" {
		scanMode()
	}
	if *mode == "diff" {
		diffMode()
	}
}

func scanMode() {
	cfg := scanner.Config{
		RootDir: *rootDir,
		Workers: *workers,
		SaveTo:  *saveTo,
	}

	start := time.Now()
	info := scanner.NewScanner(cfg).Start(context.Background())
	elapsed := time.Since(start)

	fmt.Println("Scanning time:", elapsed.Milliseconds(), "ms")

	info.CalcSize()

	file, err := json.Marshal(info)
	if err != nil {
		panic(err)
	}
	err = os.WriteFile(*saveTo, file, 0644)
	if err != nil {
		panic(err)
	}
	fmt.Println("Result saved.")
}

func diffMode() {
	oldRes := readResult(*oldReportPath)
	newRes := readResult(*newReportPath)

	change, err := differ.Diff(oldRes, newRes)
	if err != nil {
		log.Fatalln("Failed to find difference between reports", err)
	}

	RunInteractive(change)

	fmt.Println("Quited from interactive mode")
}

func readResult(path string) *internal.FolderInfo {
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	info := internal.FolderInfo{}
	err = json.Unmarshal(data, &info)
	if err != nil {
		panic(err)
	}
	return &info
}

/*

1:
Scanning time: 55714 ms
Scanning time: 35653 ms
Scanning time: 14232 ms
Scanning time: 29872 ms
Scanning time: 24069 ms
Scanning time: 19862 ms

64:
Scanning time: 5961 ms
Scanning time: 5851 ms
Scanning time: 5695 ms




full

Scanning time: 20000 ms
Scanning time: 14228 ms

*/
