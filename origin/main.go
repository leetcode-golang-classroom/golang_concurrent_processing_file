package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type pair struct {
	hash, path string
}
type fileList []string
type results map[string]fileList

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

func searchTree(dir string) (results, error) {
	hashes := make(results)

	err := filepath.Walk(dir, func(p string, fi os.FileInfo, err error) error {
		//ignore the error parm for now
		if fi.Mode().IsRegular() && fi.Size() > 0 {
			h := hashFile(p)
			hashes[h.hash] = append(hashes[h.hash], h.path)
		}
		return nil
	})
	return hashes, err
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing parameter, provide dir name!")
	}
	start := time.Now()
	if hashes, err := searchTree(os.Args[1]); err == nil {
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
