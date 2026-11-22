// Luat "trang YouTube nay co duoc xem khong" — chay trong extension tren may con.
// File script thuong (khong phai module) de dung chung cho content script, service worker va bo kiem
// (worker/tests/kenh.test.js nap file nay roi doc globalThis.LuatKenh).
//
// tt (trang thai tu dich vu tren may, GET http://127.0.0.1:47123/ext/trang-thai):
//   { han_che: bool, kenh: [{ id: 'UC...', handle: '@abc', ten }] }
//   han_che = dang gio chan VA YouTube nam trong danh sach chan. Khong hoi duoc dich vu => coi nhu han_che.
// video (doc tu trinh phat tren trang, hoac tu oEmbed): { id, kenh_id, handle } — co the thieu truong.
(function () {
  const MA_VIDEO = /^[A-Za-z0-9_-]{11}$/;

  function giaiMa(s) { try { return decodeURIComponent(s); } catch { return s; } }
  const chuanHandle = (h) => {
    h = giaiMa(String(h || '')).trim().toLowerCase();
    if (!h) return '';
    return h.startsWith('@') ? h : '@' + h;
  };

  // url => { loai: 'video'|'kenh'|'nhac'|'khac', video_id?, kenh_id?, handle? }
  function loaiTrang(url) {
    let u;
    try { u = new URL(url); } catch { return { loai: 'khac' }; }
    const host = u.hostname.toLowerCase();
    if (host === 'music.youtube.com') return { loai: 'nhac' };
    const phan = u.pathname.split('/').filter(Boolean);
    const dau = phan[0] || '';
    if (host === 'youtu.be') return MA_VIDEO.test(dau) ? { loai: 'video', video_id: dau } : { loai: 'khac' };
    if (dau === 'watch') {
      const v = u.searchParams.get('v') || '';
      return MA_VIDEO.test(v) ? { loai: 'video', video_id: v } : { loai: 'khac' };
    }
    if (['shorts', 'embed', 'live', 'v'].includes(dau) && MA_VIDEO.test(phan[1] || '')) return { loai: 'video', video_id: phan[1] };
    if (dau === 'channel' && /^UC[A-Za-z0-9_-]{22}$/.test(phan[1] || '')) return { loai: 'kenh', kenh_id: phan[1] };
    if (dau.startsWith('@') && dau.length > 1) return { loai: 'kenh', handle: chuanHandle(dau) };
    return { loai: 'khac' };
  }

  // Kenh co trong danh sach bo me chon khong (so ma kenh, hoac @handle khong phan biet hoa thuong).
  function kenhDuocXem(ds, k) {
    if (!k) return false;
    const id = String(k.kenh_id || k.id || '');
    const h = chuanHandle(k.handle);
    return (Array.isArray(ds) ? ds : []).some((x) =>
      (id && x.id === id) || (h && chuanHandle(x.handle) === h));
  }

  // => { ket_qua: 'cho' | 'chan' | 'doi', ly_do }
  //   doi = trang video chua biet kenh => tam dung video, che trang, cho doc them (qua han thi coi nhu chan).
  function quyetDinh(tt, url, video) {
    if (tt && tt.han_che === false) return { ket_qua: 'cho', ly_do: 'Đang giờ được xem YouTube thoải mái' };
    const ds = (tt && tt.kenh) || [];
    const tr = loaiTrang(url);
    if (tr.loai === 'nhac') return { ket_qua: 'chan', ly_do: 'YouTube Music không dùng được trong giờ học' };
    if (tr.loai === 'kenh') {
      return kenhDuocXem(ds, tr) ? { ket_qua: 'cho', ly_do: 'Kênh được bố mẹ cho xem' } : { ket_qua: 'chan', ly_do: 'Kênh này không nằm trong danh sách bố mẹ cho xem' };
    }
    if (tr.loai === 'video') {
      // Thong tin trinh phat con cua video truoc (YouTube doi trang khong tai lai) => chua tinh.
      const v = video && video.id === tr.video_id ? video : null;
      if (v && kenhDuocXem(ds, v)) return { ket_qua: 'cho', ly_do: 'Video của kênh được bố mẹ cho xem' };
      if (!v || (!v.kenh_id && !v.handle)) return { ket_qua: 'doi', ly_do: 'Đang kiểm tra kênh của video…' };
      return { ket_qua: 'chan', ly_do: 'Video này không thuộc kênh bố mẹ cho xem' };
    }
    return { ket_qua: 'chan', ly_do: 'Trong giờ học chỉ xem được các kênh bố mẹ chọn' };
  }

  // oEmbed author_url => { handle } | { kenh_id } | null
  function tacGiaTuOembed(authorUrl) {
    const tr = loaiTrang(String(authorUrl || '').replace(/^http:/, 'https:'));
    if (tr.loai !== 'kenh') return null;
    return tr.kenh_id ? { kenh_id: tr.kenh_id } : { handle: tr.handle };
  }

  globalThis.LuatKenh = { loaiTrang, kenhDuocXem, quyetDinh, tacGiaTuOembed, chuanHandle };
})();
