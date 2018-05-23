package main

import "log"

var loggers []*log.Logger

func logIt(level int, v ...interface{}) {
	if level < *verbose {
		for _, l := range loggers {
			l.Println(v...)
		}
	}
}
func logItf(level int, format string, v ...interface{}) {
	if level < *verbose {
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
