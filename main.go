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
const (
	currentTagIdx = iota
	lastTagIdx
	timeMarkerIdx
	numberOfRecords
)

var helpMessage = "tag is missed"
var pathError = &fs.PathError{}

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Error: ", r)
		}
	}()

	newTag, comment := getParams()

	now := time.Now()
	log, err := getLogFile(now)
	if err != nil {
		panic(err)
	}
	defer log.Close()

	mBytes, err := os.ReadFile(timeMarkerFPath)
	if err != nil && !errors.As(err, &pathError) {
		panic(err)
	}

	var curTag, lastTag string
	var marker int
	if err == nil {
		curTag, lastTag, marker = getMarkers(mBytes)
	}

	if newTag == curTag {
		panic("New tag has to be different from the last one.")
	}

	mFile, err := os.OpenFile(timeMarkerFPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer mFile.Close()

	hours, mins := getTimeDifference(marker, now)
	if _, err = log.WriteString(fmt.Sprintf("%s,%s,%d,%d,%s\n", now.Format(time.RFC3339), lastTag, hours, mins, comment)); err != nil {
		panic(err)
	}

	curTag, lastTag = newTag, curTag
	if _, err = mFile.WriteString(fmt.Sprintf("%s\n%s\n%d", newTag, curTag, now.Unix())); err != nil {
		panic(err)
	}

	fmt.Println("OK")
}

func getTimeDifference(marker int, now time.Time) (int, int) { //bench it
	var hours, mins int
	if marker != 0 {
		timeMarker := time.Unix(int64(marker), 0)
		diff := now.Sub(timeMarker)
		hours, mins = int(diff.Hours()), int(diff.Minutes())%60
	}

	return hours, mins
}

func getMarkers(mBytes []byte) (string, string, int) {
	strs := strings.SplitN(string(mBytes), "\n", numberOfRecords)
	if len(strs) < 3 {
		panic("less than 2 lines in the '" + timeMarkerFName + "' file")
	}
	currentOperation := strs[currentTagIdx]
	lastOperation := strs[lastTagIdx]
	marker, err := strconv.Atoi(strs[timeMarkerIdx])
	if err != nil {
		panic(err)
	}

	return currentOperation, lastOperation, marker
}

// the caller of the function IS RESPONSIBLE for closing the returned file
func getLogFile(now time.Time) (*os.File, error) {
	err := os.MkdirAll("./logs", 0755)
	if err != nil {
		return nil, err
	}

	fname := fmt.Sprintf("%d-%d-%d.csv", now.Year(), now.Month(), now.Day())
	log, err := os.OpenFile("./logs/"+fname, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return log, nil
}

func getParams() (string, string) {
	args := os.Args
	if len(args) < 2 {
		fmt.Println(helpMessage)
		os.Exit(1)
	}

	tag := args[1]
	var comment string
	if len(args) > 2 {
		comment = args[2]
	}

	return tag, comment
}
