// BAN XEM THU: <dia-chi-server>/?demo — du lieu gia, khong goi server, khong can PIN.
// Sinh boi lenh trong CLAUDE.md muc 'Ban xem thu'; danh_muc chep tu src/danhmuc.js.
(function () {
  const now = Date.now(), p = 60000;
  const dm = [{"ma":"facebook","ten":"Facebook","nhom":"mxh","mau":"#1877f2","chu":"f"},{"ma":"tiktok","ten":"TikTok","nhom":"mxh","mau":"#111","chu":"Tt"},{"ma":"youtube","ten":"YouTube","nhom":"mxh","mau":"#ff0000","chu":"▶"},{"ma":"instagram","ten":"Instagram","nhom":"mxh","mau":"#e1306c","chu":"Ig"},{"ma":"threads","ten":"Threads","nhom":"mxh","mau":"#111","chu":"@"},{"ma":"x","ten":"X (Twitter)","nhom":"mxh","mau":"#111","chu":"X"},{"ma":"telegram","ten":"Telegram","nhom":"mxh","mau":"#27a7e7","chu":"Tg"},{"ma":"discord","ten":"Discord","nhom":"mxh","mau":"#5865f2","chu":"Dc"},{"ma":"roblox","ten":"Roblox","nhom":"game","mau":"#393b3d","chu":"Rb"},{"ma":"minecraft","ten":"Minecraft","nhom":"game","mau":"#5b8731","chu":"Mc"},{"ma":"garena","ten":"Garena","nhom":"game","mau":"#ee4d2d","chu":"G"},{"ma":"riot","ten":"Riot (LMHT, Valorant)","nhom":"game","mau":"#d13639","chu":"R"},{"ma":"steam","ten":"Steam / Epic","nhom":"game","mau":"#1b2838","chu":"St"},{"ma":"gameweb","ten":"Game trên web","nhom":"game","mau":"#7b4fd6","chu":"Gw"}];
  const S = {
    cau_hinh: { bat: true, nen_tang: dm.map((x) => x.ma).filter((m) => m !== 'minecraft'), web_khac: ['shopee.vn'],
      giai_lao: [{ tu: '11:30', den: '12:30' }, { tu: '19:30', den: '20:30' }], dong_trinh_duyet: true, khoa_cai: true, chup_phut: 10, chup_luc: 0, mo_tam_den: 0, khoa_luc: 0 },
    dang_chan: true, danh_muc: dm, nhom: [{"ma":"mxh","ten":"Mạng xã hội"},{"ma":"game","ten":"Game"}],
    may: [{ id: 'may-cua-con', ten: 'MAY-CUA-CON', phien_ban: '1.4.0', lan_cuoi: now - 12000, dang_chan: 1, ghi_chu: 'Extension YouTube: Chrome ✓, Edge ✗' }],
    yeu_cau: [
      { id: 'demo-gio', ts: now - 2 * p, loai: 'gio', phut: 30, ly_do: 'con làm nốt bài tập nhóm trên Zalo ạ' },
      { id: 'demo-chay', ts: now - 20000, loai: 'chay', ten: 'FreeVPN_Setup.exe', duong_dan: 'C:\\Users\\Con\\Downloads\\FreeVPN_Setup.exe', kich_thuoc: 18874368 },
    ],
    quy_tac: [
      { id: 1, ts: now - 864e5, loai: 'thu_muc', gia_tri: 'c:\\users\\con\\appdata\\local\\zoom\\', ten: 'Zoom.exe', cho_phep: 1 },
      { id: 2, ts: now - 3600e3, loai: 'hash', gia_tri: 'ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12cd34ef56ab12', ten: 'MinecraftCrack.exe', cho_phep: 0 },
    ],
    nhat_ky: [
      { ts: now - 20000, loai: 'xin_chay', noi_dung: 'Con mở phần mềm lạ FreeVPN_Setup.exe (18.0 MB), đang chờ duyệt' },
      { ts: now - 2 * p, loai: 'xin_gio', noi_dung: 'Con xin thêm 30 phút: con làm nốt bài tập nhóm trên Zalo ạ' },
      { ts: now - 25 * p, loai: 'dong', noi_dung: 'Đã đóng robloxplayerbeta.exe' },
      { ts: now - 3 * 3600e3, loai: 'ap_dung', noi_dung: 'Bắt đầu chặn 58 tên miền, 30 ứng dụng' },
      { ts: now - 4 * 3600e3, loai: 'khoi_dong', noi_dung: 'Phần mềm khởi động, phiên bản 1.4.0' },
      { ts: now - 864e5, loai: 'kenh', noi_dung: 'Bố mẹ cho xem kênh YouTube TED-Ed' },
    ],
    han_chay_ms: 120000, push_khoa: null, so_thiet_bi_bao: 0,
    co_extension: true,
    kenh: [
      { id: 'UC4a-Gbdw7vOaccHmFo40b9g', ten: 'Khan Academy', handle: '@khanacademy', ts: now - 3 * 864e5 },
      { id: 'UCsooa4yRKGN_zEE8iknghZA', ten: 'TED-Ed', handle: '@TEDEd', ts: now - 864e5 },
    ],
  };
  const anh = [
    { id: 1, ts: now - 18 * p, ly_do: 'định kỳ', kich_thuoc: 78456, cua_so: 'chrome.exe — cách vượt chặn web trên máy tính - Tìm trên Google - Google Chrome' },
    { id: 3, ts: now - 25 * p, ly_do: 'Con mở robloxplayerbeta.exe (đang bị chặn)', kich_thuoc: 68628, cua_so: 'RobloxPlayerBeta.exe — Roblox' },
    { id: 2, ts: now - 40 * p, ly_do: 'định kỳ', kich_thuoc: 94516, cua_so: 'WINWORD.EXE — Bai tap van.docx - Word' },
  ];
  const hd = [];
  const them = (phutTruoc, exe, td, dai) => { for (let i = 0; i < dai; i += 2) hd.push({ ts: now - (phutTruoc - i) * p, exe, tieu_de: td, ranh: 4 }); };
  them(95, 'chrome.exe', 'Bài 6 - Hàm số - Toán lớp 7 - Google Chrome', 20);
  them(74, 'WINWORD.EXE', 'Bai tap van.docx - Word', 30);
  them(42, 'Zalo.exe', 'Zalo - Nhóm 7A3', 12);
  them(28, 'chrome.exe', 'cách vượt chặn web trên máy tính - Tìm trên Google - Google Chrome', 6);
  them(21, 'chrome.exe', 'Free VPN Proxy - Chrome Web Store - Google Chrome', 4);
  them(16, 'chrome.exe', 'Bài 7 - Toán lớp 7 - Google Chrome', 14);
  window.DEMO_API = async (duong, body) => {
    await new Promise((r) => setTimeout(r, 120));
    S.now = Date.now();
    if (duong === 'trang-thai') return S;
    if (duong.startsWith('anh')) return { anh };
    if (duong.startsWith('hoat-dong')) return { hoat_dong: hd };
    if (duong.startsWith('yeu-cau/')) { S.yeu_cau = S.yeu_cau.filter((y) => !duong.endsWith(y.id)); return S; }
    if (duong === 'cau-hinh') { Object.assign(S.cau_hinh, body); return S; }
    if (duong === 'mo-tam') { S.cau_hinh.mo_tam_den = body.phut ? Date.now() + body.phut * p : 0; S.dang_chan = !body.phut; return S; }
    if (duong === 'kenh/xoa') { S.kenh = S.kenh.filter((k) => k.id !== body.id); return S; }
    if (duong === 'kenh/them') {
      // Ban that hoi YouTube; ban xem thu chi lay chu sau @ trong link.
      const m = /@([\w.-]+)/.exec(body.link || '');
      if (!m) throw new Error('Bản xem thử chỉ nhận link dạng youtube.com/@ten');
      if (!S.kenh.some((k) => k.handle === '@' + m[1])) S.kenh.push({ id: 'UCdemo' + Math.random().toString(36).slice(2, 20).padEnd(18, 'x'), ten: m[1], handle: '@' + m[1], ts: Date.now() });
      return S;
    }
    if (duong === 'quy-tac/xoa') { S.quy_tac = S.quy_tac.filter((q) => q.id !== body.id); return S; }
    return S;
  };
})();
