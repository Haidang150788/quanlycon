-- Kenh YouTube con duoc xem trong gio chan (khi YouTube nam trong danh sach chan).
-- id = ma kenh 'UC...' (khong doi), handle '@...' (co the doi) de so khop du phong.
CREATE TABLE IF NOT EXISTS kenh_yt (id TEXT PRIMARY KEY, ten TEXT, handle TEXT, ts INTEGER NOT NULL);
