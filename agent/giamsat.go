package main

// Ban 4 — chup man hinh + nhat ky cua so dang dung.
// Dich vu chay o phien 0 (SYSTEM) nen KHONG nhin thay man hinh cua con. Moi lan can, dich vu chay
// `quanlycon.exe nhin [anh]` NGAY TRONG PHIEN CUA CON (chup_windows.go chayTrongPhien), chuong trinh do
// in ra 1 dong JSON (cua so dang dung, so giay khong dung may) + anh JPEG qua ong dan ve cho dich vu.
//   - moi 30 giay: ghi cua so dang dung (chi gui khi doi cua so hoac moi 2 phut)
//   - moi chup_phut phut (Sep chon 5/10/15/30, 0 = tat): chup man hinh neu con dang dung may (< 5 phut ranh)
//   - Sep bam "Chup ngay"; con mo app bi chan; con mo phan mem la => chup ngay (lam bang chung)

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type NhinThay struct {
	Exe    string `json:"exe"`
	TieuDe string `json:"tieu_de"`
	Ranh   int    `json:"ranh"` // giay khong cham chuot/ban phim
}

type HoatDong struct {
	Ts     int64  `json:"ts"`
	Exe    string `json:"exe"`
	TieuDe string `json:"tieu_de"`
	Ranh   int    `json:"ranh"`
}

// Tach dau ra cua `nhin`: dong JSON dau tien, phan con lai la JPEG (co the rong).
func tachNhin(raw []byte) (NhinThay, []byte, error) {
	var n NhinThay
	i := bytes.IndexByte(raw, '\n')
	if i < 0 {
		return n, nil, fmt.Errorf("đầu ra không có dòng JSON: %q", string(raw[:min(len(raw), 200)]))
	}
	if err := json.Unmarshal(raw[:i], &n); err != nil {
		return n, nil, err
	}
	anh := raw[i+1:]
	if len(anh) > 0 && (len(anh) < 4 || anh[0] != 0xff || anh[1] != 0xd8) {
		return n, nil, fmt.Errorf("ảnh không phải JPEG")
	}
	return n, anh, nil
}

type GiamSat struct {
	t           *TacNhan
	mu          sync.Mutex // mot lan chup mot luc
	cuoi        HoatDong
	lanGhi      time.Time
	lanChup     time.Time
	lanSuKien   time.Time
	hang        []HoatDong
	hangMu      sync.Mutex
	loiLienTiep int
}

func (g *GiamSat) layHang() []HoatDong {
	g.hangMu.Lock()
	defer g.hangMu.Unlock()
	h := g.hang
	g.hang = nil
	return h
}

func (g *GiamSat) traHang(h []HoatDong) {
	g.hangMu.Lock()
	defer g.hangMu.Unlock()
	if len(g.hang)+len(h) <= 2000 {
		g.hang = append(h, g.hang...)
	}
}

// Moi 30 giay.
func (g *GiamSat) ghiCuaSo() {
	if !g.mu.TryLock() {
		return
	}
	defer g.mu.Unlock()
	raw, err := chayTrongPhien("nhin", 15*time.Second)
	var n NhinThay
	if err == nil {
		n, _, err = tachNhin(raw)
	}
	if err != nil {
		g.loiLienTiep++
		if g.loiLienTiep == 3 {
			log.Printf("nhin loi 3 lan lien: %v", err)
		}
		n = NhinThay{TieuDe: "(không ai đăng nhập hoặc màn hình đang khoá)"}
	} else {
		g.loiLienTiep = 0
	}
	hd := HoatDong{Ts: time.Now().UnixMilli(), Exe: n.Exe, TieuDe: n.TieuDe, Ranh: n.Ranh}
	doi := hd.Exe != g.cuoi.Exe || hd.TieuDe != g.cuoi.TieuDe
	if doi || time.Since(g.lanGhi) >= 2*time.Minute {
		g.hangMu.Lock()
		if len(g.hang) < 2000 {
			g.hang = append(g.hang, hd)
		}
		g.hangMu.Unlock()
		g.cuoi = hd
		g.lanGhi = time.Now()
	}
	// Chup dinh ky ngay trong cung nhip (dung lai thong tin "ranh" vua lay).
	g.t.mu.Lock()
	b := g.t.ban
	g.t.mu.Unlock()
	if err == nil && b != nil && b.ChupPhut > 0 && n.Ranh < 300 && time.Since(g.lanChup) >= time.Duration(b.ChupPhut)*time.Minute {
		g.chupKhoa("định kỳ")
	}
}

// Chup ngay (khong quan tam con co dang dung may khong). Goi tu goroutine rieng.
func (g *GiamSat) Chup(lyDo string, giuNhip bool) {
	if giuNhip {
		// Su kien (app bi chan, phan mem la): toi da 1 anh / 2 phut de khong do day bo nho.
		g.mu.Lock()
		gan := time.Since(g.lanSuKien) < 2*time.Minute
		if !gan {
			g.lanSuKien = time.Now()
		}
		g.mu.Unlock()
		if gan {
			return
		}
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.chupKhoa(lyDo)
}

// Goi khi da giu g.mu.
func (g *GiamSat) chupKhoa(lyDo string) {
	g.lanChup = time.Now()
	raw, err := chayTrongPhien("nhin anh", 30*time.Second)
	if err != nil {
		log.Printf("chup man hinh loi: %v", err)
		return
	}
	n, anh, err := tachNhin(raw)
	if err != nil || len(anh) == 0 {
		log.Printf("chup man hinh khong ra anh: %v", err)
		return
	}
	if err := g.t.guiAnh(anh, taoAnhNho(anh), lyDo, n.Exe+" — "+n.TieuDe); err != nil {
		log.Printf("gui anh loi: %v", err)
	}
}

// Anh nho rong 400px cho luoi anh tren dien thoai (~15KB thay vi ~200KB). Loi thi bo qua.
func taoAnhNho(anh []byte) []byte {
	img, err := jpeg.Decode(bytes.NewReader(anh))
	if err != nil {
		return nil
	}
	var b bytes.Buffer
	if jpeg.Encode(&b, thuNho(img, 400), &jpeg.Options{Quality: 55}) != nil {
		return nil
	}
	return b.Bytes()
}

func (t *TacNhan) guiAnh(anh, nho []byte, lyDo, cuaSo string) error {
	q := url.Values{}
	q.Set("may", t.ch.ID)
	q.Set("ly_do", lyDo)
	q.Set("cua_so", cuaSo)
	q.Set("ts", fmt.Sprint(time.Now().UnixMilli()))
	req, _ := http.NewRequest("POST", strings.TrimRight(t.ch.Server, "/")+"/api/may/anh?"+q.Encode(), bytes.NewReader(append(append([]byte{}, nho...), anh...)))
	req.Header.Set("content-type", "image/jpeg")
	req.Header.Set("x-nho-len", fmt.Sprint(len(nho)))
	req.Header.Set("authorization", "Bearer "+t.ch.Token)
	c := &http.Client{Timeout: 60 * time.Second}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 500))
		return fmt.Errorf("server trả %d: %s", res.StatusCode, b)
	}
	return nil
}

// Thu nho anh ve chieu ngang toi da `rong` (lay trung binh o vuong), giu ty le. Man hinh 1920 => 1600.
func thuNho(src image.Image, rong int) image.Image {
	b := src.Bounds()
	if b.Dx() <= rong {
		return src
	}
	cao := b.Dy() * rong / b.Dx()
	dst := image.NewRGBA(image.Rect(0, 0, rong, cao))
	for y := 0; y < cao; y++ {
		y0, y1 := b.Min.Y+y*b.Dy()/cao, b.Min.Y+(y+1)*b.Dy()/cao
		if y1 <= y0 {
			y1 = y0 + 1
		}
		for x := 0; x < rong; x++ {
			x0, x1 := b.Min.X+x*b.Dx()/rong, b.Min.X+(x+1)*b.Dx()/rong
			if x1 <= x0 {
				x1 = x0 + 1
			}
			var r, g, bl, n uint32
			for yy := y0; yy < y1; yy++ {
				for xx := x0; xx < x1; xx++ {
					cr, cg, cb, _ := src.At(xx, yy).RGBA()
					r, g, bl, n = r+cr, g+cg, bl+cb, n+1
				}
			}
			i := dst.PixOffset(x, y)
			dst.Pix[i], dst.Pix[i+1], dst.Pix[i+2], dst.Pix[i+3] = uint8(r/n>>8), uint8(g/n>>8), uint8(bl/n>>8), 255
		}
	}
	return dst
}
