// Kenh YouTube duoc xem trong gio chan. Bo me dan link kenh hoac link mot video cua kenh,
// server tu tra ra ma kenh (UC...) + ten + @handle roi luu vao D1 (bang kenh_yt).
// Luat "video nay co duoc xem khong" nam o extension/luat-kenh.js (chay trong trinh duyet may con).

const MA_KENH = /^UC[A-Za-z0-9_-]{22}$/;
const MA_VIDEO = /^[A-Za-z0-9_-]{11}$/;
const MIEN_YT = /^(?:www\.|m\.|music\.)?(?:youtube\.com|youtube-nocookie\.com)$/;

// Link/chu bo me dan => { loai: 'kenh'|'handle'|'duong'|'video', gia_tri } ; khong hieu => null
//   kenh   : 'UCxxxxxxxxxxxxxxxxxxxxxx'
//   handle : '@ten' (giu nguyen hoa thuong, so khop thi bo qua hoa thuong)
//   duong  : '/c/ten' | '/user/ten' (link kieu cu)
//   video  : ma video 11 ky tu
export function phanTichLinkKenh(s) {
  let t = String(s ?? '').trim();
  if (!t) return null;
  if (MA_KENH.test(t)) return { loai: 'kenh', gia_tri: t };
  if (/^@[^\s/?#]{1,100}$/.test(t)) return { loai: 'handle', gia_tri: giaiMa(t) };
  if (!/^[a-z]+:\/\//i.test(t)) t = 'https://' + t;
  let u;
  try { u = new URL(t); } catch { return null; }
  const host = u.hostname.toLowerCase();
  const phan = u.pathname.split('/').filter(Boolean);
  if (host === 'youtu.be') return phan[0] && MA_VIDEO.test(phan[0]) ? { loai: 'video', gia_tri: phan[0] } : null;
  if (!MIEN_YT.test(host)) return null;
  const dau = phan[0] || '';
  if (dau === 'watch') {
    const v = u.searchParams.get('v') || '';
    return MA_VIDEO.test(v) ? { loai: 'video', gia_tri: v } : null;
  }
  if (['shorts', 'embed', 'live', 'v'].includes(dau)) return phan[1] && MA_VIDEO.test(phan[1]) ? { loai: 'video', gia_tri: phan[1] } : null;
  if (dau === 'channel') return phan[1] && MA_KENH.test(phan[1]) ? { loai: 'kenh', gia_tri: phan[1] } : null;
  if (dau.startsWith('@') && dau.length > 1) return { loai: 'handle', gia_tri: giaiMa(dau) };
  if ((dau === 'c' || dau === 'user') && phan[1]) return { loai: 'duong', gia_tri: `/${dau}/${phan[1]}` };
  return null;
}

function giaiMa(s) { try { return decodeURIComponent(s); } catch { return s; } }

const THUC_THE = { amp: '&', lt: '<', gt: '>', quot: '"', apos: "'", '#39': "'" };
function boThucThe(s) {
  return s.replace(/&(#x[0-9a-f]+|#\d+|[a-z0-9]+);/gi, (m, k) => {
    if (k[0] === '#') return String.fromCodePoint(k[1] === 'x' || k[1] === 'X' ? parseInt(k.slice(2), 16) : Number(k.slice(1)));
    return THUC_THE[k.toLowerCase()] ?? m;
  });
}

// HTML trang kenh YouTube => { id, ten, handle } ; khong thay ma kenh => null
export function docKenhTuHtml(html) {
  const h = String(html || '');
  const id = (/<link rel="canonical" href="https:\/\/www\.youtube\.com\/channel\/(UC[A-Za-z0-9_-]{22})"/.exec(h)
    || /"externalId":"(UC[A-Za-z0-9_-]{22})"/.exec(h)
    || /<meta itemprop="(?:identifier|channelId)" content="(UC[A-Za-z0-9_-]{22})"/.exec(h))?.[1];
  if (!id) return null;
  const ten = boThucThe((/<meta property="og:title" content="([^"]*)"/.exec(h) || /<title>([^<]*?)(?: - YouTube)?<\/title>/.exec(h) || [])[1] || '').trim();
  let handle = (/"vanityChannelUrl":"https?:\/\/www\.youtube\.com\/(@[^"\\/?]+)"/.exec(h)
    || /"canonicalBaseUrl":"\/(@[^"\\/?]+)"/.exec(h))?.[1] || '';
  handle = giaiMa(handle.replace(/\\u0026/g, '&'));
  return { id, ten: ten.slice(0, 120), handle: handle.slice(0, 120) };
}

// oEmbed author_url "https://www.youtube.com/@abc" | ".../channel/UC..." | ".../user/abc" => duong dan trang kenh
export function duongTuTacGia(authorUrl) {
  const p = phanTichLinkKenh(authorUrl);
  if (!p) return null;
  if (p.loai === 'kenh') return '/channel/' + p.gia_tri;
  if (p.loai === 'handle') return '/' + encodeURIComponent(p.gia_tri).replace(/^%40/, '@');
  if (p.loai === 'duong') return p.gia_tri;
  return null;
}

const DAU_HOI = {
  'accept-language': 'vi,en;q=0.8',
  'user-agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36',
  cookie: 'SOCS=CAI; CONSENT=YES+1', // bo qua trang "dong y cookie" o chau Au
};

// Tra link => { id, ten, handle } hoac { loi }. layTrang(url) => Promise<{ok,status,text()}> (de kiem thu thay fetch).
export async function traKenh(link, layTrang = (u) => fetch(u, { headers: DAU_HOI, redirect: 'follow' })) {
  const p = phanTichLinkKenh(link);
  if (!p) return { loi: 'Không hiểu link. Dán link kênh (youtube.com/@ten) hoặc link một video của kênh.' };
  let duong;
  if (p.loai === 'kenh') duong = '/channel/' + p.gia_tri;
  else if (p.loai === 'handle') duong = '/' + encodeURIComponent(p.gia_tri).replace(/^%40/, '@');
  else if (p.loai === 'duong') duong = p.gia_tri;
  else {
    const r = await layTrang('https://www.youtube.com/oembed?format=json&url=' + encodeURIComponent('https://www.youtube.com/watch?v=' + p.gia_tri));
    if (!r.ok) return { loi: r.status === 404 || r.status === 400 ? 'Không thấy video này (video riêng tư hoặc đã xoá?)' : `YouTube trả lỗi ${r.status} khi đọc video` };
    const j = await r.json().catch(() => ({}));
    duong = duongTuTacGia(j.author_url || '');
    if (!duong) return { loi: 'Không đọc được kênh của video này' };
  }
  const r = await layTrang('https://www.youtube.com' + duong);
  if (!r.ok) return { loi: r.status === 404 ? 'Không thấy kênh này trên YouTube' : `YouTube trả lỗi ${r.status} khi đọc kênh` };
  const k = docKenhTuHtml(await r.text());
  if (!k) return { loi: 'Không đọc được mã kênh từ trang YouTube (YouTube đổi giao diện?)' };
  if (!k.handle && duong.startsWith('/@')) k.handle = giaiMa(duong.slice(1)).slice(0, 120);
  return k;
}
