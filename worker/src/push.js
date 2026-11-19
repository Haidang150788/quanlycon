// Web Push (RFC 8030 + VAPID RFC 8292 + ma hoa aes128gcm RFC 8291), chi dung WebCrypto.
// Khong qua Lark/Zalo: thong bao di thang tu Cloudflare toi dien thoai Sep.
// Khoa VAPID: secret VAPID_JWK (JWK ECDSA P-256 co "d"). Khoa cong khai suy ra tu x,y.

const b64u = {
  enc(buf) {
    const b = new Uint8Array(buf);
    let s = '';
    for (const x of b) s += String.fromCharCode(x);
    return btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
  },
  dec(s) {
    s = s.replace(/-/g, '+').replace(/_/g, '/');
    while (s.length % 4) s += '=';
    return Uint8Array.from(atob(s), (c) => c.charCodeAt(0));
  },
};
export { b64u };

const noi = (...arr) => {
  const n = arr.reduce((a, x) => a + x.length, 0);
  const out = new Uint8Array(n);
  let i = 0;
  for (const x of arr) { out.set(x, i); i += x.length; }
  return out;
};
const chu = (s) => new TextEncoder().encode(s);

export function khoaCongKhai(jwk) {
  return b64u.enc(noi(new Uint8Array([4]), b64u.dec(jwk.x), b64u.dec(jwk.y)));
}

async function hkdf(salt, ikm, info, len) {
  const k = await crypto.subtle.importKey('raw', ikm, 'HKDF', false, ['deriveBits']);
  return new Uint8Array(await crypto.subtle.deriveBits({ name: 'HKDF', hash: 'SHA-256', salt, info }, k, len * 8));
}

// Ma hoa noi dung cho mot thue bao. Tra ve body aes128gcm.
export async function maHoa(noiDung, p256dhB64, authB64, { salt, asKeys } = {}) {
  const uaPub = b64u.dec(p256dhB64);
  const auth = b64u.dec(authB64);
  salt = salt || crypto.getRandomValues(new Uint8Array(16));
  asKeys = asKeys || (await crypto.subtle.generateKey({ name: 'ECDH', namedCurve: 'P-256' }, true, ['deriveBits']));
  const asPub = new Uint8Array(await crypto.subtle.exportKey('raw', asKeys.publicKey));
  const uaKey = await crypto.subtle.importKey('raw', uaPub, { name: 'ECDH', namedCurve: 'P-256' }, false, []);
  const ecdh = new Uint8Array(await crypto.subtle.deriveBits({ name: 'ECDH', public: uaKey }, asKeys.privateKey, 256));
  const ikm = await hkdf(auth, ecdh, noi(chu('WebPush: info\0'), uaPub, asPub), 32);
  const cek = await hkdf(salt, ikm, chu('Content-Encoding: aes128gcm\0'), 16);
  const nonce = await hkdf(salt, ikm, chu('Content-Encoding: nonce\0'), 12);
  const key = await crypto.subtle.importKey('raw', cek, 'AES-GCM', false, ['encrypt']);
  const ct = new Uint8Array(await crypto.subtle.encrypt({ name: 'AES-GCM', iv: nonce }, key, noi(chu(noiDung), new Uint8Array([2]))));
  const rs = new Uint8Array([0, 0, 16, 0]); // 4096
  return noi(salt, rs, new Uint8Array([asPub.length]), asPub, ct);
}

async function vapidJWT(jwk, aud, sub) {
  const key = await crypto.subtle.importKey('jwk', { ...jwk, key_ops: ['sign'], ext: true }, { name: 'ECDSA', namedCurve: 'P-256' }, false, ['sign']);
  const dau = b64u.enc(chu(JSON.stringify({ typ: 'JWT', alg: 'ES256' })));
  const than = b64u.enc(chu(JSON.stringify({ aud, exp: Math.floor(Date.now() / 1000) + 12 * 3600, sub })));
  const ky = await crypto.subtle.sign({ name: 'ECDSA', hash: 'SHA-256' }, key, chu(dau + '.' + than));
  return dau + '.' + than + '.' + b64u.enc(ky);
}

// Gui toi moi thue bao. Thue bao chet (404/410) thi xoa. Tra ve [{endpoint, status}].
export async function guiPush(env, tin) {
  if (!env.VAPID_JWK) return [];
  const jwk = JSON.parse(env.VAPID_JWK);
  const pub = khoaCongKhai(jwk);
  const { results } = await env.DB.prepare('SELECT endpoint,p256dh,auth FROM push_sub').all();
  const kq = [];
  for (const s of results) {
    try {
      const aud = new URL(s.endpoint).origin;
      const body = await maHoa(JSON.stringify(tin), s.p256dh, s.auth);
      const r = await fetch(s.endpoint, {
        method: 'POST',
        headers: {
          authorization: `vapid t=${await vapidJWT(jwk, aud, env.VAPID_SUB || "mailto:quanlycon@example.com")}, k=${pub}`,
          'content-encoding': 'aes128gcm',
          'content-type': 'application/octet-stream',
          ttl: '3600',
          urgency: 'high',
        },
        body,
      });
      kq.push({ endpoint: s.endpoint.slice(0, 60), status: r.status });
      if (r.status === 404 || r.status === 410) await env.DB.prepare('DELETE FROM push_sub WHERE endpoint=?').bind(s.endpoint).run();
    } catch (e) {
      kq.push({ endpoint: s.endpoint.slice(0, 60), loi: String(e && e.message) });
    }
  }
  return kq;
}
