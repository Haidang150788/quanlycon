package main

// Luat "luc nay co chan khong". BAN SAO cua worker/src/luat.js — sua mot ben phai sua ben kia.
// Gio Viet Nam = UTC+7 co dinh. Thoi diem "now" la GIO SERVER (xem dongHo trong main.go),
// khong phai dong ho may con, de con doi gio Windows khong lach duoc.

import (
	"strconv"
	"strings"
)

type Khung struct {
	Tu  string `json:"tu"`
	Den string `json:"den"`
}

// Ban = dung khuon JSON cua POST /api/may/hoi (worker/src/index.js banChoMay).
type Ban struct {
	Now            int64    `json:"now"`
	Bat            bool     `json:"bat"`
	TenMien        []string `json:"ten_mien"`
	UrlChan        []string `json:"url_chan"`
	TienTrinh      []string `json:"tien_trinh"`
	GiaiLao        []Khung  `json:"giai_lao"`
	DongTrinhDuyet bool     `json:"dong_trinh_duyet"`
	MoTamDen       int64    `json:"mo_tam_den"`
	KhoaLuc        int64    `json:"khoa_luc"`
	// Ban 3 — khoa cai phan mem
	KhoaCai   bool        `json:"khoa_cai"`
	ChoThuMuc []string    `json:"cho_thu_muc"` // tien to duong dan, chu thuong
	ChoHash   []string    `json:"cho_hash"`
	CamHash   []string    `json:"cam_hash"`
	QuyetDinh []QuyetDinh `json:"quyet_dinh"` // yeu cau Sep vua tra loi (15 phut gan nhat)
	// Ban 4 — chup man hinh
	ChupPhut int   `json:"chup_phut"` // 0 = tat
	ChupLuc  int64 `json:"chup_luc"`  // Sep bam "Chup ngay"
	// Ban 5 — YouTube theo kenh (kenhyt.go)
	MienTheoKenh []string `json:"mien_theo_kenh"` // ten mien YouTube duoc noi khi extension dang chay
	KenhCho      []KenhYT `json:"kenh_cho"`
	ExtChrome    string   `json:"ext_chrome"` // "ma" hoac "ma;update_url"
	ExtEdge      string   `json:"ext_edge"`
}

type QuyetDinh struct {
	ID        string `json:"id"`
	Loai      string `json:"loai"`       // gio | chay
	TrangThai string `json:"trang_thai"` // dong_y | tu_choi | het_han
	Phut      int    `json:"phut"`
}

const lechVN = 7 * 3600 * 1000

func phutTrongNgayVN(ms int64) int {
	m := ((ms + lechVN) / 60000) % 1440
	if m < 0 {
		m += 1440
	}
	return int(m)
}

func phutCua(hhmm string) (int, bool) {
	p := strings.SplitN(hhmm, ":", 2)
	if len(p) != 2 {
		return 0, false
	}
	h, e1 := strconv.Atoi(p[0])
	m, e2 := strconv.Atoi(p[1])
	if e1 != nil || e2 != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, false
	}
	return h*60 + m, true
}

// Phut "den" tinh ca phut do. Tu > den => khung qua dem.
func trongKhung(m int, k Khung) bool {
	a, ok1 := phutCua(k.Tu)
	b, ok2 := phutCua(k.Den)
	if !ok1 || !ok2 {
		return false
	}
	if a <= b {
		return m >= a && m <= b
	}
	return m >= a || m <= b
}

func DangChan(b *Ban, nowMs int64) bool {
	if b == nil || !b.Bat {
		return false
	}
	if b.MoTamDen > nowMs {
		return false
	}
	m := phutTrongNgayVN(nowMs)
	for _, k := range b.GiaiLao {
		if trongKhung(m, k) {
			return false
		}
	}
	return true
}

// ConBaoLauThiChan: so ms tu now toi luc bat dau chan lai (0 = dang chan; -1 = trong 24h toi khong chan).
func ConBaoLauThiChan(b *Ban, nowMs int64) int64 {
	if DangChan(b, nowMs) {
		return 0
	}
	if b == nil || !b.Bat {
		return -1
	}
	ung := []int64{}
	if b.MoTamDen > nowMs {
		ung = append(ung, b.MoTamDen)
	}
	dauPhut := (nowMs/60000 + 1) * 60000
	for t := dauPhut; t <= nowMs+24*3600*1000; t += 60000 {
		ung = append(ung, t)
	}
	tot := int64(-1)
	for _, t := range ung {
		if DangChan(b, t) && (tot < 0 || t < tot) {
			tot = t
		}
	}
	if tot < 0 {
		return -1
	}
	return tot - nowMs
}
