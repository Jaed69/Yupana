package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	_ "image/jpeg"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/zxing-cpp/zxing-cpp/wrappers/go/zxingcpp"
)

func decode(path string) *string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil
	}
	bcs, err := zxingcpp.ReadBarcodes(img, zxingcpp.WithFormats(zxingcpp.BarcodeFormatQRCode), zxingcpp.TryHarder(true))
	if err != nil {
		return nil
	}
	var out *string
	for _, b := range bcs {
		if out == nil && b.IsValid() {
			s := b.Text()
			out = &s
		}
		b.Close()
	}
	return out
}

func rss() string {
	f, _ := os.Open("/proc/self/status")
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "VmRSS") {
			return strings.TrimSpace(strings.TrimPrefix(sc.Text(), "VmRSS:"))
		}
	}
	return "?"
}

func main() {
	dir := os.Args[2]
	if os.Args[1] == "serve" {
		http.HandleFunc("/decode", func(w http.ResponseWriter, r *http.Request) {
			s := decode(filepath.Join(dir, r.URL.Query().Get("f")))
			json.NewEncoder(w).Encode(map[string]any{"text": s})
		})
		http.HandleFunc("/rss", func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, rss()) })
		http.ListenAndServe(":8000", nil)
		return
	}
	files, _ := filepath.Glob(filepath.Join(dir, "*.jpg"))
	sort.Strings(files)
	res := map[string]*string{}
	t0 := time.Now()
	for _, f := range files {
		res[filepath.Base(f)] = decode(f)
	}
	ms := float64(time.Since(t0).Microseconds()) / 1000 / float64(len(files))
	json.NewEncoder(os.Stdout).Encode(map[string]any{
		"results":      map[string]any{"go_zxingcpp_cgo": res},
		"ms_per_image": map[string]float64{"go_zxingcpp_cgo": ms},
	})
}
