// Tao 2 bi mat cho server: AGENT_TOKEN (khoa cua may con) va VAPID_JWK (khoa thong bao day).
// Chay: node scripts/tao-bi-mat.mjs   — in ra lenh `wrangler secret put` de dan.
import { webcrypto as c } from 'node:crypto';
const token = Buffer.from(c.getRandomValues(new Uint8Array(24))).toString('hex');
const k = await c.subtle.generateKey({ name: 'ECDSA', namedCurve: 'P-256' }, true, ['sign']);
const { kty, crv, x, y, d } = await c.subtle.exportKey('jwk', k.privateKey);
const jwk = JSON.stringify({ kty, crv, x, y, d });
console.log('Ghi lại 2 dòng này ở nơi an toàn (mất VAPID_JWK thì điện thoại phải bật lại thông báo):\n');
console.log('AGENT_TOKEN=' + token);
console.log('VAPID_JWK=' + jwk + '\n');
console.log('Rồi chạy trong thư mục worker/ (mỗi lệnh sẽ hỏi giá trị, dán vào):');
console.log('  npx wrangler secret put PIN            # mật khẩu bạn tự đặt để vào web trên điện thoại');
console.log('  npx wrangler secret put AGENT_TOKEN    # dán AGENT_TOKEN ở trên');
console.log('  npx wrangler secret put VAPID_JWK      # dán VAPID_JWK ở trên');
