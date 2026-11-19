package main

// Ban 3 — khoa cai / chay phan mem la.
// Windows Home khong co AppLocker, nen lam bang cach: moi giay chup danh sach tien trinh, tien trinh MOI
// nao chay tu cho nguoi dung ghi duoc (Downloads, Desktop, Temp, USB, AppData moi...) thi TAM DUNG ngay,
// gui yeu cau len server de Sep duyet tren dien thoai. Dong y => cho chay tiep; tu choi / qua 2 phut => dong.
//
// Khong hoi khi:
//   - file nam trong Windows / Program Files (chi Admin moi ghi duoc => Sep da cai)        [tru Microsoft Store]
//   - tien trinh dang chay voi quyen Admin (da nhap mat khau Admin => Sep dang ngoi do)
//   - tien trinh con cua mot chuong trinh vua duoc duyet (bo cai giai nen ra file .tmp roi chay)
//   - thu muc app co san luc bat khoa (nen.json) hoac Sep da "luon cho" (cho_thu_muc), file Sep da cho (cho_hash)
// Tien trinh dang chay luc phan mem khoi dong khong bi hoi (chi xet tien trinh moi).

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const ChoDuyet = 2*time.Minute + 10*time.Second // server het han o 2 phut; agent cho them 10 giay cho tre mang

type TienTrinh struct {
	PID, Cha uint32
	Ten      string
}

type ChiTietTT struct {
	DuongDan  string
	Phien     uint32
	NangQuyen bool
}

// Phan phu thuoc he dieu hanh — Windows o may_windows.go, ban gia o khoacai_test.go.
type NenTang interface {
	QuetTienTrinh() []TienTrinh
	ChiTiet(pid uint32) (ChiTietTT, error)
	TamDung(pid uint32) error
	TiepTuc(pid uint32) error
	Giet(pid uint32) error
	PhienManHinh() uint32
	Bao(tieuDe, noiDung string)
}

type cho struct {
	pids []uint32
	ten  string
	sha  string
	han  time.Time
}

type KhoaCai struct {
	nt      NenTang
	xin     func(ten, duong, sha string, kt int64) (string, error)
	suKien  func(loai, noiDung string)
	heThong []string // tien to duong dan khong bao gio hoi (chu thuong, co \ cuoi)
	nen     []string // thu muc app co san luc bat khoa

	mu          sync.Mutex
	daThay      map[uint32]bool
	duocChay    map[uint32]bool
	dangCho     map[string]*cho
	choTheoHash map[string]string
	lanDau      bool
}

func MoiKhoaCai(nt NenTang, heThong, nen []string, xin func(ten, duong, sha string, kt int64) (string, error), suKien func(string, string)) *KhoaCai {
	return &KhoaCai{nt: nt, xin: xin, suKien: suKien, heThong: heThong, nen: nen,
		daThay: map[uint32]bool{}, duocChay: map[uint32]bool{}, dangCho: map[string]*cho{}, choTheoHash: map[string]string{}, lanDau: true}
}

func coTienTo(p string, ds []string) bool {
	for _, t := range ds {
		if t != "" && strings.HasPrefix(p, t) {
			return true
		}
	}
	return false
}

func co(ds []string, x string) bool {
	for _, v := range ds {
		if v == x {
			return true
		}
	}
	return false
}

func bamFile(p string) (string, int64, error) {
	f, err := os.Open(p)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()
	h := sha256.New()
	n, err := io.Copy(h, f)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

func (k *KhoaCai) ConCho() int {
	k.mu.Lock()
	defer k.mu.Unlock()
	return len(k.dangCho)
}

func (k *KhoaCai) DatNen(nen []string) {
	k.mu.Lock()
	k.nen = nen
	k.mu.Unlock()
}

// Nhip: goi moi giay.
func (k *KhoaCai) Nhip(b *Ban, now time.Time) {
	ds := k.nt.QuetTienTrinh()
	k.mu.Lock()
	defer k.mu.Unlock()

	moi := map[uint32]bool{}
	var tinMoi []TienTrinh
	for _, tt := range ds {
		moi[tt.PID] = true
		if !k.daThay[tt.PID] {
			tinMoi = append(tinMoi, tt)
		}
	}
	for pid := range k.duocChay {
		if !moi[pid] {
			delete(k.duocChay, pid)
		}
	}
	k.daThay = moi
	dauTien := k.lanDau
	k.lanDau = false

	if b == nil || !b.KhoaCai {
		// Tat khoa giua chung => tha het nhung cai dang cho.
		for id, c := range k.dangCho {
			for _, pid := range c.pids {
				_ = k.nt.TiepTuc(pid)
			}
			delete(k.dangCho, id)
			delete(k.choTheoHash, c.sha)
		}
		return
	}
	if !dauTien {
		phien := k.nt.PhienManHinh()
		for _, tt := range tinMoi {
			k.xet(tt, b, phien, now)
		}
	}
	for id, c := range k.dangCho {
		if now.After(c.han) {
			k.ketThuc(id, c, false, "Hết thời gian chờ", c.ten+" đã bị đóng vì bố mẹ chưa trả lời. Con nhắn bố mẹ rồi mở lại nhé.")
			k.suKien("het_han", "Hết 2 phút chờ duyệt, đã đóng "+c.ten)
		}
	}
}

func (k *KhoaCai) xet(tt TienTrinh, b *Ban, phien uint32, now time.Time) {
	if k.duocChay[tt.Cha] {
		k.duocChay[tt.PID] = true
		return
	}
	ct, err := k.nt.ChiTiet(tt.PID)
	if err != nil || ct.DuongDan == "" {
		return // tien trinh he thong duoc bao ve / da thoat
	}
	if ct.Phien == 0 || ct.Phien != phien || ct.NangQuyen {
		return
	}
	p := strings.ToLower(ct.DuongDan)
	ten := filepath.Base(ct.DuongDan)
	laStore := strings.EqualFold(ten, "WinStore.App.exe")
	if !laStore && coTienTo(p, k.heThong) {
		return
	}
	if coTienTo(p, k.nen) || coTienTo(p, b.ChoThuMuc) {
		return
	}
	if err := k.nt.TamDung(tt.PID); err != nil {
		k.suKien("loi", fmt.Sprintf("Không tạm dừng được %s: %v — đóng luôn", ten, err))
		_ = k.nt.Giet(tt.PID)
		return
	}
	sha, kt, err := bamFile(ct.DuongDan)
	if err != nil {
		_ = k.nt.Giet(tt.PID)
		k.suKien("loi", fmt.Sprintf("Không đọc được %s để kiểm: %v — đã đóng", ten, err))
		return
	}
	if co(b.CamHash, sha) {
		_ = k.nt.Giet(tt.PID)
		k.nt.Bao("Bố mẹ không cho phép", ten+" đã bị bố mẹ cấm trên máy này.")
		k.suKien("cam", "Đã đóng "+ten+" (file bị cấm)")
		return
	}
	if co(b.ChoHash, sha) {
		_ = k.nt.TiepTuc(tt.PID)
		k.duocChay[tt.PID] = true
		return
	}
	if id, ok := k.choTheoHash[sha]; ok {
		if c := k.dangCho[id]; c != nil {
			c.pids = append(c.pids, tt.PID)
			return
		}
	}
	id, err := k.xin(ten, ct.DuongDan, sha, kt)
	if err != nil {
		_ = k.nt.Giet(tt.PID)
		k.nt.Bao("Chưa xin phép được", "Máy không liên lạc được với bố mẹ nên "+ten+" đã bị đóng. Thử lại khi có mạng.")
		k.suKien("loi", "Không gửi được yêu cầu duyệt "+ten+": "+err.Error())
		return
	}
	k.dangCho[id] = &cho{pids: []uint32{tt.PID}, ten: ten, sha: sha, han: now.Add(ChoDuyet)}
	k.choTheoHash[sha] = id
	k.nt.Bao("Đang chờ bố mẹ duyệt", ten+" cần bố mẹ cho phép mới chạy được.\nChương trình đang tạm dừng, tối đa 2 phút.")
}

func (k *KhoaCai) ketThuc(id string, c *cho, choChay bool, tieuDe, noiDung string) {
	for _, pid := range c.pids {
		if choChay {
			_ = k.nt.TiepTuc(pid)
			k.duocChay[pid] = true
		} else {
			_ = k.nt.Giet(pid)
		}
	}
	delete(k.dangCho, id)
	delete(k.choTheoHash, c.sha)
	k.nt.Bao(tieuDe, noiDung)
}

// Sep tra loi tren dien thoai (den qua /api/may/hoi -> quyet_dinh).
func (k *KhoaCai) XuLyQuyetDinh(qd QuyetDinh) bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	c := k.dangCho[qd.ID]
	if c == nil {
		return false
	}
	if qd.TrangThai == "dong_y" {
		k.ketThuc(qd.ID, c, true, "Bố mẹ đã cho phép", c.ten+" được chạy tiếp.")
		k.suKien("cho_chay", "Đã cho chạy "+c.ten)
	} else {
		k.ketThuc(qd.ID, c, false, "Bố mẹ chưa cho phép", c.ten+" đã bị đóng.")
		k.suKien("dong", "Đã đóng "+c.ten+" theo quyết định của bố mẹ")
	}
	return true
}

// Thu muc "app" cua mot file .exe trong AppData/ProgramData — cung cach voi worker/src/luat.js
// quyTacTuDuongDan, nhung voi ...\Microsoft\ thi lay them mot cap (Microsoft\OneDrive\).
func thuMucApp(p string) string {
	p = strings.ToLower(p)
	phan := strings.Split(p, `\`)
	// c: users <u> appdata local [programs|microsoft] <app> ...
	if len(phan) >= 7 && phan[1] == "users" && phan[3] == "appdata" && (phan[4] == "local" || phan[4] == "roaming" || phan[4] == "locallow") {
		n := 6
		if phan[5] == "programs" || phan[5] == "microsoft" {
			n = 7
		}
		if phan[5] == "temp" || phan[5] == "packages" || len(phan) <= n {
			return ""
		}
		return strings.Join(phan[:n], `\`) + `\`
	}
	if len(phan) >= 4 && phan[1] == "programdata" {
		n := 3
		if phan[2] == "microsoft" {
			n = 4
		}
		if len(phan) <= n || phan[2] == "quanlycon" {
			return ""
		}
		return strings.Join(phan[:n], `\`) + `\`
	}
	return ""
}
