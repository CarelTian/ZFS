package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

//
//import (
//	"ZFS/cmd"
//)
//
//func main() {
//	cmd.Start()
//}

func isInStorage(root, target string) (bool, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return false, err
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return false, err
	}
	relative, err := filepath.Rel(absRoot, absTarget)
	if err != nil {
		return false, err
	}
	// 如果relative以".."开头，则说明target不在root下
	return !strings.HasPrefix(relative, ".."), nil
}

func main() {
	ret, err := isInStorage("storage", "storage/2")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ret)

}
