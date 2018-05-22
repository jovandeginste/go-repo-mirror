package main

import "log"

func logIt(v ...interface{}) {
	if *verbose {
		log.Println(v...)
	}
}
func logItf(format string, v ...interface{}) {
	if *verbose {
		log.Printf(format, v...)
	}
}
func logItFatal(v ...interface{}) {
	log.Fatal(v...)
}
