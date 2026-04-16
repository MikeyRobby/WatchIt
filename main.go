package main

import (
	"fmt"
)

func main() {
	fmt.Println(addItem())
}

func addItem() Item {
	watch := Item{Brand: "Seiko", Model: "5 Sport", CaseSize: "38mm"}
	return watch
}
