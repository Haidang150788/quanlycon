// Content script tren moi trang/khung YouTube (ca video nhung youtube.com/embed, youtube-nocookie.com).
// Moi 400ms: hoi luat (luat-kenh.js) => cho xem | che trang + dung video | doi (dang kiem kenh).
(function () {
  if (window.__qlcDaChay) return;
  window.__qlcDaChay = true;
  const L = globalThis.LuatKenh;
  const CHO_TOI_DA_MS = 8000; // qua lau van chua biet kenh => chan
  const KHUNG_TRONG = window.top !== window;

  let tt = null;
  let ttLuc = 0;
  let trinhPhat = null; // {id, kenh_id, tac_gia} tu trinh-phat.js
  const oembed = new Map(); // ma video => {kenh_id}|{handle}|null
  let urlCu = '';
  let doiTu = 0;
  let dangChan = false;

  function hoiNen(m) {
    return new Promise((ok) => {
      try { chrome.runtime.sendMessage(m, (r) => { void chrome.runtime.lastError; ok(r); }); } catch { ok(undefined); }
    });
  }
  async function taiTT() {
    const r = await hoiNen({ hoi: 'tt' });
    tt = r || { han_che: true, kenh: [], ok: false };
    ttLuc = Date.now();
  }

  window.addEventListener('message', (e) => {
    if (e.source === window && e.data && e.data.__qlc_trinh_phat) trinhPhat = e.data.__qlc_trinh_phat;
  });

  // Video dang bi chan ma con bam phat => dung ngay (khong doi vong 400ms).
  // Danh dau data-qlc-dung: kiem xong ma duoc xem thi tu phat tiep (YouTube tu phat luc dang "kiem kenh").
  document.addEventListener('play', (e) => {
    if (dangChan && e.target && e.target.pause) { e.target.setAttribute('data-qlc-dung', '1'); e.target.pause(); }
  }, true);

  function dungVideo() {
    document.querySelectorAll('video, audio').forEach((v) => { try { if (!v.paused) { v.setAttribute('data-qlc-dung', '1'); v.pause(); } if (!v.muted) { v.setAttribute('data-qlc-tat-tieng', '1'); v.muted = true; } } catch { /* */ } });
  }
  function moTiengLai() {
    document.querySelectorAll('video[data-qlc-tat-tieng]').forEach((v) => { v.muted = false; v.removeAttribute('data-qlc-tat-tieng'); });
    document.querySelectorAll('video[data-qlc-dung]').forEach((v) => { v.removeAttribute('data-qlc-dung'); const r = v.play(); if (r && r.catch) r.catch(() => {}); });
  }

  // ---------- lop che ----------
  let vo = null;
  let goc = null;
  function veChe(kq) {
    if (!vo) {
      vo = document.createElement('qlc-che');
      goc = vo.attachShadow({ mode: 'closed' });
      goc.innerHTML = `<style>
        :host{all:initial}
        .nen{position:fixed;inset:0;z-index:2147483647;background:#0f2233;color:#fff;display:flex;align-items:center;justify-content:center;
          font:16px/1.5 "Segoe UI",system-ui,sans-serif;padding:24px;box-sizing:border-box;overflow:auto}
        .hop{max-width:560px;width:100%}
        h1{font-size:24px;margin:0 0 6px} p{margin:0 0 16px;color:#c7d3de}
        .ds{display:grid;gap:8px} a{display:block;background:#fff;color:#13202c;text-decoration:none;border-radius:12px;padding:12px 14px;font-weight:600}
        a small{display:block;font-weight:400;color:#66727e;font-size:13px}
        .nho h1{font-size:16px}.nho p{font-size:13px;margin:0}
      </style><div class="nen"><div class="hop"><h1></h1><p></p><div class="ds"></div></div></div>`;
    }
    if (!vo.isConnected) (document.documentElement || document).appendChild(vo);
    const h = goc.querySelector('.hop');
    h.classList.toggle('nho', KHUNG_TRONG);
    h.querySelector('h1').textContent = kq.ket_qua === 'doi' ? '⏳ Đang kiểm tra kênh…' : '🔒 Giờ học — YouTube chỉ mở các kênh bố mẹ chọn';
    h.querySelector('p').textContent = kq.ly_do + (tt && tt.ok === false && kq.ket_qua !== 'doi' ? ' (không kết nối được phần mềm quản lý trên máy nên tạm chặn)' : '');
    const ds = h.querySelector('.ds');
    const khoa = JSON.stringify((tt && tt.kenh) || []) + kq.ket_qua;
    if (ds.dataset.khoa === khoa) return;
    ds.dataset.khoa = khoa;
    ds.textContent = '';
    if (kq.ket_qua === 'doi' || KHUNG_TRONG) return;
    const kenh = (tt && tt.kenh) || [];
    if (!kenh.length) { const p = document.createElement('p'); p.textContent = 'Bố mẹ chưa chọn kênh nào.'; ds.appendChild(p); return; }
    for (const k of kenh) {
      const a = document.createElement('a');
      a.href = 'https://www.youtube.com/' + (k.handle ? encodeURIComponent(k.handle).replace(/^%40/, '@') : 'channel/' + k.id);
      a.textContent = k.ten || k.handle || k.id;
      if (k.handle) { const s = document.createElement('small'); s.textContent = k.handle; a.appendChild(s); }
      ds.appendChild(a);
    }
  }
  function boChe() { if (vo && vo.isConnected) vo.remove(); }

  // Trong gio chan, tren trang duoc xem: an de xuat / tim kiem / Shorts de con khong bam sang kenh khac.
  const CSS_AN = `#related, ytd-watch-next-secondary-results-renderer, .ytp-endscreen-content, .ytp-ce-element, .ytp-videowall-still,
    .ytp-autonav-endscreen-countdown-overlay, .ytp-pause-overlay, ytd-reel-shelf-renderer, ytd-rich-shelf-renderer[is-shorts],
    ytd-searchbox, #masthead #center, yt-searchbox, ytd-mini-guide-renderer, #guide-button, tp-yt-app-drawer#guide,
    ytm-item-section-renderer[section-identifier="related-items"], ytm-searchbox, .ytp-show-cards-title { display: none !important; }`;
  let theAn = null;
  function anDeXuat(bat) {
    if (bat && !theAn) { theAn = document.createElement('style'); theAn.textContent = CSS_AN; }
    if (bat && !theAn.isConnected) (document.head || document.documentElement).appendChild(theAn);
    if (!bat && theAn && theAn.isConnected) theAn.remove();
  }

  function thongTinVideo(idTrang) {
    if (!idTrang) return null;
    const v = { id: idTrang, kenh_id: '', handle: '' };
    if (trinhPhat && trinhPhat.id === idTrang) v.kenh_id = trinhPhat.kenh_id;
    const o = oembed.get(idTrang);
    if (o === undefined) {
      oembed.set(idTrang, null);
      hoiNen({ hoi: 'oembed', id: idTrang }).then((r) => { oembed.set(idTrang, r || null); });
    } else if (o) {
      if (o.handle) v.handle = o.handle;
      if (o.kenh_id && !v.kenh_id) v.kenh_id = o.kenh_id;
    }
    return v;
  }

  function vong() {
    if (!tt || Date.now() - ttLuc > 15e3) taiTT();
    const url = location.href;
    if (url !== urlCu) { urlCu = url; doiTu = 0; }
    const tr = L.loaiTrang(url);
    let kq = tt ? L.quyetDinh(tt, url, thongTinVideo(tr.video_id)) : { ket_qua: 'doi', ly_do: 'Đang hỏi phần mềm trên máy…' };
    if (kq.ket_qua === 'doi') {
      if (!doiTu) doiTu = Date.now();
      // Co thong tin kenh roi ma van "doi" thi khong xay ra; day la truong hop khong doc duoc kenh.
      if (Date.now() - doiTu > CHO_TOI_DA_MS) kq = { ket_qua: 'chan', ly_do: 'Không kiểm tra được kênh của video này nên tạm chặn' };
    } else doiTu = 0;

    const chan = kq.ket_qua !== 'cho';
    if (chan) {
      dungVideo();
      veChe(kq);
    } else {
      if (dangChan) moTiengLai();
      boChe();
    }
    dangChan = chan;
    anDeXuat(!!(tt && tt.han_che));
  }

  taiTT().then(vong);
  setInterval(vong, 400);
  document.addEventListener('yt-navigate-finish', vong);
})();
