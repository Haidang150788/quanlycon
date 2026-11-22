package main

// Ban 2 — trang "Xin them gio" chay ngay tren may con: http://127.0.0.1:47123
// Dich vu (SYSTEM) giu khoa may nen trang nay khong can dang nhap; chi nghe 127.0.0.1 va kiem Host
// de trang web la khong goi vao duoc (chong DNS rebinding).
// Loi tat "Xin thêm giờ" ngoai man hinh (Public Desktop) tro vao day — tao luc `cai`.

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const DiaChiTrangCon = "127.0.0.1:47123"

func (t *TacNhan) moTrangCon() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", t.trangChu)
	mux.HandleFunc("/trang-thai", t.apiTrangThaiCon)
	mux.HandleFunc("/xin", t.apiXinGio)
	mux.HandleFunc("/ext/trang-thai", t.apiExt)
	srv := &http.Server{Handler: chiNoiBo(mux), ReadTimeout: 10 * time.Second, WriteTimeout: 20 * time.Second}
	ln, err := net.Listen("tcp", DiaChiTrangCon)
	if err != nil {
		t.suKien("loi", "Không mở được trang Xin thêm giờ: "+err.Error())
		return
	}
	go func() { log.Print(srv.Serve(ln)) }()
}

func chiNoiBo(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := strings.ToLower(r.Host)
		if host != DiaChiTrangCon && host != "localhost:47123" {
			http.Error(w, "forbidden", 403)
			return
		}
		if r.Method == "POST" && r.URL.Path != "/ext/trang-thai" { // /ext/: apiExt tu kiem Origin extension
			if o := r.Header.Get("Origin"); o != "" && o != "http://"+host {
				http.Error(w, "forbidden", 403)
				return
			}
		}
		w.Header().Set("cache-control", "no-store")
		h.ServeHTTP(w, r)
	})
}

type ttCon struct {
	DangChan bool   `json:"dang_chan"`
	TatChan  bool   `json:"tat_chan"`
	ConMs    int64  `json:"con_ms"` // >0: con bao lau thi chan lai
	MoTamDen int64  `json:"mo_tam_den"`
	Now      int64  `json:"now"`
	XinID    string `json:"xin_id"`
	XinTT    string `json:"xin_tt"` // cho | dong_y | tu_choi | het_han
	XinPhut  int    `json:"xin_phut"`
	XinLuc   int64  `json:"xin_luc"`
	ChoDuyet int    `json:"cho_duyet"`
	CoMang   bool   `json:"co_mang"`
}

func (t *TacNhan) trangThaiCon() ttCon {
	now := t.gioHienTai()
	t.mu.Lock()
	b := t.ban
	x := t.xinGio
	coMang := time.Since(t.lanHoiOK) < 2*time.Minute
	t.mu.Unlock()
	tt := ttCon{Now: now, XinID: x.ID, XinTT: x.TrangThai, XinPhut: x.Phut, XinLuc: x.Luc, CoMang: coMang}
	if b != nil {
		tt.DangChan = DangChan(b, now)
		tt.TatChan = !b.Bat
		tt.ConMs = ConBaoLauThiChan(b, now)
		tt.MoTamDen = b.MoTamDen
	}
	if t.kc != nil {
		tt.ChoDuyet = t.kc.ConCho()
	}
	return tt
}

func (t *TacNhan) apiTrangThaiCon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(t.trangThaiCon())
}

func (t *TacNhan) apiXinGio(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	if r.Method != "POST" {
		http.Error(w, `{"loi":"POST"}`, 405)
		return
	}
	var b struct {
		Phut int    `json:"phut"`
		LyDo string `json:"ly_do"`
	}
	_ = json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&b)
	kq, err := t.xinServer(map[string]any{"loai": "gio", "phut": b.Phut, "ly_do": b.LyDo})
	if err != nil {
		w.WriteHeader(400)
		_ = json.NewEncoder(w).Encode(map[string]string{"loi": err.Error()})
		return
	}
	t.mu.Lock()
	moi := kq.ID != t.xinGio.ID
	if moi {
		t.xinGio = trangXin{ID: kq.ID, TrangThai: "cho", Phut: b.Phut, Luc: time.Now().UnixMilli()}
	}
	t.mu.Unlock()
	if moi {
		go t.hoiNgay() // bat dau hoi nhanh de nhan cau tra loi
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "ghi_chu": kq.GhiChu})
}

func (t *TacNhan) trangChu(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("content-type", "text/html; charset=utf-8")
	fmt.Fprint(w, htmlTrangCon)
}

const htmlTrangCon = `<!doctype html><html lang="vi"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1"><title>Xin thêm giờ</title>
<style>
:root{--nen:#f3f5f8;--the:#fff;--chu:#13202c;--mo:#66727e;--vien:#e3e7ec;--cam:#f28c1b;--xanh:#1f9d55;--do:#d63b3b}
@media (prefers-color-scheme:dark){:root{--nen:#0c141c;--the:#15202b;--chu:#e8edf2;--mo:#93a1ae;--vien:#26323e}}
*{box-sizing:border-box}body{margin:0;background:var(--nen);color:var(--chu);font:16px/1.5 "Segoe UI",system-ui,sans-serif}
main{max-width:520px;margin:40px auto;padding:0 16px}
.the{background:var(--the);border:1px solid var(--vien);border-radius:16px;padding:20px;margin-bottom:14px}
h1{font-size:24px;margin:0 0 6px}.mo{color:var(--mo);font-size:14px}
.lon{font-size:20px;font-weight:700;margin:4px 0}
.chon{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;margin:14px 0}
.chon button{border:2px solid var(--vien);background:var(--nen);color:var(--chu);border-radius:12px;padding:14px 0;font:inherit;font-weight:700;cursor:pointer}
.chon button.dang{border-color:var(--cam);background:#fff4e6;color:#a85a00}
textarea{width:100%;min-height:70px;font:inherit;padding:10px;border-radius:10px;border:1px solid var(--vien);background:var(--nen);color:var(--chu)}
.gui{width:100%;margin-top:12px;border:0;border-radius:12px;padding:14px;font:inherit;font-weight:700;background:var(--cam);color:#fff;cursor:pointer}
.gui:disabled{opacity:.5;cursor:default}
.kq{margin-top:12px;font-weight:600;min-height:24px}
</style></head><body><main>
<div class="the"><h1>Máy của con</h1><div class="lon" id="tt">…</div><div class="mo" id="phu"></div></div>
<div class="the"><div class="lon">Xin bố mẹ thêm giờ</div>
<div class="mo">Bố mẹ sẽ nhận thông báo trên điện thoại.</div>
<div class="chon"><button data-p="15">15 phút</button><button data-p="30" class="dang">30 phút</button><button data-p="60">1 tiếng</button></div>
<textarea id="lydo" maxlength="200" placeholder="Con xin để làm gì? (vd: làm nốt bài tập nhóm trên Zalo)"></textarea>
<button class="gui" id="gui">Gửi bố mẹ</button><div class="kq" id="kq"></div></div>
</main><script>
let phut=30;const $=id=>document.getElementById(id);
document.querySelectorAll('[data-p]').forEach(b=>b.onclick=()=>{phut=+b.dataset.p;document.querySelectorAll('[data-p]').forEach(x=>x.classList.toggle('dang',x===b))});
const gio=ms=>new Date(ms).toLocaleTimeString('vi-VN',{hour:'2-digit',minute:'2-digit',timeZone:'Asia/Ho_Chi_Minh'});
const p2=ms=>{const m=Math.ceil(ms/60000);return m>=60?Math.floor(m/60)+' giờ '+(m%60?m%60+' phút':''):m+' phút'};
async function tai(){try{const s=await (await fetch('/trang-thai')).json();
 if(s.tat_chan){$('tt').textContent='Đang không chặn gì';$('phu').textContent='';}
 else if(s.dang_chan){$('tt').textContent='🔒 Đang giờ học';$('phu').textContent='Mạng xã hội và game đang bị chặn.';}
 else{$('tt').textContent='🎮 Đang được giải lao';$('phu').textContent=s.con_ms>0?'Còn '+p2(s.con_ms)+' (đến '+gio(s.now+s.con_ms)+')':'';}
 const k=$('kq');
 if(s.xin_tt==='cho'){k.style.color='var(--cam)';k.textContent='⏳ Đã gửi xin '+s.xin_phut+' phút, đang chờ bố mẹ trả lời…';$('gui').disabled=true;}
 else{$('gui').disabled=false;
  if(s.xin_tt==='dong_y'){k.style.color='var(--xanh)';k.textContent='✅ Bố mẹ đã đồng ý!';}
  else if(s.xin_tt==='tu_choi'){k.style.color='var(--do)';k.textContent='Bố mẹ chưa đồng ý lần này.';}
  else if(s.xin_tt==='het_han'){k.style.color='var(--mo)';k.textContent='Bố mẹ chưa xem kịp, con nhắn bố mẹ nhé.';}
  else k.textContent='';}
 if(!s.co_mang){k.style.color='var(--do)';k.textContent='Máy đang mất mạng, chưa gửi được cho bố mẹ.';}
}catch(e){$('tt').textContent='Phần mềm chưa chạy';}}
$('gui').onclick=async()=>{$('gui').disabled=true;$('kq').textContent='Đang gửi…';
 try{const r=await fetch('/xin',{method:'POST',headers:{'content-type':'application/json'},body:JSON.stringify({phut,ly_do:$('lydo').value})});
 const j=await r.json();if(!r.ok)throw new Error(j.loi||'Lỗi');if(j.ghi_chu)$('kq').textContent=j.ghi_chu;}catch(e){$('kq').style.color='var(--do)';$('kq').textContent=e.message;$('gui').disabled=false;return}
 tai();};
tai();setInterval(tai,4000);
</script></body></html>`
