package utils

import "log"

func CheckError(err error, str string) {
	if err != nil {
		str += "%v"
		log.Fatalf(str, err)
	}
}
