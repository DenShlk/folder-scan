package main

import (
	"fmt"
	"folder-scan/internal/differ"
)

func RunInteractive(root *differ.FolderChangeInfo) {
	for {
		fmt.Println(printFolderChange(root))

		fmt.Println("Enter name of folder to go to. '.' to go up")
		var input string
		fmt.Scanln(&input)
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
