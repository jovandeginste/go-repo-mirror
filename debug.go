package main

import "log"

var loggers []*log.Logger

func logIt(v ...interface{}) {
	if *verbose {
		for _, l := range loggers {
			l.Println(v...)
		}
	}
}
func logItf(format string, v ...interface{}) {
	if *verbose {
		for _, l := range loggers {
			l.Printf(format, v...)
		}
	}
}
func logItFatal(v ...interface{}) {
	for _, l := range loggers {
		l.Fatal(v...)
	}
}
