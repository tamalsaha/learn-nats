package main

import "fmt"

type User struct {
	Name  string
	Email string
}

func main() {
	u := User{
		Name:  "Tamal Saha",
		Email: "tamal@appscode.com",
	}
	fmt.Println(u)

	// make to verified
	// get token
	// helm command
	// render and get kubectl link
}
