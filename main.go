package main

import (
	"log"

	"cornchip.com/libwara/v2"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	nup := libwara.InitNup()
	err := nup.AddTopic("test")
	checkError(err)

	_, err = nup.AddPost("test")
	checkError(err)

	nup.Render("1stNUP.xml")
}
