package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"
	"time"
)

const timeMarkerFName = "timemarker"
const timeMarkerFPath = "./" + timeMarkerFName

var helpMessage = "command name is missed,\n" +
	"options: start, stop"
var pathError = &fs.PathError{}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error: ", r)
		}
	}()

	op := getOperator()

	now := time.Now()
	log, err := getLogFile(&now)
	defer log.Close()

	mBytes, err := os.ReadFile(timeMarkerFPath)
	if err != nil && !errors.As(err, &pathError) {
		panic(err)
	}

	var lastOperation string
	var marker int
	if err == nil {
		lastOperation, marker = getMarkers(mBytes)
	}

	if op == lastOperation {
		panic("New operator has to be different from the last one.")
	}

	mFile, err := os.OpenFile(timeMarkerFPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer mFile.Close()

	hours, mins := getTimeDIfference(marker, &now)
	if _, err = log.WriteString(fmt.Sprintf("%s,%s,%d,%d\n", now.Format(time.RFC3339), op, hours, mins)); err != nil {
		panic(err)
	}

	if _, err = mFile.WriteString(fmt.Sprintf("%s\n%d\n", op, now.Unix())); err != nil {
		panic(err)
	}

	fmt.Println("OK")
}

func getTimeDIfference(marker int, now *time.Time) (int, int) {
	var hours, mins int
	if marker != 0 {
		timeMarker := time.Unix(int64(marker), 0)
		diff := now.Sub(timeMarker)
		hours, mins = int(diff.Hours()), int(diff.Minutes())%60
	}
	return hours, mins
}

func getMarkers(mBytes []byte) (string, int) {
	strs := strings.SplitN(string(mBytes), "\n", 3)
	if len(strs) < 2 {
		panic("less than 2 lines in the '" + timeMarkerFName + "' file")
	}
	lastOperation := strs[0]
	marker, err := strconv.Atoi(strs[1])
	if err != nil {
		panic(err)
	}
	return lastOperation, marker
}

// the caller of the function IS RESPONSIBLE for closing the returned file
func getLogFile(now *time.Time) (*os.File, error) {
	fname := fmt.Sprintf("%d-%d-%d.csv", now.Year(), now.Month(), now.Day())
	log, err := os.OpenFile("./logs/"+fname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}

	return log, err
}

func getOperator() string {
	args := os.Args
	if len(args) < 2 {
		fmt.Println(helpMessage)
		os.Exit(1)
	}

	switch args[1] {
	case "start":
	case "stop":
	default:
		fmt.Println(helpMessage)
		os.Exit(1)
	}

	return args[1]
}
