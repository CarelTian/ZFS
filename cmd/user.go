package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func readInput(r io.Reader, w io.Writer) {
	reader := bufio.NewReader(r)
	for {
		fmt.Fprint(w, "请输入内容: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			// 如果遇到 EOF 则退出循环
			if err == io.EOF {
				os.Exit(0)
			}
		}
		input = strings.TrimSpace(input)
		if input == "exit" {
			fmt.Fprintln(w, "bye")
			break
		}
		fmt.Fprintf(w, "您输入了: %s\n", input)
	}
}

func readInputV2(r io.Reader, op chan string) {
	reader := bufio.NewReader(r)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				os.Exit(0)
			} else {
				log.Fatal(err)
			}
		}
		input = strings.TrimSpace(input)
		if input == "exit" {
			fmt.Println("bye")
			os.Exit(0)
		}
		op <- input
	}
}
