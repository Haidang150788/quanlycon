import { test } from 'node:test';
import assert from 'node:assert/strict';
import { dangChan, chuanHoaGio, chuanHoaWeb, chuanHoaCauHinh, phutTrongNgayVN } from '../src/luat.js';

// 2026-11-17 11:45 gio VN = 04:45 UTC
const luc = (hhmm) => Date.UTC(2026, 10, 17, 0, 0) - 7 * 3600e3 + (Number(hhmm.slice(0, 2)) * 60 + Number(hhmm.slice(3))) * 60e3;
const CH = { bat: true, giai_lao: [{ tu: '11:30', den: '12:59' }], mo_tam_den: 0 };

test('gio VN', () => { assert.equal(phutTrongNgayVN(luc('00:05')), 5); assert.equal(phutTrongNgayVN(luc('23:59')), 23 * 60 + 59); });
test('trong gio giai lao thi khong chan, ca phut cuoi', () => {
  assert.equal(dangChan(CH, luc('11:29')), true);
  assert.equal(dangChan(CH, luc('11:30')), false);
  assert.equal(dangChan(CH, luc('12:59')), false);
  assert.equal(dangChan(CH, luc('13:00')), true);
});
test('khung qua dem', () => {
  const c = { ...CH, giai_lao: [{ tu: '21:00', den: '01:00' }] };
  assert.equal(dangChan(c, luc('23:00')), false);
  assert.equal(dangChan(c, luc('00:30')), false);
  assert.equal(dangChan(c, luc('01:01')), true);
  assert.equal(dangChan(c, luc('20:59')), true);
});
test('mo tam va tat cong tac', () => {
  const t = luc('15:00');
  assert.equal(dangChan({ ...CH, mo_tam_den: t + 60e3 }, t), false);
  assert.equal(dangChan({ ...CH, mo_tam_den: t - 1 }, t), true);
  assert.equal(dangChan({ ...CH, bat: false }, t), false);
});
test('chuan hoa gio', () => {
  assert.equal(chuanHoaGio('11h30'), '11:30');
  assert.equal(chuanHoaGio('7:05'), '07:05');
  assert.equal(chuanHoaGio('11h'), '11:00');
  assert.equal(chuanHoaGio('24:00'), null);
  assert.equal(chuanHoaGio('11:3'), null);
  assert.equal(chuanHoaGio('vd 11:30'), null);
});
test('chuan hoa web', () => {
  assert.equal(chuanHoaWeb('https://Shopee.vn/'), 'shopee.vn');
  assert.equal(chuanHoaWeb('thomaytinh.com/blog'), 'thomaytinh.com/blog');
  assert.equal(chuanHoaWeb('abc'), null);
  assert.equal(chuanHoaWeb('a b.com'), null);
});
test('cau hinh: bo dong trong, bao loi gio sai, loai nen tang la', () => {
  const ma = new Set(['facebook']);
  const r = chuanHoaCauHinh({ nen_tang: ['facebook', 'xyz'], web_khac: ['', 'shopee.vn', 'shopee.vn'], giai_lao: [{ tu: '', den: '' }, { tu: '11h30', den: '12:59' }] }, ma);
  assert.deepEqual(r.cau_hinh.nen_tang, ['facebook']);
  assert.deepEqual(r.cau_hinh.web_khac, ['shopee.vn']);
  assert.deepEqual(r.cau_hinh.giai_lao, [{ tu: '11:30', den: '12:59' }]);
  assert.match(chuanHoaCauHinh({ giai_lao: [{ tu: '11:30', den: '' }] }, ma).loi, /không hợp lệ/);
});
