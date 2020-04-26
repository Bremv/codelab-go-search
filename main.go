package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	recursiveFlag = flag.Bool("r", false, "recursive search: for directories")
	numberingFlag = flag.Bool("n", false, "show line numbers")
)

type ScanResult struct {
	file       string
	lineNumber int
	line       string
}

// []string
func scanFile(fpath, pattern string) ([]ScanResult, []int, error) {
	f, err := os.Open(fpath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	result := make([]string, 0)
	res := make([]ScanResult, 0)
	lines := make([]int, 0)
	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		if strings.Contains(line, pattern) {
			res = append(res, ScanResult{file: fpath, lineNumber: lineNum, line: line})
			result = append(result, line)
			lines = append(lines, lineNum)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	return res, lines, nil
}

func exit(format string, val ...interface{}) {
	if len(val) == 0 {
		fmt.Println(format)
	} else {
		fmt.Printf(format, val)
		fmt.Println()
	}
	os.Exit(1)
}

func processFile(fpath string, pattern string) []ScanResult {
	result, _, err := scanFile(fpath, pattern)
	if err != nil {
		exit("Error scanning %s: %s", fpath, err.Error())
	}
	return result
}

func processDirectory(dir string, pattern string) (chan ScanResult, chan error) {
	res := make(chan ScanResult)
	errc := make(chan error, 1)
	go func() {
		var wg sync.WaitGroup
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Println(err)
				return err
			}
			if info.IsDir() {
				return nil
			}
			wg.Add(1)
			go func() {
				data := processFile(dir+info.Name(), pattern)
				for _, el := range data {
					res <- el
				}
				wg.Done()
			}()
			return nil
		})
		errc <- err
		wg.Wait()
		close(res)
	}()
	return res, errc
}
func printFile(data []ScanResult, numbering bool) {
	for _, d := range data {
		if numbering {
			fmt.Printf("%s:%d : %s\n", d.file, d.lineNumber, d.line)
		} else {
			fmt.Printf("%s: %s\n", d.file, d.line)
		}
	}
}

func printDir(data chan ScanResult, numbering bool) {
	for el := range data {
		if numbering {
			fmt.Printf("%s:%d : %s\n", el.file, el.lineNumber, el.line)
		} else {
			fmt.Printf("%s: %s\n", el.file, el.line)
		}
	}

}

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		exit("usage: go-search <path> <pattern> to search")
	}

	path := flag.Arg(0)
	pattern := flag.Arg(1)

	info, err := os.Stat(path)
	if err != nil {
		panic(err)
	}

	recursive := *recursiveFlag
	if info.IsDir() && !recursive {
		exit("%s: is a directory", info.Name())
	}

	if info.IsDir() && recursive {
		data, _ := processDirectory(path, pattern)
		printDir(data, *numberingFlag)
	} else {
		data := processFile(path, pattern)
		printFile(data, *numberingFlag)
	}
}
