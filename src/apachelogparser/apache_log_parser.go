package apachelogparser

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"time"
)

type Line struct {
	RemoteHost string
	Time       time.Time
	Request    string
	Status     int
	Bytes      int
	Referer    string
	UserAgent  string
	Url        string
}

func (li *Line) String() string {
	return fmt.Sprintf(
		"%s\t%s\t%s\t%d\t%d\t%s\t%s\t%s",
		li.RemoteHost,
		li.Time,
		li.Request,
		li.Status,
		li.Bytes,
		li.Referer,
		li.UserAgent,
		li.Url,
	)
}

func Load(path string) ([]Line, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return read(file)
}

func LoadCompressed(filename string) ([]Line, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return read(reader)
}

func read(reader io.Reader) ([]Line, error) {
	var lines []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if scanner.Err() != nil {
		return nil, scanner.Err()
	}
	return parse(lines)
}

func parse(lines []string) ([]Line, error) {
	var items []Line
	for _, line := range lines {
		var buffer bytes.Buffer
		buffer.WriteString(`^(\S+)\s`)                  // 1) IP
		buffer.WriteString(`\S+\s+`)                    // remote logname
		buffer.WriteString(`(?:\S+\s+)+`)               // remote user
		buffer.WriteString(`\[([^]]+)\]\s`)             // 2) date
		buffer.WriteString(`"(\S*)\s?`)                 // 3) method
		buffer.WriteString(`(?:((?:[^"]*(?:\\")?)*)\s`) // 4) URL
		buffer.WriteString(`([^"]*)"\s|`)               // 5) protocol
		buffer.WriteString(`((?:[^"]*(?:\\")?)*)"\s)`)  // 6) or, possibly URL with no protocol
		buffer.WriteString(`(\S+)\s`)                   // 7) status code
		buffer.WriteString(`(\S+)\s`)                   // 8) bytes
		buffer.WriteString(`"((?:[^"]*(?:\\")?)*)"\s`)  // 9) referrer
		buffer.WriteString(`"(.*)"$`)                   // 10) user agent

		re1, err := regexp.Compile(buffer.String())
		if err != nil {
			return items, err
		}
		result := re1.FindStringSubmatch(line)

		lineItem := new(Line)
		lineItem.RemoteHost = result[1]
		// [05/Oct/2014:04:06:21 -0500]
		value := result[2]
		layout := "02/Jan/2006:15:04:05 -0700"
		t, _ := time.Parse(layout, value)
		lineItem.Time = t
		lineItem.Request = result[3] + " " + result[4] + " " + result[5]
		status, err := strconv.Atoi(result[7])
		if err != nil {
			status = 0
		}
		bytes, err := strconv.Atoi(result[8])
		if err != nil {
			bytes = 0
		}
		lineItem.Status = status
		lineItem.Bytes = bytes
		lineItem.Referer = result[9]
		lineItem.UserAgent = result[10]
		url := result[4]
		altUrl := result[6]
		if url == "" && altUrl != "" {
			url = altUrl
		}
		lineItem.Url = url
		items = append(items, *lineItem)
		//for k, v := range result {
		//	fmt.Printf("%d. %s\n", k, v)
		//}
	}
	return items, nil
}

// func Parse(file string) ([]Line, error) {
// 	lines, err := parse(file)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return lines, nil
// }
