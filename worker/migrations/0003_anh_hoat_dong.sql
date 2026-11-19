-- Anh chup man hinh may con (JPEG ~100-250KB). Giu 7 ngay, xoa dan moi lan ghi.
CREATE TABLE IF NOT EXISTS anh (
  id INTEGER PRIMARY KEY AUTOINCREMENT, ts INTEGER NOT NULL, may TEXT, ly_do TEXT,
  kich_thuoc INTEGER, cua_so TEXT, jpeg BLOB NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_anh_ts ON anh(ts);
-- Cua so dang dung tren may con: ghi khi doi cua so hoac moi 2 phut. Giu 30 ngay.
CREATE TABLE IF NOT EXISTS hoat_dong (
  id INTEGER PRIMARY KEY AUTOINCREMENT, ts INTEGER NOT NULL, may TEXT, exe TEXT, tieu_de TEXT, ranh INTEGER
);
CREATE INDEX IF NOT EXISTS idx_hoat_dong_ts ON hoat_dong(ts);
