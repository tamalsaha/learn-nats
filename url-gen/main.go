package main

import (
	"fmt"
	"net/url"
	"path"
)

func main() {
	u, err := url.Parse("https://appscode.ninja")
	if err != nil {
		panic(err)
	}

	u1 := *u
	u1.Path = path.Join(u1.Path, "api/v1/register")
	fmt.Println(u1.String())

	u.Path = path.Join(u.Path, "api/v1/join")
	fmt.Println(u.String())
}
