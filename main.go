package main

import (
	"fmt"
	log "libs/log"
)

func main() {
	fmt.Println("print stdout")
	log.Trace.Println("Trace msg")
	log.Info.Println("Info msg")
	log.Error.Println("Error msg")
	log.MyLogger("ssssss")
}
