# golang_concurrent_file_processing

This repository use goroutine and channels to fulfill CSP(Concurrent Sequencial Process)

## Problme

File Walk Example

find duplicate files based on their **content**

Use a secure hash, because the names /dates may differ
## origin process
```golang
package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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
}

```
## a concurrent approach(like map reduce)

use a fixed pool of goroutines and a collector and channels
![img](https://i.imgur.com/VJNNY2t.png)

## another concurrent approach

add a goroutine for each directory in the tree

## conccurent approach next

use a goroutine for every directory and file hash

Notice: without some control, we'll run out of threads

GOMAXPROCS doesn't limit threads blocked on syscalls(all our disk I/O)

We'll limit the number of **active** goroutines instead(the ones making syscalls)

A gorountines can't proceed without sending on the channel

A channel with buffer size N can accept N sends without blocking

The buffer provides a fixed upper bound

One goroutine can start for each one that finishes

![](https://i.imgur.com/hYBic0z.png)

## Amdal's law

speedup is limited by the part(not) parallelized

S = 1/( 1 - p + (p/s))

S = 6.25 on s = 8 processors, or about p = 96 % parallel
