import { test } from 'node:test';
import assert from 'node:assert/strict';
import { maHoa, b64u, khoaCongKhai } from '../src/push.js';
import { quyTacTuDuongDan } from '../src/luat.js';

const chu = (s) => new TextEncoder().encode(s);
async function hkdf(salt, ikm, info, len) {
  const k = await crypto.subtle.importKey('raw', ikm, 'HKDF', false, ['deriveBits']);
  return new Uint8Array(await crypto.subtle.deriveBits({ name: 'HKDF', hash: 'SHA-256', salt, info }, k, len * 8));
}
const noi = (...a) => { const o = new Uint8Array(a.reduce((n, x) => n + x.length, 0)); let i = 0; for (const x of a) { o.set(x, i); i += x.length; } return o; };

// Giai ma phia trinh duyet theo RFC 8291 — doc lap voi code ma hoa.
test('ma hoa aes128gcm giai ma lai duoc dung noi dung', async () => {
  const ua = await crypto.subtle.generateKey({ name: 'ECDH', namedCurve: 'P-256' }, true, ['deriveBits']);
  const uaPub = new Uint8Array(await crypto.subtle.exportKey('raw', ua.publicKey));
  const auth = crypto.getRandomValues(new Uint8Array(16));
  const body = await maHoa('{"tieu_de":"Con xin thêm 30 phút"}', b64u.enc(uaPub), b64u.enc(auth));
  const salt = body.slice(0, 16);
  assert.equal(new DataView(body.buffer, body.byteOffset).getUint32(16), 4096);
  const idlen = body[20];
  const asPub = body.slice(21, 21 + idlen);
  const ct = body.slice(21 + idlen);
  const asKey = await crypto.subtle.importKey('raw', asPub, { name: 'ECDH', namedCurve: 'P-256' }, false, []);
  const ecdh = new Uint8Array(await crypto.subtle.deriveBits({ name: 'ECDH', public: asKey }, ua.privateKey, 256));
  const ikm = await hkdf(auth, ecdh, noi(chu('WebPush: info\0'), uaPub, asPub), 32);
  const cek = await hkdf(salt, ikm, chu('Content-Encoding: aes128gcm\0'), 16);
  const nonce = await hkdf(salt, ikm, chu('Content-Encoding: nonce\0'), 12);
  const k = await crypto.subtle.importKey('raw', cek, 'AES-GCM', false, ['decrypt']);
  const pt = new Uint8Array(await crypto.subtle.decrypt({ name: 'AES-GCM', iv: nonce }, k, ct));
  assert.equal(pt[pt.length - 1], 2);
  assert.equal(new TextDecoder().decode(pt.slice(0, -1)), '{"tieu_de":"Con xin thêm 30 phút"}');
});

test('khoa cong khai VAPID 65 byte bat dau 0x04', async () => {
  const k = await crypto.subtle.generateKey({ name: 'ECDSA', namedCurve: 'P-256' }, true, ['sign']);
  const jwk = await crypto.subtle.exportKey('jwk', k.privateKey);
  const raw = b64u.dec(khoaCongKhai(jwk));
  assert.equal(raw.length, 65); assert.equal(raw[0], 4);
});

const H = 'a'.repeat(64);
test('luon cho phep: app trong AppData => ca thu muc app', () => {
  assert.deepEqual(quyTacTuDuongDan('C:\\Users\\Khoi\\AppData\\Local\\Zoom\\bin\\Zoom.exe', H), { loai: 'thu_muc', gia_tri: 'c:\\users\\khoi\\appdata\\local\\zoom\\' });
  assert.deepEqual(quyTacTuDuongDan('C:\\Users\\Khoi\\AppData\\Local\\Programs\\Scratch\\Scratch.exe', H), { loai: 'thu_muc', gia_tri: 'c:\\users\\khoi\\appdata\\local\\programs\\scratch\\' });
  assert.deepEqual(quyTacTuDuongDan('C:\\ProgramData\\Vendor\\x.exe', H), { loai: 'thu_muc', gia_tri: 'c:\\programdata\\vendor\\' });
});
test('luon cho phep: Downloads/Temp/USB/thu muc qua rong => chi file do', () => {
  for (const p of ['C:\\Users\\Khoi\\Downloads\\setup.exe', 'C:\\Users\\Khoi\\AppData\\Local\\Temp\\is-1\\setup.exe',
    'E:\\game.exe', 'C:\\Users\\Khoi\\AppData\\Local\\Microsoft\\x\\a.exe', 'C:\\ProgramData\\Microsoft\\a\\b.exe',
    'C:\\Users\\Khoi\\AppData\\Local\\game.exe']) {
    assert.deepEqual(quyTacTuDuongDan(p, H), { loai: 'hash', gia_tri: H }, p);
  }
  assert.equal(quyTacTuDuongDan('E:\\x.exe', 'sai'), null);
});
