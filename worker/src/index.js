// Worker quanlycon — server dieu khien may tinh cua con.
// Hai loai khach:
//   1) Dien thoai cua Sep: /api/* dang nhap bang PIN (cookie).
//   2) May con (agent Windows): /api/may/*, xac thuc bang Bearer AGENT_TOKEN.
// Moi quyet dinh "luc nay co chan khong" agent tu tinh theo GIO SERVER gui kem (khong tin dong ho may con).
// Cung mot luat duoc viet lai o agent/lich.go — sua mot ben phai sua ben kia (co test o ca hai).
import { DANH_MUC, NHOM, MA_HOP_LE } from './danhmuc.js';
import { dangChan, chuanHoaCauHinh, quyTacTuDuongDan } from './luat.js';
import { guiPush, khoaCongKhai } from './push.js';

const json = (data, status = 200, headers = {}) =>
  new Response(JSON.stringify(data), {
    status,
    headers: { 'content-type': 'application/json; charset=utf-8', 'cache-control': 'no-store', ...headers },
  });
const loi = (msg, status = 400) => json({ loi: msg }, status);
const docBody = (req) => req.json().catch(() => ({}));

// Phan mem la cho duyet toi da 2 phut, agent tu huy sau moc nay (agent/khoacai.go CHO_DUYET).
const HAN_CHAY_MS = 2 * 60e3;
// Xin them gio: con chua tra loi sau 30 phut thi coi nhu het han.
const HAN_GIO_MS = 30 * 60e3;
const TOI_DA_XIN_GIO_NGAY = 10;

const MAC_DINH = {
  bat: true,
  nen_tang: [...MA_HOP_LE],
  web_khac: [],
  giai_lao: [],
  dong_trinh_duyet: true,
  khoa_cai: true,
  chup_phut: 10, // chup man hinh dinh ky (phut), 0 = tat
  chup_luc: 0, // Sep bam "Chup ngay"
  mo_tam_den: 0,
  khoa_luc: 0,
};

async function docCauHinh(env) {
  const r = await env.DB.prepare("SELECT v FROM cai_dat WHERE k='cau_hinh'").first();
  if (!r) return { ...MAC_DINH };
  try { return chuanHoaCauHinh({ ...MAC_DINH, ...JSON.parse(r.v) }, MA_HOP_LE).cau_hinh; } catch { return { ...MAC_DINH }; }
}
async function ghiCauHinh(env, ch) {
  await env.DB.prepare("INSERT INTO cai_dat(k,v) VALUES('cau_hinh',?) ON CONFLICT(k) DO UPDATE SET v=excluded.v")
    .bind(JSON.stringify(ch)).run();
}
async function ghiNhatKy(env, may, loai, noiDung, ts = Date.now()) {
  await env.DB.prepare('INSERT INTO nhat_ky(ts,may,loai,noi_dung) VALUES(?,?,?,?)').bind(ts, may, loai, noiDung).run();
}

// Gui thong bao nhung khong de loi push lam hong yeu cau chinh.
function baoSep(env, ctx, tin) {
  const p = guiPush(env, tin).catch((e) => console.log('push loi', e));
  if (ctx && ctx.waitUntil) ctx.waitUntil(p);
}

function bang(a, b) {
  if (typeof a !== 'string' || typeof b !== 'string' || a.length !== b.length) return false;
  let x = 0;
  for (let i = 0; i < a.length; i++) x |= a.charCodeAt(i) ^ b.charCodeAt(i);
  return x === 0;
}

// ---------- PIN dien thoai ----------
async function chuKy(pin) {
  const key = await crypto.subtle.importKey('raw', new TextEncoder().encode(pin), { name: 'HMAC', hash: 'SHA-256' }, false, ['sign']);
  const sig = await crypto.subtle.sign('HMAC', key, new TextEncoder().encode('quanlycon-ok-v1'));
  return [...new Uint8Array(sig)].map((b) => b.toString(16).padStart(2, '0')).join('');
}
function docCookie(req, ten) {
  for (const phan of (req.headers.get('cookie') || '').split(';')) {
    const [k, ...v] = phan.trim().split('=');
    if (k === ten) return decodeURIComponent(v.join('='));
  }
  return null;
}
async function duocVao(req, env) {
  if (!env.PIN) return false; // chua dat PIN => dong, khong bao gio mo tu do
  return bang(docCookie(req, 'qlc_ok') || '', await chuKy(env.PIN));
}

// Chan doan PIN: 5 lan sai / 15 phut => khoa 15 phut.
async function dangNhap(req, env) {
  if (!env.PIN) return loi('Chưa đặt PIN trên server', 503);
  const now = Date.now();
  const r = await env.DB.prepare("SELECT v FROM cai_dat WHERE k='sai_pin'").first();
  const sai = r ? JSON.parse(r.v).filter((t) => now - t < 15 * 60e3) : [];
  if (sai.length >= 5) return loi('Sai quá 5 lần, thử lại sau 15 phút', 429);
  const { pin } = await docBody(req);
  if (!bang(String(pin || ''), env.PIN)) {
    sai.push(now);
    await env.DB.prepare("INSERT INTO cai_dat(k,v) VALUES('sai_pin',?) ON CONFLICT(k) DO UPDATE SET v=excluded.v").bind(JSON.stringify(sai)).run();
    await ghiNhatKy(env, null, 'pin_sai', `Đăng nhập sai PIN (${sai.length}/5)`);
    return loi('Sai PIN', 401);
  }
  const cookie = `qlc_ok=${await chuKy(env.PIN)}; Path=/; HttpOnly; Secure; SameSite=Strict; Max-Age=${60 * 60 * 24 * 180}`;
  return json({ ok: true }, 200, { 'set-cookie': cookie });
}

// ---------- Yeu cau ----------
// Danh dau het han lazily (khong can cron).
async function donHetHan(env, now) {
  await env.DB.batch([
    env.DB.prepare("UPDATE yeu_cau SET trang_thai='het_han', xu_ly_luc=? WHERE trang_thai='cho' AND loai='chay' AND ts < ?").bind(now, now - HAN_CHAY_MS),
    env.DB.prepare("UPDATE yeu_cau SET trang_thai='het_han', xu_ly_luc=? WHERE trang_thai='cho' AND loai='gio' AND ts < ?").bind(now, now - HAN_GIO_MS),
  ]);
}

async function docQuyTac(env) {
  const { results } = await env.DB.prepare('SELECT id,ts,loai,gia_tri,ten,cho_phep FROM quy_tac ORDER BY id DESC').all();
  return results;
}

// ---------- Ban gui xuong may con ----------
async function banChoMay(env, ch, now) {
  const mien = new Set();
  const url = new Set();
  const exe = new Set();
  for (const n of DANH_MUC) {
    if (!ch.nen_tang.includes(n.ma)) continue;
    n.mien.forEach((m) => { mien.add(m); url.add(m); });
    n.exe.forEach((e) => exe.add(e.toLowerCase()));
  }
  for (const w of ch.web_khac) {
    url.add(w);
    const host = w.split('/')[0];
    if (!w.includes('/')) { mien.add(host); if (!host.startsWith('www.')) mien.add('www.' + host); }
  }
  const qt = await docQuyTac(env);
  const { results: qd } = await env.DB.prepare(
    "SELECT id,loai,trang_thai,xu_ly_phut AS phut FROM yeu_cau WHERE trang_thai!='cho' AND xu_ly_luc > ?").bind(now - 15 * 60e3).all();
  return {
    now,
    bat: ch.bat,
    ten_mien: [...mien].sort(),
    url_chan: [...url].sort(),
    tien_trinh: [...exe].sort(),
    giai_lao: ch.giai_lao,
    dong_trinh_duyet: ch.dong_trinh_duyet,
    mo_tam_den: ch.mo_tam_den,
    khoa_luc: ch.khoa_luc,
    khoa_cai: ch.khoa_cai,
    chup_phut: ch.chup_phut,
    chup_luc: ch.chup_luc,
    cho_thu_muc: qt.filter((q) => q.cho_phep && q.loai === 'thu_muc').map((q) => q.gia_tri),
    cho_hash: qt.filter((q) => q.cho_phep && q.loai === 'hash').map((q) => q.gia_tri),
    cam_hash: qt.filter((q) => !q.cho_phep && q.loai === 'hash').map((q) => q.gia_tri),
    quyet_dinh: qd,
  };
}

function laMay(req, env) {
  const auth = req.headers.get('authorization') || '';
  return !!env.AGENT_TOKEN && bang(auth, 'Bearer ' + env.AGENT_TOKEN);
}

async function mayHoi(req, env, ctx) {
  const body = await docBody(req);
  const id = String(body.id || 'may-con').slice(0, 64);
  const now = Date.now();
  await donHetHan(env, now);
  const ch = await docCauHinh(env);
  const chanNay = dangChan(ch, now);

  const lenh = [
    env.DB.prepare(`INSERT INTO may(id,ten,phien_ban,lan_cuoi,dang_chan,ghi_chu) VALUES(?,?,?,?,?,?)
      ON CONFLICT(id) DO UPDATE SET ten=excluded.ten, phien_ban=excluded.phien_ban, lan_cuoi=excluded.lan_cuoi,
      dang_chan=excluded.dang_chan, ghi_chu=excluded.ghi_chu`)
      .bind(id, String(body.ten || '').slice(0, 64), String(body.phien_ban || '').slice(0, 32), now, chanNay ? 1 : 0,
        String(body.ghi_chu || '').slice(0, 300)),
  ];
  const suKien = Array.isArray(body.su_kien) ? body.su_kien.slice(0, 50) : [];
  for (const s of suKien) {
    const ts = Number(s.ts);
    const loai = String(s.loai || 'khac').slice(0, 32);
    const nd = String(s.noi_dung || '').slice(0, 300);
    lenh.push(env.DB.prepare('INSERT INTO nhat_ky(ts,may,loai,noi_dung) VALUES(?,?,?,?)')
      .bind(Number.isFinite(ts) && Math.abs(ts - now) < 7 * 864e5 ? ts : now, id, loai, nd));
    // Chi bao Sep nhung su kien can Sep lam gi do (xem memory "canh bao phai chi ra viec").
    if (loai === 'dung_tay') baoSep(env, ctx, { tieu_de: '⚠️ Phần mềm trên máy con bị dừng', noi_dung: nd + ' — không phải do tắt máy. Nếu không phải bố mẹ làm thì kiểm tra máy.', the: 'dung' });
  }
  const hd = Array.isArray(body.hoat_dong) ? body.hoat_dong.slice(0, 200) : [];
  for (const h of hd) {
    const ts = Number(h.ts);
    if (!Number.isFinite(ts) || Math.abs(ts - now) > 7 * 864e5) continue;
    lenh.push(env.DB.prepare('INSERT INTO hoat_dong(ts,may,exe,tieu_de,ranh) VALUES(?,?,?,?,?)')
      .bind(ts, id, String(h.exe || '').slice(0, 80), String(h.tieu_de || '').slice(0, 300), Number(h.ranh) || 0));
  }
  if (Math.random() < 0.02) lenh.push(env.DB.prepare('DELETE FROM hoat_dong WHERE ts < ?').bind(now - 30 * 864e5));
  await env.DB.batch(lenh);
  return json(await banChoMay(env, ch, now));
}

// May con gui yeu cau: { loai:'gio', phut, ly_do } | { loai:'chay', ten, duong_dan, sha256, kich_thuoc }
async function mayXin(req, env, ctx) {
  const b = await docBody(req);
  const now = Date.now();
  const may = String(b.may || 'may-con').slice(0, 64);
  await donHetHan(env, now);
  if (b.loai === 'gio') {
    const phut = Number(b.phut);
    if (![15, 30, 60].includes(phut)) return loi('Chỉ xin 15, 30 hoặc 60 phút');
    const dangCho = await env.DB.prepare("SELECT id FROM yeu_cau WHERE loai='gio' AND trang_thai='cho'").first();
    if (dangCho) return json({ id: dangCho.id, trang_thai: 'cho', ghi_chu: 'Đã có một yêu cầu đang chờ bố mẹ trả lời' });
    const dem = await env.DB.prepare("SELECT COUNT(*) n FROM yeu_cau WHERE loai='gio' AND ts > ?").bind(now - 864e5).first();
    if (dem.n >= TOI_DA_XIN_GIO_NGAY) return loi(`Hôm nay đã xin ${dem.n} lần rồi`, 429);
    const id = crypto.randomUUID();
    const lyDo = String(b.ly_do || '').trim().slice(0, 200);
    await env.DB.prepare("INSERT INTO yeu_cau(id,ts,may,loai,phut,ly_do) VALUES(?,?,?,'gio',?,?)").bind(id, now, may, phut, lyDo).run();
    await ghiNhatKy(env, may, 'xin_gio', `Con xin thêm ${phut} phút${lyDo ? ': ' + lyDo : ''}`);
    baoSep(env, ctx, { tieu_de: `Con xin thêm ${phut} phút`, noi_dung: lyDo || '(không ghi lý do)', the: 'gio' });
    return json({ id, trang_thai: 'cho' });
  }
  if (b.loai === 'chay') {
    const sha = String(b.sha256 || '').toLowerCase();
    if (!/^[0-9a-f]{64}$/.test(sha)) return loi('Thiếu sha256');
    const cu = await env.DB.prepare("SELECT id FROM yeu_cau WHERE loai='chay' AND trang_thai='cho' AND sha256=?").bind(sha).first();
    if (cu) return json({ id: cu.id, trang_thai: 'cho' }); // cung file dang cho => khong bao lai
    const id = crypto.randomUUID();
    const ten = String(b.ten || '').slice(0, 120);
    const dd = String(b.duong_dan || '').slice(0, 400);
    const kt = Number(b.kich_thuoc) || 0;
    await env.DB.prepare("INSERT INTO yeu_cau(id,ts,may,loai,ten,duong_dan,sha256,kich_thuoc) VALUES(?,?,?,'chay',?,?,?,?)")
      .bind(id, now, may, ten, dd, sha, kt).run();
    await ghiNhatKy(env, may, 'xin_chay', `Con mở phần mềm lạ ${ten} (${(kt / 1048576).toFixed(1)} MB), đang chờ duyệt`);
    baoSep(env, ctx, { tieu_de: `Con mở phần mềm lạ: ${ten}`, noi_dung: `${dd}\nChờ duyệt tối đa 2 phút, quá hạn tự huỷ.`, the: 'chay-' + id });
    return json({ id, trang_thai: 'cho' });
  }
  return loi('Loại yêu cầu không hợp lệ');
}

// ---------- API dien thoai ----------
async function trangThai(env) {
  const now = Date.now();
  await donHetHan(env, now);
  const ch = await docCauHinh(env);
  const [may, nk, yc, qt, sub] = await Promise.all([
    env.DB.prepare('SELECT * FROM may ORDER BY lan_cuoi DESC').all(),
    env.DB.prepare('SELECT ts,may,loai,noi_dung FROM nhat_ky ORDER BY id DESC LIMIT 80').all(),
    env.DB.prepare("SELECT id,ts,loai,phut,ly_do,ten,duong_dan,kich_thuoc FROM yeu_cau WHERE trang_thai='cho' ORDER BY ts").all(),
    docQuyTac(env),
    env.DB.prepare('SELECT COUNT(*) n FROM push_sub').first(),
  ]);
  return json({
    now, cau_hinh: ch, dang_chan: dangChan(ch, now),
    danh_muc: DANH_MUC.map(({ ma, ten, nhom, mau, chu }) => ({ ma, ten, nhom, mau, chu })), nhom: NHOM,
    may: may.results, nhat_ky: nk.results, yeu_cau: yc.results, quy_tac: qt,
    han_chay_ms: HAN_CHAY_MS,
    push_khoa: env.VAPID_JWK ? khoaCongKhai(JSON.parse(env.VAPID_JWK)) : null,
    so_thiet_bi_bao: sub.n,
  });
}

async function luuCauHinh(req, env) {
  const body = await docBody(req);
  const cu = await docCauHinh(env);
  const { cau_hinh, loi: l } = chuanHoaCauHinh({ ...cu, ...body, mo_tam_den: cu.mo_tam_den, khoa_luc: cu.khoa_luc, chup_luc: cu.chup_luc }, MA_HOP_LE);
  if (l) return loi(l);
  await ghiCauHinh(env, cau_hinh);
  await ghiNhatKy(env, null, 'cau_hinh', `Bố mẹ lưu cài đặt: chặn ${cau_hinh.nen_tang.length} nền tảng, ${cau_hinh.web_khac.length} web khác, ${cau_hinh.giai_lao.length} khung giải lao, khoá cài ${cau_hinh.khoa_cai ? 'BẬT' : 'tắt'}${cau_hinh.bat ? '' : ' — ĐÃ TẮT chặn'}`);
  return trangThai(env);
}

async function moTam(req, env) {
  const { phut } = await docBody(req);
  const p = Number(phut);
  if (!Number.isInteger(p) || p < 0 || p > 240) return loi('Số phút 0–240');
  const ch = await docCauHinh(env);
  ch.mo_tam_den = p === 0 ? 0 : Date.now() + p * 60e3;
  await ghiCauHinh(env, ch);
  await ghiNhatKy(env, null, 'mo_tam', p === 0 ? 'Bố mẹ chặn lại ngay' : `Bố mẹ mở tạm ${p} phút`);
  return trangThai(env);
}

async function khoaNgay(env) {
  const ch = await docCauHinh(env);
  ch.khoa_luc = Date.now();
  await ghiCauHinh(env, ch);
  await ghiNhatKy(env, null, 'khoa', 'Bố mẹ bấm khoá màn hình');
  return trangThai(env);
}

// { quyet: 'dong_y'|'tu_choi', phut?, luon?: bool }
async function xuLyYeuCau(req, env, id) {
  const b = await docBody(req);
  const now = Date.now();
  await donHetHan(env, now);
  const yc = await env.DB.prepare('SELECT * FROM yeu_cau WHERE id=?').bind(id).first();
  if (!yc) return loi('Không thấy yêu cầu', 404);
  if (yc.trang_thai !== 'cho') return loi(yc.trang_thai === 'het_han' ? 'Yêu cầu đã quá hạn' : 'Yêu cầu đã được trả lời', 409);
  const dongY = b.quyet === 'dong_y';
  if (!dongY && b.quyet !== 'tu_choi') return loi('quyet phải là dong_y hoặc tu_choi');
  let phut = null;
  let ghi = '';
  if (yc.loai === 'gio') {
    if (dongY) {
      phut = Number(b.phut) || yc.phut;
      if (!Number.isInteger(phut) || phut < 5 || phut > 240) return loi('Số phút 5–240');
      const ch = await docCauHinh(env);
      ch.mo_tam_den = Math.max(now, ch.mo_tam_den || 0) + phut * 60e3;
      await ghiCauHinh(env, ch);
      ghi = `Bố mẹ cho thêm ${phut} phút`;
    } else ghi = 'Bố mẹ không cho thêm giờ';
  } else {
    const luon = !!b.luon;
    if (luon) {
      const qt = dongY ? quyTacTuDuongDan(yc.duong_dan, yc.sha256) : { loai: 'hash', gia_tri: yc.sha256 };
      if (qt) {
        await env.DB.prepare(`INSERT INTO quy_tac(ts,loai,gia_tri,ten,cho_phep) VALUES(?,?,?,?,?)
          ON CONFLICT(loai,gia_tri) DO UPDATE SET cho_phep=excluded.cho_phep, ten=excluded.ten, ts=excluded.ts`)
          .bind(now, qt.loai, qt.gia_tri, yc.ten, dongY ? 1 : 0).run();
      }
    }
    ghi = `Bố mẹ ${dongY ? 'cho' : 'không cho'} chạy ${yc.ten}${luon ? (dongY ? ' (luôn cho phép)' : ' (cấm luôn)') : ' (lần này)'}`;
  }
  await env.DB.prepare('UPDATE yeu_cau SET trang_thai=?, xu_ly_luc=?, xu_ly_phut=? WHERE id=?')
    .bind(dongY ? 'dong_y' : 'tu_choi', now, phut, id).run();
  await ghiNhatKy(env, yc.may, 'tra_loi', ghi);
  return trangThai(env);
}

async function xoaQuyTac(req, env) {
  const { id } = await docBody(req);
  const r = await env.DB.prepare('SELECT ten,loai,cho_phep FROM quy_tac WHERE id=?').bind(Number(id)).first();
  if (!r) return loi('Không thấy', 404);
  await env.DB.prepare('DELETE FROM quy_tac WHERE id=?').bind(Number(id)).run();
  await ghiNhatKy(env, null, 'quy_tac', `Bố mẹ bỏ quy tắc ${r.cho_phep ? 'cho phép' : 'cấm'} ${r.ten || ''}`);
  return trangThai(env);
}

async function dangKyPush(req, env) {
  const { subscription: s, ten } = await docBody(req);
  if (!s || typeof s.endpoint !== 'string' || !/^https:\/\//.test(s.endpoint) || !s.keys?.p256dh || !s.keys?.auth) return loi('Đăng ký không hợp lệ');
  await env.DB.prepare(`INSERT INTO push_sub(endpoint,p256dh,auth,ts,ten) VALUES(?,?,?,?,?)
    ON CONFLICT(endpoint) DO UPDATE SET p256dh=excluded.p256dh, auth=excluded.auth, ts=excluded.ts`)
    .bind(s.endpoint, s.keys.p256dh, s.keys.auth, Date.now(), String(ten || '').slice(0, 80)).run();
  return json({ ok: true });
}

// ---------- Anh chup man hinh + hoat dong ----------
const GIU_ANH_MS = 7 * 864e5;
const TOI_DA_ANH = 1800 * 1024; // D1 gioi han 2MB moi dong

// Khoang [tu, den) cua mot ngay theo gio VN, ngay dang YYYY-MM-DD (mac dinh hom nay).
function khoangNgay(url) {
  let n = url.searchParams.get('ngay');
  if (!n || !/^\d{4}-\d{2}-\d{2}$/.test(n)) n = new Date(Date.now() + 7 * 3600e3).toISOString().slice(0, 10);
  const tu = Date.parse(n + 'T00:00:00+07:00');
  return [tu, tu + 864e5];
}

async function mayGuiAnh(req, env, url) {
  let buf = new Uint8Array(await req.arrayBuffer());
  // Agent >= 1.3: than = [anh nho][anh goc], header x-nho-len = do dai anh nho.
  let nho = null;
  const lenNho = Number(req.headers.get('x-nho-len') || 0);
  if (lenNho > 0 && lenNho < buf.length) {
    nho = buf.slice(0, lenNho);
    buf = buf.slice(lenNho);
    if (nho[0] !== 0xff || nho[1] !== 0xd8) nho = null;
  }
  if (buf.length < 100 || buf.length > TOI_DA_ANH || buf[0] !== 0xff || buf[1] !== 0xd8) return loi('Ảnh không hợp lệ hoặc quá lớn');
  const now = Date.now();
  let ts = Number(url.searchParams.get('ts'));
  if (!Number.isFinite(ts) || Math.abs(ts - now) > 864e5) ts = now;
  await env.DB.batch([
    env.DB.prepare('INSERT INTO anh(ts,may,ly_do,kich_thuoc,cua_so,jpeg,nho) VALUES(?,?,?,?,?,?,?)')
      .bind(ts, String(url.searchParams.get('may') || '').slice(0, 64), String(url.searchParams.get('ly_do') || '').slice(0, 120),
        buf.length, String(url.searchParams.get('cua_so') || '').slice(0, 300), buf, nho),
    env.DB.prepare('DELETE FROM anh WHERE ts < ?').bind(now - GIU_ANH_MS),
  ]);
  return json({ ok: true });
}

async function chupNgay(env) {
  const ch = await docCauHinh(env);
  ch.chup_luc = Date.now();
  await ghiCauHinh(env, ch);
  return json({ ok: true });
}

async function dsAnh(env, url) {
  const [tu, den] = khoangNgay(url);
  const { results } = await env.DB.prepare('SELECT id,ts,ly_do,kich_thuoc,cua_so FROM anh WHERE ts>=? AND ts<? ORDER BY ts DESC').bind(tu, den).all();
  return json({ anh: results });
}

// ?nho => anh thu nho (anh cu chua co thi tra anh goc).
async function xemAnh(env, id, nho) {
  const r = await env.DB.prepare(nho ? 'SELECT COALESCE(nho, jpeg) AS jpeg FROM anh WHERE id=?' : 'SELECT jpeg FROM anh WHERE id=?').bind(id).first();
  if (!r) return loi('Không thấy ảnh', 404);
  const b = r.jpeg instanceof ArrayBuffer ? new Uint8Array(r.jpeg) : Uint8Array.from(r.jpeg);
  return new Response(b, { headers: { 'content-type': 'image/jpeg', 'cache-control': 'private, max-age=604800' } });
}

async function dsHoatDong(env, url) {
  const [tu, den] = khoangNgay(url);
  const { results } = await env.DB.prepare('SELECT ts,exe,tieu_de,ranh FROM hoat_dong WHERE ts>=? AND ts<? ORDER BY ts').bind(tu, den).all();
  return json({ hoat_dong: results });
}

export default {
  async fetch(req, env, ctx) {
    const url = new URL(req.url);
    const p = url.pathname;
    try {
      if (p.startsWith('/api/may/')) {
        if (!laMay(req, env)) return loi('Sai khoá máy', 401);
        if (p === '/api/may/hoi' && req.method === 'POST') return await mayHoi(req, env, ctx);
        if (p === '/api/may/xin' && req.method === 'POST') return await mayXin(req, env, ctx);
        if (p === '/api/may/anh' && req.method === 'POST') return await mayGuiAnh(req, env, url);
        return loi('Không có API này', 404);
      }
      if (p === '/api/dang-nhap' && req.method === 'POST') return await dangNhap(req, env);
      if (p.startsWith('/api/')) {
        if (!(await duocVao(req, env))) return loi('Chưa đăng nhập', 401);
        if (p === '/api/trang-thai' && req.method === 'GET') return await trangThai(env);
        if (p === '/api/cau-hinh' && req.method === 'POST') return await luuCauHinh(req, env);
        if (p === '/api/mo-tam' && req.method === 'POST') return await moTam(req, env);
        if (p === '/api/khoa' && req.method === 'POST') return await khoaNgay(env);
        const m = /^\/api\/yeu-cau\/([0-9a-f-]{36})$/.exec(p);
        if (m && req.method === 'POST') return await xuLyYeuCau(req, env, m[1]);
        if (p === '/api/quy-tac/xoa' && req.method === 'POST') return await xoaQuyTac(req, env);
        if (p === '/api/chup' && req.method === 'POST') return await chupNgay(env);
        if (p === '/api/anh' && req.method === 'GET') return await dsAnh(env, url);
        const ma = /^\/api\/anh\/(\d+)$/.exec(p);
        if (ma && req.method === 'GET') return await xemAnh(env, Number(ma[1]), url.searchParams.has('nho'));
        if (p === '/api/hoat-dong' && req.method === 'GET') return await dsHoatDong(env, url);
        if (p === '/api/push/dang-ky' && req.method === 'POST') return await dangKyPush(req, env);
        if (p === '/api/push/thu' && req.method === 'POST')
          return json({ kq: await guiPush(env, { tieu_de: 'Thử thông báo', noi_dung: 'Điện thoại này sẽ nhận yêu cầu từ máy con.', the: 'thu' }) });
        return loi('Không có API này', 404);
      }
      return env.ASSETS.fetch(req);
    } catch (e) {
      return loi('Lỗi server: ' + (e && e.message), 500);
    }
  },
};
