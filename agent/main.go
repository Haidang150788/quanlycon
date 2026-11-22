// Phan mem chay nen tren may Windows cua con. Moi 30 giay hoi server (Cloudflare Worker o ../worker),
// moi 2 giay tu tinh "co chan khong" theo gio server roi ap dung:
//   - file hosts + chinh sach Chrome/Edge (URLBlocklist, tat DNS-over-HTTPS, tat an danh)
//   - dong cac tien trinh game/app bi chan
//   - bao truoc 5 & 1 phut, dong trinh duyet khi het gio giai lao
//   - khoa man hinh khi Sep bam tren dien thoai
//
// Lenh (chay bang quyen Administrator):
//
//	quanlycon.exe cai <AGENT_TOKEN> [server]   cai dich vu chay nen
//	quanlycon.exe go                            go dich vu + tra hosts/chinh sach ve nhu cu
//	quanlycon.exe thu                           chay o cua so (go loi), Ctrl+C de dung
//	quanlycon.exe xem                           in trang thai dang ap dung
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const PhienBan = "1.4.0"
// Gan luc build (scripts/build-exe.sh): -X main.ServerMacDinh=https://... -X main.CauHoiBo=...
// Bo trong thi luc cai se hoi dia chi server.
var ServerMacDinh = ""

// Hien cho con moi khi mo game/app bi chan (toi da 1 lan/5 phut moi app).
var CauHoiBo = "Hỏi bố mẹ nếu muốn chơi."

// Trinh duyet bi dong khi het gio giai lao (neu Sep bat "tu dong trinh duyet").
var trinhDuyet = []string{"chrome.exe", "msedge.exe", "firefox.exe", "opera.exe", "brave.exe", "browser.exe", "coccoc.exe"}

type CauHinhMay struct {
	Server string `json:"server"`
	Token  string `json:"token"`
	ID     string `json:"id"`
}

type SuKien struct {
	Ts      int64  `json:"ts"`
	Loai    string `json:"loai"`
	NoiDung string `json:"noi_dung"`
}

// Luu xuong dia de mat mang / khoi dong lai van chan theo lan cuoi.
type BoNho struct {
	Ban    *Ban  `json:"ban"`
	DaKhoa int64 `json:"da_khoa"`
	DaChup int64 `json:"da_chup"`
}

type TacNhan struct {
	ch      CauHinhMay
	mu      sync.Mutex
	ban     *Ban
	gioSv   int64     // gio server lan hoi cuoi (ms)
	gocDH   time.Time // thoi diem may nhan gio do (monotonic, doi gio Windows khong anh huong)
	coGioSv bool
	daKhoa  int64
	hang    []SuKien
	ghiChu  string

	daApDung   string // khoa cua bo hosts/chinh sach dang ap
	chanTruoc  bool
	daBao      map[int]bool
	lanBaoDong map[string]time.Time

	// Ban 2 + 3
	kc       *KhoaCai
	xinGio   trangXin
	lanHoiOK time.Time
	daQD     map[string]bool
	dangHoi  int32
	daThuNen bool

	// Ban 4
	gs     *GiamSat
	daChup int64

	// Ban 5
	extLuc map[string]time.Time // "chrome"|"edge" => lan cuoi extension bao ve
	daExt  string
}

type trangXin struct {
	ID        string
	TrangThai string
	Phut      int
	Luc       int64
}

type ketQuaXin struct {
	ID        string `json:"id"`
	TrangThai string `json:"trang_thai"`
	GhiChu    string `json:"ghi_chu"`
	Loi       string `json:"loi"`
}

// Ly do dich vu dung: "tay" (ai do bam Stop) | "tat_may". Dat boi may_windows.go truoc khi dong kenh dung.
var lyDoDung atomic.Value

func thuMuc() string { return thuMucDuLieu() }

func (t *TacNhan) gioHienTai() int64 {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.coGioSv {
		return t.gioSv + time.Since(t.gocDH).Milliseconds()
	}
	return time.Now().UnixMilli()
}

func (t *TacNhan) suKien(loai, noiDung string) {
	log.Printf("[%s] %s", loai, noiDung)
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.hang) < 200 {
		t.hang = append(t.hang, SuKien{Ts: time.Now().UnixMilli(), Loai: loai, NoiDung: noiDung})
	}
}

func docJSON(p string, v any) error {
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

func ghiJSON(p string, v any) error {
	b, _ := json.MarshalIndent(v, "", "  ")
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func (t *TacNhan) luuBoNho() {
	t.mu.Lock()
	bn := BoNho{Ban: t.ban, DaKhoa: t.daKhoa, DaChup: t.daChup}
	t.mu.Unlock()
	if err := ghiJSON(filepath.Join(thuMuc(), "bonho.json"), bn); err != nil {
		log.Printf("khong luu duoc bonho.json: %v", err)
	}
}

var http15 = &http.Client{Timeout: 15 * time.Second}

func (t *TacNhan) hoiServer() error {
	t.mu.Lock()
	gui := t.hang
	t.hang = nil
	ghiChu := t.ghiChu
	t.mu.Unlock()
	ten, _ := os.Hostname()
	var hd []HoatDong
	if t.gs != nil {
		hd = t.gs.layHang()
	}
	body, _ := json.Marshal(map[string]any{"id": t.ch.ID, "ten": ten, "phien_ban": PhienBan, "su_kien": gui, "ghi_chu": ghiChu, "hoat_dong": hd})
	req, _ := http.NewRequest("POST", strings.TrimRight(t.ch.Server, "/")+"/api/may/hoi", bytes.NewReader(body))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "Bearer "+t.ch.Token)
	gocGui := time.Now()
	res, err := http15.Do(req)
	traLai := func() { // gui hong => giu su kien de lan sau gui lai
		t.mu.Lock()
		t.hang = append(gui, t.hang...)
		t.mu.Unlock()
		if t.gs != nil {
			t.gs.traHang(hd)
		}
	}
	if err != nil {
		traLai()
		return err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode != 200 {
		traLai()
		return fmt.Errorf("server tra %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	var b Ban
	if err := json.Unmarshal(raw, &b); err != nil || b.Now == 0 {
		traLai()
		return fmt.Errorf("server tra sai khuon: %v", err)
	}
	// Gio server o giua luc gui va luc nhan.
	tre := time.Since(gocGui) / 2
	t.mu.Lock()
	t.ban = &b
	t.gioSv = b.Now
	t.gocDH = gocGui.Add(tre)
	t.coGioSv = true
	t.lanHoiOK = time.Now()
	t.mu.Unlock()
	t.luuBoNho()
	t.xuLyQuyetDinh(b.QuyetDinh)
	return nil
}

// Gui yeu cau len server (xin them gio / xin chay phan mem la).
func (t *TacNhan) xinServer(body map[string]any) (ketQuaXin, error) {
	body["may"] = t.ch.ID
	raw, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", strings.TrimRight(t.ch.Server, "/")+"/api/may/xin", bytes.NewReader(raw))
	req.Header.Set("content-type", "application/json")
	req.Header.Set("authorization", "Bearer "+t.ch.Token)
	c := &http.Client{Timeout: 10 * time.Second}
	res, err := c.Do(req)
	if err != nil {
		return ketQuaXin{}, fmt.Errorf("không kết nối được: %w", err)
	}
	defer res.Body.Close()
	var kq ketQuaXin
	_ = json.NewDecoder(io.LimitReader(res.Body, 1<<16)).Decode(&kq)
	if res.StatusCode != 200 {
		if kq.Loi == "" {
			kq.Loi = fmt.Sprintf("server trả %d", res.StatusCode)
		}
		return kq, fmt.Errorf("%s", kq.Loi)
	}
	if kq.ID == "" {
		return kq, fmt.Errorf("server không trả mã yêu cầu")
	}
	return kq, nil
}

func (t *TacNhan) xuLyQuyetDinh(ds []QuyetDinh) {
	for _, qd := range ds {
		t.mu.Lock()
		if t.daQD == nil || len(t.daQD) > 500 {
			t.daQD = map[string]bool{}
		}
		thay := t.daQD[qd.ID]
		t.daQD[qd.ID] = true
		laXinCuaCon := qd.Loai == "gio" && qd.ID == t.xinGio.ID
		if laXinCuaCon {
			t.xinGio.TrangThai = qd.TrangThai
		}
		t.mu.Unlock()
		if thay {
			continue
		}
		switch {
		case qd.Loai == "chay" && t.kc != nil:
			t.kc.XuLyQuyetDinh(qd)
		case laXinCuaCon && qd.TrangThai == "dong_y":
			baoNguoiDung("Bố mẹ đã đồng ý", fmt.Sprintf("Con được thêm %d phút. Chơi vui nhé!", qd.Phut))
		case laXinCuaCon && qd.TrangThai == "tu_choi":
			baoNguoiDung("Bố mẹ chưa đồng ý", "Lần này bố mẹ chưa cho thêm giờ.")
		}
	}
}

func coKhong(b bool) string {
	if b {
		return "xem theo kênh"
	}
	return "chặn hẳn"
}

// Hien tren dien thoai canh ten may: extension YouTube co dang chay khong.
func ghiChuExt(b *Ban, chrome, edge bool) string {
	if b == nil || (b.ExtChrome == "" && b.ExtEdge == "") {
		return ""
	}
	p := []string{}
	if b.ExtChrome != "" {
		p = append(p, "Chrome "+map[bool]string{true: "✓", false: "✗"}[chrome])
	}
	if b.ExtEdge != "" {
		p = append(p, "Edge "+map[bool]string{true: "✓", false: "✗"}[edge])
	}
	return "Extension YouTube: " + strings.Join(p, ", ")
}

func khoaDanhSach(a, b []string) string {
	x := append([]string{}, a...)
	y := append([]string{}, b...)
	sort.Strings(x)
	sort.Strings(y)
	return strings.Join(x, ",") + "|" + strings.Join(y, ",")
}

func (t *TacNhan) nhip() {
	now := t.gioHienTai()
	t.mu.Lock()
	b := t.ban
	t.mu.Unlock()
	if b == nil {
		return // chua tung nhan cau hinh => khong lam gi (khong doan)
	}
	biChan := DangChan(b, now)

	extChrome, extEdge := t.extDangSong()
	d := TinhDanhSachChan(b, biChan, extChrome, extEdge)
	mien := d.Mien
	k := khoaDanhSach(mien, d.UrlChrome) + "|" + khoaDanhSach(nil, d.UrlEdge)
	if k != t.daApDung {
		if err := apDungChan(d); err != nil {
			t.suKien("loi", "Không áp dụng được danh sách chặn: "+err.Error())
		} else {
			t.daApDung = k
			if biChan {
				ghi := ""
				if len(b.KenhCho) > 0 && len(b.MienTheoKenh) > 0 {
					ghi = fmt.Sprintf(" · YouTube theo %d kênh: Chrome %s, Edge %s", len(b.KenhCho), coKhong(len(d.UrlChrome) < len(b.UrlChan)), coKhong(len(d.UrlEdge) < len(b.UrlChan)))
				}
				t.suKien("ap_dung", fmt.Sprintf("Bắt đầu chặn %d tên miền, %d ứng dụng%s", len(mien), len(b.TienTrinh), ghi))
			} else {
				t.suKien("ap_dung", "Mở chặn (giờ giải lao / cho chơi tạm / tắt chặn)")
			}
		}
	}

	// Ep cai extension YouTube theo kenh khi server co ma extension (doi ma => ghi lai).
	if ke := b.ExtChrome + "|" + b.ExtEdge; ke != t.daExt {
		if err := datExtension(dongForcelist(b.ExtChrome, khoChrome), dongForcelist(b.ExtEdge, khoEdge)); err != nil {
			t.suKien("loi", "Không đặt được chính sách cài extension: "+err.Error())
		} else {
			t.daExt = ke
		}
	}
	t.mu.Lock()
	t.ghiChu = ghiChuExt(b, extChrome, extEdge)
	t.mu.Unlock()

	if biChan {
		ds := map[string]bool{}
		for _, e := range b.TienTrinh {
			ds[strings.ToLower(e)] = true
		}
		if d.DongTrinhDuyetKhac {
			for _, e := range trinhDuyetKhongExt {
				ds[e] = true
			}
		}
		if !t.chanTruoc && b.DongTrinhDuyet {
			for _, e := range trinhDuyet {
				ds[e] = true
			}
		}
		for _, ten := range dongTienTrinh(ds) {
			if lan, ok := t.lanBaoDong[ten]; !ok || time.Since(lan) > 5*time.Minute {
				t.lanBaoDong[ten] = time.Now()
				t.suKien("dong", "Đã đóng "+ten)
				if !laTrinhDuyet(ten) {
					baoNguoiDung("Đang bị chặn", ten+" đang bị chặn trong giờ này nên đã bị đóng.\n\n"+CauHoiBo)
				}
			}
			if t.gs != nil && !laTrinhDuyet(ten) {
				go t.gs.Chup("Con mở "+ten+" (đang bị chặn)", true)
			}
		}
		t.daBao = map[int]bool{}
	} else if b.DongTrinhDuyet {
		con := ConBaoLauThiChan(b, now)
		for _, moc := range []int{5, 1} {
			if con > 0 && con <= int64(moc)*60000 && !t.daBao[moc] {
				t.daBao[moc] = true
				baoNguoiDung("Sắp hết giờ giải lao", fmt.Sprintf("Còn %d phút nữa là hết giờ giải lao. Con lưu bài lại nhé — trình duyệt và game sẽ tự đóng.\n\nMuốn xin thêm giờ: bấm biểu tượng \"Xin thêm giờ\" ngoài màn hình.", moc))
			}
		}
	}
	t.chanTruoc = biChan

	// Ban 3: khoa cai phan mem
	if t.kc != nil {
		if b.KhoaCai && !t.daThuNen {
			t.daThuNen = true
			t.lapNen()
		}
		t.kc.Nhip(b, time.Now())
	}

	t.mu.Lock()
	canChup := b.ChupLuc > t.daChup
	if canChup {
		t.daChup = b.ChupLuc
	}
	t.mu.Unlock()
	if canChup {
		t.luuBoNho()
		if now-b.ChupLuc < 5*60000 && t.gs != nil {
			go t.gs.Chup("Bố mẹ bấm chụp ngay", false)
		}
	}

	t.mu.Lock()
	can := b.KhoaLuc > t.daKhoa
	khoaLuc := b.KhoaLuc
	t.mu.Unlock()
	if can {
		t.mu.Lock()
		t.daKhoa = khoaLuc
		t.mu.Unlock()
		t.luuBoNho()
		if now-khoaLuc < 10*60000 { // lenh qua cu (may tat luc Sep bam) thi bo qua
			if err := khoaManHinh(); err != nil {
				t.suKien("loi", "Không khoá được màn hình: "+err.Error())
			} else {
				t.suKien("khoa", "Đã khoá màn hình theo lệnh")
			}
		}
	}
}

func (t *TacNhan) chay(dung <-chan struct{}) {
	t.daBao = map[int]bool{}
	t.lanBaoDong = map[string]time.Time{}
	// Coi nhu "vua chan" luc khoi dong: khoi dong lai dich vu giua gio chan khong duoc dong trinh duyet cua con.
	t.chanTruoc = true
	var bn BoNho
	if docJSON(filepath.Join(thuMuc(), "bonho.json"), &bn) == nil {
		t.ban = bn.Ban
		t.daKhoa = bn.DaKhoa
		t.daChup = bn.DaChup
	}
	if err := datChongLach(); err != nil {
		t.suKien("loi", "Không đặt được chính sách trình duyệt: "+err.Error())
	}
	t.suKien("khoi_dong", "Phần mềm khởi động, phiên bản "+PhienBan)

	if nt := nenTangThat(); nt != nil {
		var nen DanhSachNen
		_ = docJSON(filepath.Join(thuMuc(), "nen.json"), &nen)
		t.daThuNen = len(nen.ThuMuc) > 0
		t.kc = MoiKhoaCai(nt, thuMucHeThong(), nen.ThuMuc, func(ten, duong, sha string, kt int64) (string, error) {
			kq, err := t.xinServer(map[string]any{"loai": "chay", "ten": ten, "duong_dan": duong, "sha256": sha, "kich_thuoc": kt})
			if err == nil {
				go t.hoiNgay() // bat dau hoi nhanh de nhan quyet dinh
				if t.gs != nil {
					go t.gs.Chup("Con mở phần mềm lạ "+ten, true)
				}
			}
			return kq.ID, err
		}, t.suKien)
	}
	t.moTrangCon()
	t.gs = &GiamSat{t: t}

	t.hoiNgay()
	nhip := time.NewTicker(time.Second)
	hoiLai := time.NewTicker(30 * time.Second)
	nhanh := time.NewTicker(3 * time.Second)
	nhinTick := time.NewTicker(30 * time.Second)
	defer nhinTick.Stop()
	defer nhip.Stop()
	defer hoiLai.Stop()
	defer nhanh.Stop()
	for {
		t.nhip()
		select {
		case <-dung:
			ly, _ := lyDoDung.Load().(string)
			if ly == "tay" && !dangCaiLai() {
				t.suKien("dung_tay", "Dịch vụ QuanLyCon bị dừng lúc "+time.Now().Format("15:04 02/01"))
			} else {
				t.suKien("tat", "Phần mềm dừng (tắt máy / cài lại)")
			}
			_ = t.hoiServer() // co gang bao len lan cuoi
			return
		case <-nhip.C:
		case <-hoiLai.C:
			go t.hoiNgay()
		case <-nhinTick.C:
			go t.gs.ghiCuaSo()
		case <-nhanh.C:
			t.mu.Lock()
			cho := t.xinGio.TrangThai == "cho"
			t.mu.Unlock()
			if cho || (t.kc != nil && t.kc.ConCho() > 0) {
				go t.hoiNgay()
			}
		}
	}
}

// Hoi server, khong chong cheo nhau.
func (t *TacNhan) hoiNgay() {
	if !atomic.CompareAndSwapInt32(&t.dangHoi, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&t.dangHoi, 0)
	if err := t.hoiServer(); err != nil {
		log.Printf("hoi server loi: %v", err)
	}
}

type DanhSachNen struct {
	TaoLuc int64    `json:"tao_luc"`
	ThuMuc []string `json:"thu_muc"`
}

// Lan dau bat khoa cai: ghi lai cac thu muc app dang co tren may de khong hoi lai (Zoom, app hoc...).
func (t *TacNhan) lapNen() {
	ds := quetNenApp()
	nen := DanhSachNen{TaoLuc: time.Now().UnixMilli(), ThuMuc: ds}
	if len(ds) == 0 {
		nen.ThuMuc = []string{"(trong)"}
	}
	if err := ghiJSON(filepath.Join(thuMuc(), "nen.json"), nen); err != nil {
		t.suKien("loi", "Không lưu được danh sách app có sẵn: "+err.Error())
	}
	t.kc.DatNen(ds)
	t.suKien("nen", fmt.Sprintf("Bật khoá cài: ghi nhận %d thư mục app có sẵn trên máy, các app này chạy bình thường", len(ds)))
}

// File danh dau "dang cai lai" do lenh `cai` ghi truoc khi dung dich vu => khong bao Sep la bi dung trai phep.
func dangCaiLai() bool {
	st, err := os.Stat(filepath.Join(thuMuc(), "dang_cai"))
	return err == nil && time.Since(st.ModTime()) < 5*time.Minute
}

func moNhatKy() {
	_ = os.MkdirAll(thuMuc(), 0700)
	p := filepath.Join(thuMuc(), "nhatky.log")
	if st, err := os.Stat(p); err == nil && st.Size() > 2<<20 {
		_ = os.Rename(p, p+".1")
	}
	if f, err := os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600); err == nil {
		log.SetOutput(io.MultiWriter(f, os.Stderr))
	}
}

func napCauHinh() (CauHinhMay, error) {
	var ch CauHinhMay
	err := docJSON(filepath.Join(thuMuc(), "may.json"), &ch)
	if err == nil && (ch.Token == "" || ch.Server == "") {
		err = fmt.Errorf("may.json thieu token/server")
	}
	return ch, err
}

func main() {
	if laDichVu() {
		moNhatKy()
		chayDichVu(func(dung <-chan struct{}) {
			ch, err := napCauHinh()
			if err != nil {
				log.Printf("khong doc duoc cau hinh: %v", err)
				<-dung
				return
			}
			(&TacNhan{ch: ch}).chay(dung)
		})
		return
	}
	lenh := ""
	if len(os.Args) > 1 {
		lenh = os.Args[1]
	}
	switch lenh {
	case "cai":
		if len(os.Args) < 3 {
			fmt.Println("Cách dùng: quanlycon.exe cai <AGENT_TOKEN> [server]")
			os.Exit(2)
		}
		server := ServerMacDinh
		if len(os.Args) > 3 {
			server = os.Args[3]
		}
		if server == "" {
			fmt.Println("Cách dùng: quanlycon.exe cai <AGENT_TOKEN> <https://địa-chỉ-server>")
			os.Exit(2)
		}
		ketThuc(caiDat(os.Args[2], server))
	case "go":
		ketThuc(goCaiDat())
	case "thu":
		moNhatKy()
		ch, err := napCauHinh()
		ketThucNeuLoi(err)
		dung := make(chan struct{})
		choCtrlC(dung)
		(&TacNhan{ch: ch}).chay(dung)
	case "nhin":
		// Chay trong phien cua con (do dich vu goi): in JSON + anh ra stdout.
		ketThucNeuLoi(chayNhin(len(os.Args) > 2 && os.Args[2] == "anh"))
	case "xem":
		var bn BoNho
		_ = docJSON(filepath.Join(thuMuc(), "bonho.json"), &bn)
		now := time.Now().UnixMilli()
		fmt.Printf("Phiên bản %s, thư mục %s\n", PhienBan, thuMuc())
		if bn.Ban == nil {
			fmt.Println("Chưa nhận cấu hình nào từ server.")
			return
		}
		fmt.Printf("Theo giờ máy: %s\n", map[bool]string{true: "ĐANG CHẶN", false: "đang mở"}[DangChan(bn.Ban, now)])
		fmt.Printf("%d tên miền, %d ứng dụng, giải lao %v\n", len(bn.Ban.TenMien), len(bn.Ban.TienTrinh), bn.Ban.GiaiLao)
	case "":
		// Bam dup vao file .exe: tu xin quyen Admin roi cai, giu cua so cho doc ket qua.
		caiTay()
	case "caitay":
		caiTayDaAdmin()
	default:
		fmt.Println("Quản lý máy tính của con", PhienBan)
		fmt.Println("  cai <token> [server]   cài dịch vụ (chạy bằng quyền Administrator)")
		fmt.Println("  go                     gỡ dịch vụ, trả máy về như cũ")
		fmt.Println("  thu                    chạy thử trong cửa sổ này")
		fmt.Println("  xem                    xem trạng thái")
	}
}

func laTrinhDuyet(ten string) bool {
	for _, e := range trinhDuyet {
		if e == ten {
			return true
		}
	}
	return false
}

// Gan luc build file cai: -ldflags "-X main.tokenNhung=QLCTOKEN=<AGENT_TOKEN>".
// Ban chep vao Program Files bi xoa doan nay (xoaTokenTrongFile) de con khong doc duoc tu exe.
var tokenNhung string

const dauToken = "QLCTOKEN="

func layTokenNhung() string {
	if strings.HasPrefix(tokenNhung, dauToken) && len(tokenNhung) > len(dauToken) {
		return tokenNhung[len(dauToken):]
	}
	return ""
}

func xoaTokenTrongFile(p string) error {
	if layTokenNhung() == "" {
		return nil
	}
	b, err := os.ReadFile(p)
	if err != nil {
		return err
	}
	k := []byte(tokenNhung)
	if !bytes.Contains(b, k) {
		return nil
	}
	return os.WriteFile(p, bytes.ReplaceAll(b, k, make([]byte, len(k))), 0755)
}

// Cai bang tay (da co quyen Admin): lay token nhung san, khong co thi hoi.
func caiTayDaAdmin() {
	fmt.Println("=== Cài phần mềm Quản lý máy tính của con", PhienBan, "===")
	doc := bufio.NewReader(os.Stdin)
	hoi := func(cau string) string {
		fmt.Print(cau)
		dong, _ := doc.ReadString('\n')
		return strings.TrimSpace(dong)
	}
	server := ServerMacDinh
	if server == "" {
		server = strings.TrimRight(hoi("Địa chỉ server (vd https://quanlycon.ten-ban.workers.dev): "), "/")
		if server != "" && !strings.HasPrefix(server, "http") {
			server = "https://" + server
		}
	}
	token := layTokenNhung()
	if token == "" {
		token = hoi("Dán mã AGENT_TOKEN rồi bấm Enter: ")
	}
	if token == "" || server == "" {
		fmt.Println("LỖI: thiếu địa chỉ server hoặc mã AGENT_TOKEN.")
	} else if err := caiDat(token, server); err != nil {
		fmt.Println("LỖI:", err)
	} else {
		fmt.Println()
		fmt.Println("CÀI XONG. Mở", server, "trên điện thoại để điều khiển.")
	}
	dungCho()
}

func dungCho() {
	fmt.Print("\nBấm Enter để đóng cửa sổ này...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

func ketThucNeuLoi(err error) {
	if err != nil {
		fmt.Println("LỖI:", err)
		os.Exit(1)
	}
}

func ketThuc(err error) {
	ketThucNeuLoi(err)
	fmt.Println("Xong.")
}
