package main

import (
	"fmt"
	"net/http"

	"github.com/matthman4/cmdrest"
	"github.com/matthman4/cmdrest/example/ls"
)

func main() {
	h, err := cmdrest.CreateNewHandler(ls.DefaultLS)
	if err != nil {
		panic(err)
	}
	http.Handle("/ls/", h)

	fmt.Println("Started /ls on 8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
