package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type pair struct {
	hash, path string
}
type fileList []string
type results map[string]fileList

func searchTree(dir string) results {
	workers := 2 * runtime.GOMAXPROCS(0)
	wg := new(sync.WaitGroup)
	limits := make(chan bool, workers)
	pairs := make(chan pair, workers)
	result := make(chan results)
	go collectHashes(pairs, result)
	wg.Add(1)
	err := walkDir(dir, pairs, wg, limits)
	if err != nil {
		log.Fatal(err)
	}
	wg.Wait()
	// by closing pairs we signal that all the hashes
	// have been collected; we have to do it here AFTER all the workers are done
	close(pairs)
	hashes := <-result
	return hashes
}

func walkDir(dir string, pairs chan<- pair,
	wg *sync.WaitGroup, limits chan bool,
) error {
	defer wg.Done()

	visit := func(p string, fi os.FileInfo, err error) error {
		// ignore the error passed in
		// ignore dir itself to avoid an infinite loop!
		if fi.Mode().IsDir() && p != dir {
			wg.Add(1)
			go walkDir(p, pairs, wg, limits)
			return filepath.SkipDir
		}
		if fi.Mode().IsRegular() && fi.Size() > 0 {
			wg.Add(1)
			go processFiles(p, pairs, wg, limits)
		}
		return nil
	}
	limits <- true
	defer func() {
		<-limits
	}()
	return filepath.Walk(dir, visit)
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
func processFiles(path string, pairs chan<- pair,
	wg *sync.WaitGroup, limits chan bool,
) {
	defer wg.Done()
	limits <- true
	defer func() {
		<-limits
	}()
	pairs <- hashFile(path)
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
