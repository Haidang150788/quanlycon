import { test } from 'node:test';
import assert from 'node:assert/strict';
import { phanTichLinkKenh, docKenhTuHtml, traKenh, duongTuTacGia } from '../src/kenh.js';
import '../../extension/luat-kenh.js';

const L = globalThis.LuatKenh;
const UC = 'UCabcdefghijklmnopqrstuv'; // 24 ky tu
const UC2 = 'UCzzzzzzzzzzzzzzzzzzzzzz';

test('phan tich link kenh / video bo me dan', () => {
  assert.deepEqual(phanTichLinkKenh('https://www.youtube.com/@HocToan123'), { loai: 'handle', gia_tri: '@HocToan123' });
  assert.deepEqual(phanTichLinkKenh('youtube.com/@HocToan123/videos'), { loai: 'handle', gia_tri: '@HocToan123' });
  assert.deepEqual(phanTichLinkKenh('@HocToan123'), { loai: 'handle', gia_tri: '@HocToan123' });
  assert.deepEqual(phanTichLinkKenh('https://m.youtube.com/@h%E1%BB%8Dc'), { loai: 'handle', gia_tri: '@học' });
  assert.deepEqual(phanTichLinkKenh(`https://www.youtube.com/channel/${UC}`), { loai: 'kenh', gia_tri: UC });
  assert.deepEqual(phanTichLinkKenh(UC), { loai: 'kenh', gia_tri: UC });
  assert.deepEqual(phanTichLinkKenh('https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=10s'), { loai: 'video', gia_tri: 'dQw4w9WgXcQ' });
  assert.deepEqual(phanTichLinkKenh('https://youtu.be/dQw4w9WgXcQ?si=abc'), { loai: 'video', gia_tri: 'dQw4w9WgXcQ' });
  assert.deepEqual(phanTichLinkKenh('https://www.youtube.com/shorts/dQw4w9WgXcQ'), { loai: 'video', gia_tri: 'dQw4w9WgXcQ' });
  assert.deepEqual(phanTichLinkKenh('https://www.youtube.com/c/TenCu'), { loai: 'duong', gia_tri: '/c/TenCu' });
  assert.equal(phanTichLinkKenh('https://www.youtube.com/'), null);
  assert.equal(phanTichLinkKenh('https://evil.com/@abc'), null);
  assert.equal(phanTichLinkKenh('https://youtube.com.evil.com/@abc'), null);
  assert.equal(phanTichLinkKenh('https://www.youtube.com/watch?v=ngan'), null);
  assert.equal(phanTichLinkKenh(''), null);
});

test('doc ma kenh + ten + handle tu HTML trang kenh', () => {
  const html = `<html><head><title>Học Toán &amp; Vui - YouTube</title><meta property="og:title" content="Học Toán &amp; Vui">
    <link rel="canonical" href="https://www.youtube.com/channel/${UC}"></head>
    <script>var ytInitialData = {"metadata":{"channelMetadataRenderer":{"externalId":"${UC}","vanityChannelUrl":"http://www.youtube.com/@HocToanVui"}}}</script>`;
  assert.deepEqual(docKenhTuHtml(html), { id: UC, ten: 'Học Toán & Vui', handle: '@HocToanVui' });
  assert.equal(docKenhTuHtml('<html>consent.youtube.com</html>'), null);
  assert.equal(duongTuTacGia('https://www.youtube.com/@HocToanVui'), '/@HocToanVui');
  assert.equal(duongTuTacGia('https://www.youtube.com/user/abc'), '/user/abc');
});

test('tra kenh tu link video: oEmbed -> trang kenh', async () => {
  const goi = [];
  const gia = async (u) => {
    goi.push(u);
    if (u.includes('/oembed')) return { ok: true, status: 200, json: async () => ({ author_name: 'Học Toán', author_url: 'https://www.youtube.com/@HocToanVui' }) };
    if (u === 'https://www.youtube.com/@HocToanVui') return { ok: true, status: 200, text: async () => `<meta property="og:title" content="Học Toán"><link rel="canonical" href="https://www.youtube.com/channel/${UC}">` };
    return { ok: false, status: 404 };
  };
  assert.deepEqual(await traKenh('https://youtu.be/dQw4w9WgXcQ', gia), { id: UC, ten: 'Học Toán', handle: '@HocToanVui' });
  assert.equal(goi.length, 2);
  assert.match((await traKenh('https://www.youtube.com/@khongco', gia)).loi, /Không thấy kênh/);
  assert.match((await traKenh('facebook.com/abc', gia)).loi, /Không hiểu link/);
});

const TT = { han_che: true, kenh: [{ id: UC, handle: '@HocToanVui', ten: 'Học Toán' }] };
const V = (id, kenh_id, handle) => ({ id, kenh_id, handle });

test('gio mo: xem thoai mai moi trang', () => {
  const tt = { han_che: false, kenh: [] };
  for (const u of ['https://www.youtube.com/', 'https://www.youtube.com/results?search_query=game', 'https://www.youtube.com/watch?v=dQw4w9WgXcQ'])
    assert.equal(L.quyetDinh(tt, u, null).ket_qua, 'cho');
});

test('gio chan: video kenh duoc chon thi xem, kenh khac thi chan', () => {
  const u = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ';
  assert.equal(L.quyetDinh(TT, u, V('dQw4w9WgXcQ', UC)).ket_qua, 'cho');
  assert.equal(L.quyetDinh(TT, u, V('dQw4w9WgXcQ', '', '@hoctoanvui')).ket_qua, 'cho', 'so @handle khong phan biet hoa thuong');
  assert.equal(L.quyetDinh(TT, u, V('dQw4w9WgXcQ', UC2, '@GameHay')).ket_qua, 'chan');
  assert.equal(L.quyetDinh(TT, 'https://www.youtube.com/shorts/dQw4w9WgXcQ', V('dQw4w9WgXcQ', UC2)).ket_qua, 'chan');
  assert.equal(L.quyetDinh(TT, 'https://www.youtube-nocookie.com/embed/dQw4w9WgXcQ', V('dQw4w9WgXcQ', UC)).ket_qua, 'cho');
  assert.equal(L.quyetDinh(TT, 'https://www.youtube.com/embed/dQw4w9WgXcQ', V('dQw4w9WgXcQ', UC2)).ket_qua, 'chan');
});

test('gio chan: chua biet kenh hoac thong tin cua video cu thi doi, khong cho', () => {
  const u = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ';
  assert.equal(L.quyetDinh(TT, u, null).ket_qua, 'doi');
  assert.equal(L.quyetDinh(TT, u, V('xxxxxxxxxxx', UC)).ket_qua, 'doi', 'trinh phat con du lieu video truoc (doi trang SPA)');
  assert.equal(L.quyetDinh(TT, u, V('dQw4w9WgXcQ', '', '')).ket_qua, 'doi');
});

test('gio chan: trang chu, tim kiem, kenh khac, YouTube Music bi chan; trang kenh duoc chon thi vao', () => {
  for (const u of ['https://www.youtube.com/', 'https://www.youtube.com/results?search_query=a', 'https://www.youtube.com/feed/trending',
    'https://www.youtube.com/shorts', 'https://www.youtube.com/@GameHay', `https://www.youtube.com/channel/${UC2}`, 'https://music.youtube.com/watch?v=dQw4w9WgXcQ',
    'https://www.youtube.com/playlist?list=PL123'])
    assert.equal(L.quyetDinh(TT, u, null).ket_qua, 'chan', u);
  assert.equal(L.quyetDinh(TT, 'https://www.youtube.com/@hoctoanvui/videos', null).ket_qua, 'cho');
  assert.equal(L.quyetDinh(TT, `https://m.youtube.com/channel/${UC}`, null).ket_qua, 'cho');
});

test('khong hoi duoc dich vu (tt rong) hoac danh sach rong => chan', () => {
  const u = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ';
  assert.equal(L.quyetDinh(null, u, V('dQw4w9WgXcQ', UC)).ket_qua, 'chan');
  assert.equal(L.quyetDinh({ han_che: true, kenh: [] }, u, V('dQw4w9WgXcQ', UC)).ket_qua, 'chan');
});

test('oEmbed author_url => handle / ma kenh', () => {
  assert.deepEqual(L.tacGiaTuOembed('https://www.youtube.com/@HocToanVui'), { handle: '@hoctoanvui' });
  assert.deepEqual(L.tacGiaTuOembed(`http://www.youtube.com/channel/${UC}`), { kenh_id: UC });
  assert.equal(L.tacGiaTuOembed('https://www.youtube.com/user/abc'), null);
});
