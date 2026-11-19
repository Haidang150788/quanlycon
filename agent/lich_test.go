package main

import "testing"

// Cung bo ca voi worker/tests/luat.test.js.
func luc(h, m int) int64 {
	// 2026-11-17 00:00 gio VN
	const nuaDemVN = int64(1794848400000) // Date.UTC(2026,10,17) - 7h
	return nuaDemVN + int64(h*60+m)*60000
}

func TestGioVN(t *testing.T) {
	if phutTrongNgayVN(luc(0, 5)) != 5 || phutTrongNgayVN(luc(23, 59)) != 23*60+59 {
		t.Fatal("lech mui gio")
	}
}

func TestGiaiLao(t *testing.T) {
	b := &Ban{Bat: true, GiaiLao: []Khung{{"11:30", "12:59"}}}
	ca := []struct {
		h, m int
		bi   bool
	}{{11, 29, true}, {11, 30, false}, {12, 59, false}, {13, 0, true}}
	for _, c := range ca {
		if DangChan(b, luc(c.h, c.m)) != c.bi {
			t.Errorf("%02d:%02d muon chan=%v", c.h, c.m, c.bi)
		}
	}
}

func TestQuaDem(t *testing.T) {
	b := &Ban{Bat: true, GiaiLao: []Khung{{"21:00", "01:00"}}}
	if DangChan(b, luc(23, 0)) || DangChan(b, luc(0, 30)) || !DangChan(b, luc(1, 1)) || !DangChan(b, luc(20, 59)) {
		t.Fatal("khung qua dem sai")
	}
}

func TestMoTamVaTat(t *testing.T) {
	now := luc(15, 0)
	if DangChan(&Ban{Bat: true, MoTamDen: now + 60000}, now) {
		t.Error("dang mo tam ma van chan")
	}
	if !DangChan(&Ban{Bat: true, MoTamDen: now - 1}, now) {
		t.Error("het mo tam ma khong chan")
	}
	if DangChan(&Ban{Bat: false}, now) {
		t.Error("tat cong tac ma van chan")
	}
}

func TestConBaoLau(t *testing.T) {
	b := &Ban{Bat: true, GiaiLao: []Khung{{"11:30", "12:59"}}}
	if got := ConBaoLauThiChan(b, luc(12, 50)); got != 10*60000 {
		t.Errorf("12:50 -> chan luc 13:00, con 10 phut; ra %d", got)
	}
	now := luc(15, 0) + 30000
	b2 := &Ban{Bat: true, MoTamDen: now + 5*60000}
	if got := ConBaoLauThiChan(b2, now); got != 5*60000 {
		t.Errorf("mo tam 5 phut; ra %d", got)
	}
	if ConBaoLauThiChan(&Ban{Bat: false}, now) != -1 {
		t.Error("tat cong tac phai -1")
	}
	if ConBaoLauThiChan(b, luc(9, 0)) != 0 {
		t.Error("dang chan phai 0")
	}
}
