// Service worker: hoi dich vu QuanLyCon tren chinh may (127.0.0.1:47123) moi 30 giay.
// Moi lan hoi cung la "bao ve" — dich vu chi bo YouTube khoi danh sach chan cua trinh duyet nay khi nghe thay.
// Khong hoi duoc dich vu => coi nhu dang gio chan (dong khi khong chac).
importScripts('luat-kenh.js');

const DICH_VU = 'http://127.0.0.1:47123/ext/trang-thai';
let tt = null;
let ttLuc = 0;

async function hoi() {
  try {
    // POST: Chrome chi chac chan gui Origin (chrome-extension://<ma>) voi POST; dich vu dua vao do de nhan extension.
    const r = await fetch(DICH_VU, { method: 'POST', cache: 'no-store' });
    if (!r.ok) throw new Error('HTTP ' + r.status);
    const j = await r.json();
    tt = { han_che: j.han_che !== false, kenh: Array.isArray(j.kenh) ? j.kenh : [], con_ms: j.con_ms, nhan: !!j.nhan, ok: true };
    await chrome.storage.local.set({ kenh: tt.kenh });
  } catch (e) {
    const s = await chrome.storage.local.get('kenh').catch(() => ({}));
    tt = { han_che: true, kenh: Array.isArray(s.kenh) ? s.kenh : [], ok: false, loi: String(e && e.message || e) };
  }
  ttLuc = Date.now();
  return tt;
}

async function trangThai() {
  if (tt && Date.now() - ttLuc < 10e3) return tt;
  return hoi();
}

// Ma video => { kenh_id } | { handle } qua oEmbed (khong can khoa API). Nho 1 gio.
const nhoOembed = new Map();
async function oembed(id) {
  const c = nhoOembed.get(id);
  if (c && Date.now() - c.luc < 3600e3) return c.kq;
  let kq = null;
  try {
    const r = await fetch('https://www.youtube.com/oembed?format=json&url=' + encodeURIComponent('https://www.youtube.com/watch?v=' + id), { credentials: 'omit' });
    if (r.ok) kq = globalThis.LuatKenh.tacGiaTuOembed((await r.json()).author_url) || null;
  } catch { /* mat mang: tra null, content script tu dung trinh phat */ }
  nhoOembed.set(id, { luc: Date.now(), kq });
  return kq;
}

chrome.runtime.onMessage.addListener((m, _nguoiGui, traLoi) => {
  if (m && m.hoi === 'tt') { trangThai().then(traLoi); return true; }
  if (m && m.hoi === 'oembed' && /^[A-Za-z0-9_-]{11}$/.test(m.id || '')) { oembed(m.id).then(traLoi); return true; }
  return false;
});

chrome.alarms.create('hoi', { periodInMinutes: 0.5 });
chrome.alarms.onAlarm.addListener((a) => { if (a.name === 'hoi') hoi(); });
chrome.runtime.onStartup.addListener(hoi);
chrome.runtime.onInstalled.addListener(hoi);
hoi();
