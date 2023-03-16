package main

import (
	"embed"
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"
)

func serveFile(w http.ResponseWriter, r *http.Request) {
	// Which squirrel?
	n, err := strconv.Atoi(r.FormValue("n"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Expected integer parameter: n")
		return
	}
	if n < 0 || n >= len(pictures) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Unexpected value %d for n. Max allowed is %d", n, len(pictures)-1)
		return
	}

	// Default 100kB/s == a 10s pause before sending a 1MB resource
	var kbPerSecond = 100
	if s, err := strconv.Atoi(r.FormValue("speed")); err == nil {
		kbPerSecond = s
	}

	data := pictures[n]
	size := len(data)

	// Pause before writing.
	delay := ((1000 * (time.Duration(size) / 1024)) / time.Duration(kbPerSecond)) * time.Millisecond
	log.Printf("Pausing %v for picture %d having size %dkB\n", delay, n, size/1024)
	time.Sleep(delay)
	w.Header().Add("content-type", "image/jpeg")
	// TODO: streaming instead?
	w.Write(data)
}

func main() {
	sum := 0
	for _, pic := range pictures {
		sum += len(pic)
	}
	log.Printf("Ready to serve %d pictures, total size %dMB\n", len(pictures), sum/(1024*1024))

	http.HandleFunc("/squirrel", serveFile)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	log.Fatal(err)
}

var (
	//go:embed pictures/*.jpg
	picturesFolder embed.FS
	pictures       [][]byte
)

func init() {
	entries, err := picturesFolder.ReadDir("pictures")
	if err != nil {
		log.Fatal(err)
	}
	slices.SortFunc(entries, func(a, b fs.DirEntry) bool {
		return a.Name() < b.Name()
	})
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".jpg") {
			f, err := picturesFolder.Open("pictures/" + entry.Name())
			if err != nil {
				log.Fatal(err)
			}
			data, err := io.ReadAll(f)
			if err != nil {
				log.Fatal(err)
			}
			pictures = append(pictures, data)
		}
	}
}
