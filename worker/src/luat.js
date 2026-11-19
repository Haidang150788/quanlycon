// Luat "luc nay co chan khong". BAN SAO cua agent/lich.go — sua mot ben phai sua ben kia.
// Gio Viet Nam = UTC+7 co dinh (khong co gio mua he).

export const LECH_VN_MS = 7 * 3600e3;

export function phutTrongNgayVN(ms) {
  const d = new Date(ms + LECH_VN_MS);
  return d.getUTCHours() * 60 + d.getUTCMinutes();
}

// "11:30", "11h30", "11h", "7:05" => "11:30"; sai => null
export function chuanHoaGio(s) {
  const m = /^\s*(\d{1,2})\s*[:hH.]\s*(\d{0,2})\s*$/.exec(String(s ?? ''));
  if (!m) return null;
  const h = Number(m[1]);
  const p = m[2] === '' ? 0 : Number(m[2]);
  if (h > 23 || p > 59 || (m[2].length === 1)) return null;
  return String(h).padStart(2, '0') + ':' + String(p).padStart(2, '0');
}
const phut = (hhmm) => Number(hhmm.slice(0, 2)) * 60 + Number(hhmm.slice(3, 5));

// Phut "den" TINH CA phut do: 11:30–12:59 => duoc choi den het 12:59.
// tu > den => khung qua dem (vd 21:00–01:00).
export function trongKhung(m, khung) {
  const a = phut(khung.tu), b = phut(khung.den);
  return a <= b ? m >= a && m <= b : m >= a || m <= b;
}

export function dangChan(ch, nowMs) {
  if (!ch.bat) return false;
  if (ch.mo_tam_den && ch.mo_tam_den > nowMs) return false;
  const m = phutTrongNgayVN(nowMs);
  return !ch.giai_lao.some((k) => trongKhung(m, k));
}

// "https://Shopee.vn/abc/" => "shopee.vn/abc"; sai => null
export function chuanHoaWeb(s) {
  let w = String(s ?? '').trim().toLowerCase();
  w = w.replace(/^[a-z]+:\/\//, '').replace(/\/+$/, '').replace(/^\*\./, '');
  if (!w) return null;
  const [host, ...duong] = w.split('/');
  if (!/^(?=.{3,253}$)([a-z0-9]([a-z0-9-]*[a-z0-9])?\.)+[a-z]{2,}$/.test(host)) return null;
  const path = duong.join('/');
  if (path && !/^[a-z0-9._~\-/%]+$/.test(path)) return null;
  return path ? host + '/' + path : host;
}

export function chuanHoaCauHinh(v, maHopLe) {
  const ch = {
    bat: v.bat !== false,
    nen_tang: [...new Set((Array.isArray(v.nen_tang) ? v.nen_tang : []).filter((m) => maHopLe.has(m)))],
    web_khac: [],
    giai_lao: [],
    dong_trinh_duyet: v.dong_trinh_duyet !== false,
    khoa_cai: v.khoa_cai !== false,
    chup_phut: [0, 5, 10, 15, 30].includes(Number(v.chup_phut)) ? Number(v.chup_phut) : 10,
    chup_luc: Number(v.chup_luc) || 0,
    mo_tam_den: Number(v.mo_tam_den) || 0,
    khoa_luc: Number(v.khoa_luc) || 0,
  };
  const webs = Array.isArray(v.web_khac) ? v.web_khac : [];
  if (webs.length > 200) return { loi: 'Tối đa 200 website khác' };
  for (const w of webs) {
    if (String(w).trim() === '') continue;
    const c = chuanHoaWeb(w);
    if (!c) return { loi: `Website không hợp lệ: "${w}"` };
    if (!ch.web_khac.includes(c)) ch.web_khac.push(c);
  }
  const kh = Array.isArray(v.giai_lao) ? v.giai_lao : [];
  for (const k of kh) {
    if (!k || (String(k.tu ?? '').trim() === '' && String(k.den ?? '').trim() === '')) continue;
    const tu = chuanHoaGio(k.tu), den = chuanHoaGio(k.den);
    if (!tu || !den) return { loi: `Giờ giải lao không hợp lệ: "${k.tu ?? ''}" đến "${k.den ?? ''}"` };
    ch.giai_lao.push({ tu, den });
  }
  if (ch.giai_lao.length > 5) return { loi: 'Tối đa 5 khung giải lao' };
  return { cau_hinh: ch };
}

// "Luon cho phep" mot file .exe: neu nam trong thu muc app cua nguoi dung (AppData\Local\Zoom\...,
// AppData\Local\Programs\X\...) hoac ProgramData\X\ thi cho ca thu muc app do (app tu cap nhat
// sang file moi van chay duoc). Con lai (Downloads, Desktop, Temp, USB...) thi chi cho dung file do (sha256).
export function quyTacTuDuongDan(duongDan, sha256) {
  const p = String(duongDan || '').toLowerCase().replace(/\//g, '\\');
  if (!/\\temp\\/.test(p)) {
    let m = /^([a-z]:\\users\\[^\\]+\\appdata\\(?:local|roaming|locallow)\\(?:programs\\)?[^\\]+\\)[^\\]/.exec(p);
    if (!m) m = /^([a-z]:\\programdata\\[^\\]+\\)[^\\]/.exec(p);
    if (m && !/\\appdata\\local\\(?:packages|microsoft)\\$|^[a-z]:\\programdata\\microsoft\\$/.test(m[1])) return { loai: 'thu_muc', gia_tri: m[1] };
  }
  if (!/^[0-9a-f]{64}$/.test(String(sha256 || ''))) return null;
  return { loai: 'hash', gia_tri: sha256 };
}
