package main

// Ban 5 — YouTube chi xem kenh duoc chon trong gio chan.
// Extension (thu muc ../extension) duoc ep cai vao Chrome/Edge bang chinh sach ExtensionInstallForcelist.
// No hoi dich vu nay (GET http://127.0.0.1:47123/ext/trang-thai) de biet dang gio chan khong + danh sach kenh,
// roi tu chan video/kenh khac ngay tren trang.
//
// "Dong khi khong chac": YouTube chi duoc bo khoi URLBlocklist cua MOT trinh duyet khi extension trong
// chinh trinh duyet do vua bao ve (< extConSong). Extension chua cai duoc / chua chay / bi tat => YouTube van
// bi chan han nhu truoc. Con 3 dieu kien nua: dang co danh sach kenh, server gui mien_theo_kenh (YouTube dang
// nam trong danh sach chan), va ma extension cua trinh duyet do da cau hinh.

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type KenhYT struct {
	ID     string `json:"id"`
	Handle string `json:"handle"`
	Ten    string `json:"ten"`
}

// Kho extension mac dinh. Windows khong vao domain chi ep cai duoc extension co tren kho chinh hang.
const (
	khoChrome = "https://clients2.google.com/service/update2/crx"
	khoEdge   = "https://edge.microsoft.com/extensionwebstorebase/v1/crx"
)

// Lau hon thoi gian nay khong nghe extension bao => coi nhu khong co extension, chan lai YouTube.
const extConSong = 10 * time.Minute

type DanhSachChan struct {
	Mien      []string // file hosts (moi ung dung)
	UrlChrome []string
	UrlEdge   []string
	// Da bo YouTube khoi hosts => trinh duyet khac (khong co extension, khong doc URLBlocklist cua Chrome/Edge)
	// xem YouTube tu do duoc => dong chung trong gio chan.
	DongTrinhDuyetKhac bool
}

var trinhDuyetKhongExt = []string{"firefox.exe", "opera.exe", "brave.exe", "browser.exe", "coccoc.exe", "vivaldi.exe",
	"chromium.exe", "waterfox.exe", "librewolf.exe", "floorp.exe", "thorium.exe", "yandex.exe"}

func boKhoi(ds []string, bo map[string]bool) []string {
	out := []string{}
	for _, x := range ds {
		if !bo[x] {
			out = append(out, x)
		}
	}
	return out
}

// TinhDanhSachChan: chromeSong/edgeSong = extension trong trinh duyet do dang chay.
func TinhDanhSachChan(b *Ban, biChan, chromeSong, edgeSong bool) DanhSachChan {
	if b == nil || !biChan {
		return DanhSachChan{}
	}
	d := DanhSachChan{Mien: b.TenMien, UrlChrome: b.UrlChan, UrlEdge: b.UrlChan}
	if len(b.KenhCho) == 0 || len(b.MienTheoKenh) == 0 {
		return d
	}
	chromeSong = chromeSong && b.ExtChrome != ""
	edgeSong = edgeSong && b.ExtEdge != ""
	bo := map[string]bool{}
	for _, m := range b.MienTheoKenh {
		bo[m] = true
	}
	if chromeSong {
		d.UrlChrome = boKhoi(b.UrlChan, bo)
	}
	if edgeSong {
		d.UrlEdge = boKhoi(b.UrlChan, bo)
	}
	// hosts chan moi trinh duyet cung luc: co it nhat mot trinh duyet xem duoc theo kenh thi phai bo,
	// trinh duyet con lai van bi URLBlocklist chan.
	if chromeSong || edgeSong {
		d.Mien = boKhoi(b.TenMien, bo)
		d.DongTrinhDuyetKhac = true
	}
	return d
}

// HanCheYoutube: extension can chan theo kenh khong (dang gio chan + YouTube nam trong danh sach chan).
func HanCheYoutube(b *Ban, nowMs int64) bool {
	if !DangChan(b, nowMs) {
		return false
	}
	for _, u := range b.UrlChan {
		if u == "youtube.com" {
			return true
		}
	}
	return false
}

// Chuoi ExtensionInstallForcelist: "ma" => "ma;<kho mac dinh cua trinh duyet>", "ma;url" giu nguyen.
func dongForcelist(ma, khoMacDinh string) string {
	ma = strings.TrimSpace(ma)
	if ma == "" || strings.Contains(ma, ";") {
		return ma
	}
	return ma + ";" + khoMacDinh
}

func maExt(s string) string { return strings.SplitN(strings.TrimSpace(s), ";", 2)[0] }

// Extension nao goi: Origin "chrome-extension://<ma>" (trang web thuong khong gia duoc Origin nay).
// Tra "chrome" | "edge" | "" (khong phai extension cua minh). Extension goi bang POST: GET tu service worker
// co quyen host thi Chrome khong gui Origin (do thuc te 22/11/2026, Chrome for Testing 150).
func trinhDuyetCuaExt(b *Ban, origin, ua string) string {
	if b == nil {
		return ""
	}
	ma := ""
	for _, tien := range []string{"chrome-extension://", "extension://"} {
		if strings.HasPrefix(origin, tien) {
			ma = strings.TrimSuffix(strings.TrimPrefix(origin, tien), "/")
		}
	}
	if ma == "" {
		return ""
	}
	laEdge := strings.Contains(ua, " Edg/")
	if laEdge && ma == maExt(b.ExtEdge) {
		return "edge"
	}
	if !laEdge && ma == maExt(b.ExtChrome) {
		return "chrome"
	}
	return ""
}

type ttExt struct {
	HanChe bool     `json:"han_che"`
	Kenh   []KenhYT `json:"kenh"`
	ConMs  int64    `json:"con_ms"`
	Now    int64    `json:"now"`
	Nhan   bool     `json:"nhan"` // dich vu co nhan extension nay la cua minh khong (de go loi)
}

func (t *TacNhan) apiExt(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if r.Method != "POST" {
		http.Error(w, "POST", 405)
		return
	}
	if !strings.HasPrefix(origin, "chrome-extension://") && !strings.HasPrefix(origin, "extension://") {
		http.Error(w, "forbidden", 403)
		return
	}
	now := t.gioHienTai()
	t.mu.Lock()
	b := t.ban
	td := trinhDuyetCuaExt(b, origin, r.UserAgent())
	if td != "" {
		if t.extLuc == nil {
			t.extLuc = map[string]time.Time{}
		}
		t.extLuc[td] = time.Now()
	}
	t.mu.Unlock()
	kq := ttExt{HanChe: true, Kenh: []KenhYT{}, Now: now, Nhan: td != ""}
	if b != nil {
		kq.HanChe = HanCheYoutube(b, now)
		if b.KenhCho != nil {
			kq.Kenh = b.KenhCho
		}
		kq.ConMs = ConBaoLauThiChan(b, now)
	}
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.Header().Set("access-control-allow-origin", origin)
	_ = json.NewEncoder(w).Encode(kq)
}

func (t *TacNhan) extDangSong() (chrome, edge bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	return time.Since(t.extLuc["chrome"]) < extConSong, time.Since(t.extLuc["edge"]) < extConSong
}
