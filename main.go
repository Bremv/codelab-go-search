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
	wg            sync.WaitGroup
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

func processFile(fpath string, pattern string, numbering bool) {
	result, _, err := scanFile(fpath, pattern)
	if err != nil {
		exit("Error scanning %s: %s", fpath, err.Error())
	}
	if numbering {
		for _, res := range result {
			fmt.Printf("%s:%d : %s\n", res.file, res.lineNumber, res.line)
			//fmt.Printf("%s:%d : %s\n", fpath, lines[i], line)
		}
	} else {
		for _, res := range result {
			fmt.Printf("%s: %s\n", res.file, res.line)
		}
	}
	wg.Done()
}

func processDirectory(dir string, pattern string, numbering bool) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if !info.IsDir() {
			wg.Add(1)
			go processFile(dir+info.Name(), pattern, numbering)
		}

		wg.Wait()
		return nil
	})
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
		processDirectory(path, pattern, *numberingFlag)
	} else {
		processFile(path, pattern, *numberingFlag)
	}
}
