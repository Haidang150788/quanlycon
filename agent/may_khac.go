//go:build !windows

package main

// Ban gia de chay `go test` / `thu` tren Mac: khong dong gi, khong sua he thong, chi in ra.

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

func thuMucDuLieu() string { return "./du-lieu-thu" }

func apDungChan(mien, url []string) error {
	log.Printf("(gia) ap dung %d ten mien, %d url", len(mien), len(url))
	return nil
}
func datChongLach() error                       { return nil }
func dongTienTrinh(ds map[string]bool) []string { return nil }
func baoNguoiDung(tieuDe, noiDung string)       { log.Printf("(gia) BAO: %s — %s", tieuDe, noiDung) }
func khoaManHinh() error                        { log.Print("(gia) KHOA MAN HINH"); return nil }
func laDichVu() bool                            { return false }
func chayDichVu(f func(<-chan struct{}))        {}
func caiDat(token, server string) error         { return fmt.Errorf("chỉ cài được trên Windows") }
func caiTay()                                   { caiTayDaAdmin() }
func goCaiDat() error                           { return fmt.Errorf("chỉ gỡ được trên Windows") }
func choCtrlC(dung chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() { <-c; close(dung) }()
}

func nenTangThat() NenTang    { return nil } // tren Mac khong co khoa cai that
func thuMucHeThong() []string { return nil }
func quetNenApp() []string    { return nil }

// Ban gia de thu tren Mac: tra ve cua so gia + anh gia (khong chup man hinh Mac).
func chayTrongPhien(lenh string, gioiHan time.Duration) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString(`{"exe":"chrome.exe","tieu_de":"(thử trên Mac) YouTube - Google Chrome","ranh":3}` + "\n")
	if strings.HasSuffix(lenh, " anh") {
		img := image.NewRGBA(image.Rect(0, 0, 1920, 1080))
		for y := 0; y < 1080; y++ {
			for x := 0; x < 1920; x++ {
				i := img.PixOffset(x, y)
				img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = uint8(x/8), uint8(y/5), 160, 255
			}
		}
		if err := jpeg.Encode(&b, thuNho(img, 1600), &jpeg.Options{Quality: 60}); err != nil {
			return nil, err
		}
	}
	return b.Bytes(), nil
}
func chayNhin(coAnh bool) error { return fmt.Errorf("chỉ chạy được trên Windows") }
