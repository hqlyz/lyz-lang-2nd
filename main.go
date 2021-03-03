package main

import (
	"bufio"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
)

var totalLines = 0

func main() {
	// u, err := user.Current()
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// fmt.Printf("Hello %s! This is the LYZ programming language!\n", u.Username)
	// fmt.Printf("Feel free to type in commands\n")
	// repl.Start(os.Stdin, os.Stdout)

	fss := []fs{}
	walkDir(".", &fss)
	
	sort.Slice(fss, func(i, j int) bool {
		return fss[j].lines < fss[i].lines
	})

	for _, v := range fss {
		log.Printf("%s: %d\n", v.path, v.lines)
	}
	log.Printf("the total lines in source code: %d\n", totalLines)
}

func walkDir(dir string, fss *[]fs) {
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println(err)
		return
	}

	for _, fi := range fis {
		path := filepath.Join(dir, fi.Name())
		if path == ".git" {
			continue
		}
		if fi.IsDir() {
			walkDir(path, fss)
		}
		ext := filepath.Ext(path)
		if ext != ".go" {
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			log.Println(err)
			return
		}
		defer f.Close()
		count := 0
		scan := bufio.NewScanner(f)
		for scan.Scan() {
			count++
			totalLines++
		}
		// log.Printf("%s: %d\n", path, count)
		*fss = append(*fss, fs{path: path, lines: count})
	}	
}

type fs struct {
	path string
	lines int
}