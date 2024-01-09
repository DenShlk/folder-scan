package main

import (
	"bufio"
	"fmt"
	"folder-scan/internal/differ"
	"os"
	"strings"
)

func RunInteractive(root *differ.FolderChangeInfo) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(printFolderChange(root))

		fmt.Println("Enter name of folder to go to. '.' to go up")
		input, _ := reader.ReadString('\n')
		input = strings.Trim(input, "\r\n")
		if input == "." {
			return
		}

		for _, sub := range root.Subs {
			if sub.Data.Cur.Name == input {
				RunInteractive(sub)
				continue
			}
		}
		fmt.Println("No such folder!")
	}
}
