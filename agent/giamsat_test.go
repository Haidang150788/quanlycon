package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"testing"
)

func TestThuNhoGiuTyLe(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3840, 2160))
	n := thuNho(img, 1600).Bounds()
	if n.Dx() != 1600 || n.Dy() != 900 {
		t.Fatalf("ra %v", n)
	}
	if thuNho(image.NewRGBA(image.Rect(0, 0, 1280, 720)), 1600).Bounds().Dx() != 1280 {
		t.Fatal("anh nho hon thi giu nguyen")
	}
}

func TestTachNhin(t *testing.T) {
	var b bytes.Buffer
	b.WriteString(`{"exe":"chrome.exe","tieu_de":"YouTube","ranh":12}` + "\n")
	_ = jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 10, 10)), nil)
	n, anh, err := tachNhin(b.Bytes())
	if err != nil || n.Exe != "chrome.exe" || n.Ranh != 12 || len(anh) == 0 {
		t.Fatalf("%v %v %d", n, err, len(anh))
	}
	n, anh, err = tachNhin([]byte(`{"exe":"a.exe","tieu_de":"","ranh":0}` + "\n"))
	if err != nil || len(anh) != 0 {
		t.Fatal("khong anh van phai doc duoc JSON")
	}
	if _, _, err = tachNhin([]byte("LỖI: abc")); err == nil {
		t.Fatal("rac phai bao loi")
	}
	if _, _, err = tachNhin([]byte(`{"exe":""}` + "\nkhongphaijpeg")); err == nil {
		t.Fatal("anh sai dinh dang phai bao loi")
	}
}

func TestAnhNho(t *testing.T) {
	var b bytes.Buffer
	_ = jpeg.Encode(&b, image.NewRGBA(image.Rect(0, 0, 1600, 900)), nil)
	nho := taoAnhNho(b.Bytes())
	img, err := jpeg.Decode(bytes.NewReader(nho))
	if err != nil || img.Bounds().Dx() != 400 || img.Bounds().Dy() != 225 {
		t.Fatalf("anh nho sai: %v", err)
	}
	if taoAnhNho([]byte("rac")) != nil {
		t.Fatal("rac phai ra nil")
	}
}
