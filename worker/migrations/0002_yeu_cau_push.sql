-- Ban 2 + 3: yeu cau tu may con (xin them gio / xin chay phan mem la), quy tac duyet, dang ky thong bao.
CREATE TABLE IF NOT EXISTS yeu_cau (
  id TEXT PRIMARY KEY, ts INTEGER NOT NULL, may TEXT, loai TEXT NOT NULL,   -- 'gio' | 'chay'
  trang_thai TEXT NOT NULL DEFAULT 'cho',                                     -- cho | dong_y | tu_choi | het_han
  phut INTEGER, ly_do TEXT, ten TEXT, duong_dan TEXT, sha256 TEXT, kich_thuoc INTEGER,
  xu_ly_luc INTEGER, xu_ly_phut INTEGER
);
CREATE INDEX IF NOT EXISTS idx_yeu_cau_ts ON yeu_cau(ts);
-- loai 'thu_muc' (tien to duong dan, chu thuong) | 'hash' (sha256). cho_phep 1 = cho, 0 = cam.
CREATE TABLE IF NOT EXISTS quy_tac (
  id INTEGER PRIMARY KEY AUTOINCREMENT, ts INTEGER NOT NULL, loai TEXT NOT NULL, gia_tri TEXT NOT NULL,
  ten TEXT, cho_phep INTEGER NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_quy_tac ON quy_tac(loai, gia_tri);
CREATE TABLE IF NOT EXISTS push_sub (endpoint TEXT PRIMARY KEY, p256dh TEXT NOT NULL, auth TEXT NOT NULL, ts INTEGER, ten TEXT);
