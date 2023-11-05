package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

type pair struct {
	hash, path string
}
type fileList []string
type results map[string]fileList

func searchTree(dir string) results {
	workers := 2 * runtime.GOMAXPROCS(0)
	paths := make(chan string)
	pairs := make(chan pair)
	done := make(chan bool)
	result := make(chan results)
	for i := 0; i < workers; i++ {
		go processFiles(paths, pairs, done)
	}
	go collectHashes(pairs, result)

	err := filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		//ignore the error parm for now
		if fi.Mode().IsRegular() && fi.Size() > 0 {
			paths <- p
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
	close(paths)
	// wait for all the workers to be done
	for i := 0; i < workers; i++ {
		<-done
	}
	// by closing pairs we signal that all the hashes
	// have been collected; we have to do it here AFTER all the workers are done
	close(pairs)
	hashes := <-result
	return hashes
}

func hashFile(path string) pair {
	file, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	hash := md5.New()

	if _, err := io.Copy(hash, file); err != nil {
		log.Fatal(err)
	}
	return pair{fmt.Sprintf("%x", hash.Sum(nil)), path}
}
func processFiles(paths <-chan string, pairs chan<- pair, done chan<- bool) {
	for path := range paths {
		pairs <- hashFile(path)
	}
	done <- true
}
func collectHashes(pairs <-chan pair, result chan<- results) {
	hashes := make(results)

	for p := range pairs {
		hashes[p.hash] = append(hashes[p.hash], p.path)
	}

	result <- hashes
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing parameter, provide dir name!")
	}
	start := time.Now()
	if hashes := searchTree(os.Args[1]); hashes != nil {
		for hash, files := range hashes {
			if len(files) > 0 {
				// we will just 7 chars like git
				fmt.Println(hash[len(hash)-7:], len(files))
				for _, file := range files {
					fmt.Println("   ", file)
				}
			}
		}
	}
	fmt.Printf("time consume: %s\n", time.Since(start))
}
