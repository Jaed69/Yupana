package main

import (
	"encoding/json"
	"image"
	_ "image/jpeg"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

func decode(path string, hints map[gozxing.DecodeHintType]interface{}, hybrid bool) *string {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil
	}
	src := gozxing.NewLuminanceSourceFromImage(img)
	var bmp *gozxing.BinaryBitmap
	if hybrid {
		bmp, err = gozxing.NewBinaryBitmap(gozxing.NewHybridBinarizer(src))
	} else {
		bmp, err = gozxing.NewBinaryBitmap(gozxing.NewGlobalHistgramBinarizer(src))
	}
	if err != nil {
		return nil
	}
	r, err := qrcode.NewQRCodeReader().Decode(bmp, hints)
	if err != nil {
		return nil
	}
	s := r.GetText()
	return &s
}

func main() {
	files, _ := filepath.Glob(filepath.Join(os.Args[1], "*.jpg"))
	sort.Strings(files)
	tryHarder := map[gozxing.DecodeHintType]interface{}{gozxing.DecodeHintType_TRY_HARDER: true}
	variants := map[string]func(string) *string{
		"go_gozxing":           func(p string) *string { return decode(p, nil, true) },
		"go_gozxing_tryharder": func(p string) *string { return decode(p, tryHarder, true) },
		"go_gozxing_tryharder_global": func(p string) *string {
			if s := decode(p, tryHarder, true); s != nil {
				return s
			}
			return decode(p, tryHarder, false)
		},
	}
	results := map[string]map[string]*string{}
	ms := map[string]float64{}
	for name, fn := range variants {
		res := map[string]*string{}
		t0 := time.Now()
		for _, f := range files {
			res[filepath.Base(f)] = fn(f)
		}
		ms[name] = float64(time.Since(t0).Microseconds()) / 1000 / float64(len(files))
		results[name] = res
	}
	json.NewEncoder(os.Stdout).Encode(map[string]any{"results": results, "ms_per_image": ms})
}
