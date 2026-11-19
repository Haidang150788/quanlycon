package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type gia struct {
	tt    []TienTrinh
	ct    map[uint32]ChiTietTT
	dung  map[uint32]bool
	chet  map[uint32]bool
	tiep  map[uint32]bool
	bao   []string
	phien uint32
}

func moiGia() *gia {
	return &gia{ct: map[uint32]ChiTietTT{}, dung: map[uint32]bool{}, chet: map[uint32]bool{}, tiep: map[uint32]bool{}, phien: 1}
}
func (g *gia) QuetTienTrinh() []TienTrinh { return g.tt }
func (g *gia) ChiTiet(pid uint32) (ChiTietTT, error) {
	c, ok := g.ct[pid]
	if !ok {
		return c, errors.New("khong co")
	}
	return c, nil
}
func (g *gia) TamDung(pid uint32) error { g.dung[pid] = true; return nil }
func (g *gia) TiepTuc(pid uint32) error { g.tiep[pid] = true; return nil }
func (g *gia) Giet(pid uint32) error    { g.chet[pid] = true; return nil }
func (g *gia) PhienManHinh() uint32     { return g.phien }
func (g *gia) Bao(t, n string)          { g.bao = append(g.bao, t) }

// Tao file that de bam sha256; duong dan "Windows" gia lap bang duong dan tam.
func fileThat(t *testing.T, ten, noiDung string) string {
	p := filepath.Join(t.TempDir(), ten)
	if err := os.WriteFile(p, []byte(noiDung), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func dungBo(t *testing.T) (*gia, *KhoaCai, *[]string) {
	g := moiGia()
	var daXin []string
	k := MoiKhoaCai(g, []string{`c:\windows\`, `c:\program files\`}, []string{`c:\users\con\appdata\local\zoom\`},
		func(ten, d, sha string, kt int64) (string, error) {
			daXin = append(daXin, ten)
			return "id-" + ten, nil
		},
		func(string, string) {})
	// nhip dau: tien trinh co san khong bi xet
	g.tt = []TienTrinh{{PID: 1, Ten: "explorer.exe"}}
	k.Nhip(&Ban{KhoaCai: true}, time.Now())
	return g, k, &daXin
}

func TestBoQuaHeThongNangQuyenPhienKhacVaNen(t *testing.T) {
	g, k, daXin := dungBo(t)
	g.ct[10] = ChiTietTT{DuongDan: `C:\Windows\System32\notepad.exe`, Phien: 1}
	g.ct[11] = ChiTietTT{DuongDan: `C:\Users\Con\Downloads\setup.exe`, Phien: 1, NangQuyen: true}
	g.ct[12] = ChiTietTT{DuongDan: `C:\Users\Con\Downloads\setup.exe`, Phien: 0}
	g.ct[13] = ChiTietTT{DuongDan: `C:\Users\Con\AppData\Local\Zoom\bin\Zoom.exe`, Phien: 1}
	g.tt = append(g.tt, TienTrinh{PID: 10}, TienTrinh{PID: 11}, TienTrinh{PID: 12}, TienTrinh{PID: 13})
	k.Nhip(&Ban{KhoaCai: true}, time.Now())
	if len(g.dung) != 0 || len(*daXin) != 0 {
		t.Fatalf("khong duoc dung/xin: dung=%v xin=%v", g.dung, *daXin)
	}
}

func TestFileLaBiDungVaXinDuyetRoiChoChay(t *testing.T) {
	g, k, daXin := dungBo(t)
	p := fileThat(t, "game.exe", "noi dung game")
	g.ct[20] = ChiTietTT{DuongDan: p, Phien: 1}
	g.tt = append(g.tt, TienTrinh{PID: 20, Cha: 1})
	now := time.Now()
	k.Nhip(&Ban{KhoaCai: true}, now)
	if !g.dung[20] || len(*daXin) != 1 || k.ConCho() != 1 {
		t.Fatalf("phai tam dung + xin: dung=%v xin=%v", g.dung, *daXin)
	}
	// Mo lan nua cung file trong luc cho => khong xin lai
	g.ct[21] = ChiTietTT{DuongDan: p, Phien: 1}
	g.tt = append(g.tt, TienTrinh{PID: 21, Cha: 1})
	k.Nhip(&Ban{KhoaCai: true}, now)
	if len(*daXin) != 1 {
		t.Fatal("cung file dang cho thi khong duoc xin lai")
	}
	k.XuLyQuyetDinh(QuyetDinh{ID: "id-game.exe", TrangThai: "dong_y"})
	if !g.tiep[20] || !g.tiep[21] || k.ConCho() != 0 {
		t.Fatal("dong y phai cho ca hai chay tiep")
	}
	// Tien trinh con cua cai vua duoc duyet (bo cai giai nen) khong bi hoi
	g.ct[22] = ChiTietTT{DuongDan: `C:\Users\Con\AppData\Local\Temp\is-AB\setup.tmp`, Phien: 1}
	g.tt = append(g.tt, TienTrinh{PID: 22, Cha: 20})
	k.Nhip(&Ban{KhoaCai: true}, now)
	if g.dung[22] {
		t.Fatal("tien trinh con cua cai da duyet khong duoc dung")
	}
}

func TestTuChoiVaHetHan(t *testing.T) {
	g, k, _ := dungBo(t)
	p := fileThat(t, "a.exe", "a")
	g.ct[30] = ChiTietTT{DuongDan: p, Phien: 1}
	g.tt = append(g.tt, TienTrinh{PID: 30})
	now := time.Now()
	k.Nhip(&Ban{KhoaCai: true}, now)
	k.XuLyQuyetDinh(QuyetDinh{ID: "id-a.exe", TrangThai: "tu_choi"})
	if !g.chet[30] {
		t.Fatal("tu choi phai dong")
	}
	p2 := fileThat(t, "b.exe", "b")
	g.ct[31] = ChiTietTT{DuongDan: p2, Phien: 1}
	g.tt = append(g.tt, TienTrinh{PID: 31})
	k.Nhip(&Ban{KhoaCai: true}, now)
	k.Nhip(&Ban{KhoaCai: true}, now.Add(ChoDuyet+time.Second))
	if !g.chet[31] || k.ConCho() != 0 {
		t.Fatal("qua han phai dong")
	}
}

func TestHashChoVaCam(t *testing.T) {
	g, k, daXin := dungBo(t)
	p := fileThat(t, "ok.exe", "ok")
	sha, _, _ := bamFile(p)
	g.ct[40] = ChiTietTT{DuongDan: p, Phien: 1}
	g.ct[41] = ChiTietTT{DuongDan: p, Phien: 1}
	g.tt = append(g.tt, TienTrinh{PID: 40})
	k.Nhip(&Ban{KhoaCai: true, ChoHash: []string{sha}}, time.Now())
	if !g.tiep[40] || len(*daXin) != 0 {
		t.Fatal("hash da cho thi chay luon")
	}
	g.tt = append(g.tt, TienTrinh{PID: 41})
	k.Nhip(&Ban{KhoaCai: true, CamHash: []string{sha}}, time.Now())
	if !g.chet[41] {
		t.Fatal("hash bi cam phai dong")
	}
}

func TestTatKhoaThaHetVaKhongXet(t *testing.T) {
	g, k, daXin := dungBo(t)
	p := fileThat(t, "c.exe", "c")
	g.ct[50] = ChiTietTT{DuongDan: p, Phien: 1}
	g.tt = append(g.tt, TienTrinh{PID: 50})
	k.Nhip(&Ban{KhoaCai: true}, time.Now())
	k.Nhip(&Ban{KhoaCai: false}, time.Now())
	if !g.tiep[50] || k.ConCho() != 0 {
		t.Fatal("tat khoa phai tha cai dang cho")
	}
	g.ct[51] = ChiTietTT{DuongDan: p, Phien: 1}
	g.tt = append(g.tt, TienTrinh{PID: 51})
	k.Nhip(&Ban{KhoaCai: false}, time.Now())
	if g.dung[51] || len(*daXin) != 1 {
		t.Fatal("dang tat khoa thi khong xet")
	}
}

func TestMatMangThiDong(t *testing.T) {
	g := moiGia()
	k := MoiKhoaCai(g, nil, nil, func(string, string, string, int64) (string, error) { return "", errors.New("mat mang") }, func(string, string) {})
	k.Nhip(&Ban{KhoaCai: true}, time.Now())
	g.ct[60] = ChiTietTT{DuongDan: fileThat(t, "d.exe", "d"), Phien: 1}
	g.tt = []TienTrinh{{PID: 60}}
	k.Nhip(&Ban{KhoaCai: true}, time.Now())
	if !g.chet[60] {
		t.Fatal("khong xin duoc thi dong")
	}
}

func TestThuMucApp(t *testing.T) {
	ca := map[string]string{
		`C:\Users\Con\AppData\Local\Zoom\bin\Zoom.exe`:               `c:\users\con\appdata\local\zoom\`,
		`C:\Users\Con\AppData\Local\Programs\Scratch 3\Scratch.exe`:  `c:\users\con\appdata\local\programs\scratch 3\`,
		`C:\Users\Con\AppData\Local\Microsoft\OneDrive\OneDrive.exe`: `c:\users\con\appdata\local\microsoft\onedrive\`,
		`C:\Users\Con\AppData\Local\Temp\x\setup.exe`:                ``,
		`C:\Users\Con\AppData\Local\game.exe`:                        ``,
		`C:\ProgramData\Vendor\tool.exe`:                              `c:\programdata\vendor\`,
		`C:\Users\Con\Downloads\setup.exe`:                           ``,
	}
	for p, muon := range ca {
		if got := thuMucApp(p); got != muon {
			t.Errorf("%s => %q, muon %q", p, got, muon)
		}
	}
}
