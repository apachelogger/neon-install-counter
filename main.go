package main

import (
	"fmt"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"apachelogparser"
	"neon_install_counter"

	"github.com/gin-gonic/gin"
)

func process(c *gin.Context) {
	maxAge := 24.0 * 7 // hours in days
	maxedOut := false  // whether or not we found the oldest
	machineIDs := make(map[string]int)

	files, err := filepath.Glob("/var/log/apache2/releases.kde.neon.kde.org.log*")
	if err != nil {
		c.JSON(500, gin.H{})
		return
	}
	sort.Sort(neon_install_counter.BySuffix(files))
	fmt.Printf("globed: %s\n", files)
	for _, file := range files {
		if maxedOut {
			break
		}
		var lines []apachelogparser.Line
		if strings.HasSuffix(file, ".gz") {
			lines, err = apachelogparser.LoadCompressed(file)
		} else {
			lines, err = apachelogparser.Load(file)
		}
		if err != nil {
			c.JSON(500, gin.H{})
			return
		}
		for _, line := range lines {
			if time.Since(line.Time).Hours() > maxAge {
				fmt.Printf("maxed out in %s\n", file)
				maxedOut = true
				break
			}
			if !strings.HasPrefix(line.Url, "/meta-release-lts") {
				continue
			}
			_, machineID := path.Split(line.Url)
			machineIDs[machineID] = 0
		}
	}
	fmt.Printf("count: %d\n", len(machineIDs))
	c.JSON(200, gin.H{
		"install_count": len(machineIDs),
	})
}

func main() {
	fmt.Printf("Hello, world.\n")
	router := gin.Default()
	router.GET("/", process)
	router.Run(":8585")
}
