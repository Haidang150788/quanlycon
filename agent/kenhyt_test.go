package main

import (
	"reflect"
	"testing"
)

func banYT() *Ban {
	return &Ban{
		Bat:          true,
		TenMien:      []string{"facebook.com", "music.youtube.com", "youtube.com", "www.youtube.com", "yewtu.be"},
		UrlChan:      []string{"facebook.com", "music.youtube.com", "youtube.com", "www.youtube.com", "yewtu.be"},
		MienTheoKenh: []string{"youtube.com", "www.youtube.com"},
		KenhCho:      []KenhYT{{ID: "UCabcdefghijklmnopqrstuv", Handle: "@HocToan"}},
		ExtChrome:    "abcdefghijklmnopabcdefghijklmnop",
		ExtEdge:      "ponmlkjihgfedcbaponmlkjihgfedcba",
	}
}

var conLai = []string{"facebook.com", "music.youtube.com", "yewtu.be"}

func TestYoutubeTheoKenhChiNoiTrinhDuyetCoExtension(t *testing.T) {
	b := banYT()
	d := TinhDanhSachChan(b, true, true, false)
	if !reflect.DeepEqual(d.UrlChrome, conLai) {
		t.Fatalf("Chrome co extension phai bo YouTube, giu YouTube Music + trang thay the: %v", d.UrlChrome)
	}
	if !reflect.DeepEqual(d.UrlEdge, b.UrlChan) {
		t.Fatalf("Edge khong co extension phai chan YouTube han: %v", d.UrlEdge)
	}
	if !reflect.DeepEqual(d.Mien, conLai) {
		t.Fatalf("hosts phai bo YouTube khi co it nhat mot trinh duyet xem theo kenh: %v", d.Mien)
	}
	if !d.DongTrinhDuyetKhac {
		t.Fatal("hosts da noi YouTube => phai dong Firefox/Coc Coc... trong gio chan")
	}
}

func TestYoutubeDongKhiKhongChac(t *testing.T) {
	b := banYT()
	if d := TinhDanhSachChan(b, true, false, false); !reflect.DeepEqual(d.UrlChrome, b.UrlChan) || !reflect.DeepEqual(d.Mien, b.TenMien) || d.DongTrinhDuyetKhac {
		t.Fatal("khong nghe extension nao => chan han nhu cu")
	}
	b.KenhCho = nil
	if d := TinhDanhSachChan(b, true, true, true); !reflect.DeepEqual(d.UrlChrome, b.UrlChan) {
		t.Fatal("danh sach kenh rong => chan han")
	}
	b = banYT()
	b.MienTheoKenh = nil // server khong gui = YouTube khong nam trong danh sach chan / chua co kenh
	if d := TinhDanhSachChan(b, true, true, true); !reflect.DeepEqual(d.UrlEdge, b.UrlChan) {
		t.Fatal("khong co mien_theo_kenh => khong noi gi")
	}
	b = banYT()
	b.ExtChrome = "" // extension Chrome chua cau hinh: heartbeat (neu co) khong tinh
	if d := TinhDanhSachChan(b, true, true, false); !reflect.DeepEqual(d.UrlChrome, b.UrlChan) {
		t.Fatal("chua cau hinh ma extension => chan han")
	}
}

func TestGioMoKhongChanGi(t *testing.T) {
	if d := TinhDanhSachChan(banYT(), false, true, true); len(d.Mien)+len(d.UrlChrome)+len(d.UrlEdge) != 0 {
		t.Fatal("gio giai lao: khong chan gi")
	}
}

func TestNhanDungExtension(t *testing.T) {
	b := banYT()
	ua := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	if trinhDuyetCuaExt(b, "chrome-extension://abcdefghijklmnopabcdefghijklmnop", ua) != "chrome" {
		t.Fatal("phai nhan extension Chrome")
	}
	if trinhDuyetCuaExt(b, "chrome-extension://ponmlkjihgfedcbaponmlkjihgfedcba", ua+" Edg/131.0.0.0") != "edge" {
		t.Fatal("phai nhan extension Edge")
	}
	if trinhDuyetCuaExt(b, "chrome-extension://abcdefghijklmnopabcdefghijklmnop", ua+" Edg/131.0.0.0") != "" {
		t.Fatal("ma Chrome goi tu Edge (ma Edge khac) => khong tinh")
	}
	if trinhDuyetCuaExt(b, "https://evil.com", ua) != "" || trinhDuyetCuaExt(b, "chrome-extension://khac", ua) != "" {
		t.Fatal("trang web / extension la => khong tinh")
	}
	// Edge dung ban tren Chrome Web Store: cung ma, them ;update_url
	b.ExtEdge = "abcdefghijklmnopabcdefghijklmnop;" + khoChrome
	if trinhDuyetCuaExt(b, "chrome-extension://abcdefghijklmnopabcdefghijklmnop", ua+" Edg/131.0.0.0") != "edge" {
		t.Fatal("Edge dung ma Chrome Web Store")
	}
}

func TestDongForcelist(t *testing.T) {
	if dongForcelist("abc", khoChrome) != "abc;"+khoChrome || dongForcelist("abc;https://x/y", khoEdge) != "abc;https://x/y" || dongForcelist(" ", khoEdge) != "" {
		t.Fatal("sai chuoi forcelist")
	}
}

func TestHanCheYoutube(t *testing.T) {
	b := banYT()
	now := luc(15, 0)
	if !HanCheYoutube(b, now) {
		t.Fatal("dang chan + YouTube trong danh sach => han che")
	}
	b.GiaiLao = []Khung{{Tu: "14:00", Den: "16:00"}}
	if HanCheYoutube(b, now) {
		t.Fatal("gio giai lao => xem thoai mai")
	}
	b = banYT()
	b.UrlChan = []string{"facebook.com"}
	if HanCheYoutube(b, now) {
		t.Fatal("bo me khong chan YouTube => xem thoai mai")
	}
}
