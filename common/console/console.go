package console

import "log"

func Info(fmt string, args ...interface{}) {
	log.Printf("\033[31m[Info]\033[0m "+fmt, args...)
}

func Warn(fmt string, args ...interface{}) {
	log.Printf("\u001B[31m[Warn]\033[0m "+fmt, args...)
}

func Error(fmt string, args ...interface{}) {
	log.Printf("\u001B[31m[Error]\033[0m "+fmt, args...)
}

func Debug(fmt string, args ...interface{}) {
	log.Printf("\u001B[31m[Debug]\033[0m "+fmt, args...)
}
