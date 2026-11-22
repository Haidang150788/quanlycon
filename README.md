# Quản lý máy tính của con

Phần mềm tự dựng để bố mẹ quản lý máy tính **Windows** của con từ điện thoại:
chặn mạng xã hội / game theo giờ, cho chơi thêm 30 phút – 1 tiếng, khoá màn hình từ xa,
con bấm "Xin thêm giờ" thì điện thoại báo, chặn cài phần mềm lạ, xem cửa sổ đang mở và ảnh chụp màn hình.

Mọi dữ liệu nằm trên **tài khoản Cloudflare của chính bạn** (gói miễn phí là đủ cho một gia đình).
Không có máy chủ trung gian nào khác.

```
Điện thoại bố mẹ ── web (Cloudflare Worker + cơ sở dữ liệu D1)   ← thư mục worker/
                          ▲  máy con hỏi mỗi 30 giây
Máy con Windows ── dịch vụ nền "QuanLyCon" (quyền SYSTEM)       ← thư mục agent/ (Go)
```

## Điều kiện bắt buộc

- Máy con chạy Windows 10/11 (Home cũng được).
- **Tài khoản của con phải là tài khoản thường (Standard User)**, không phải Admin.
  Có quyền Admin thì con tự tắt được phần mềm. Kiểm: Settings → Accounts → Other users.
- Bạn có tài khoản Cloudflare (miễn phí) và máy tính cài [Node.js](https://nodejs.org) 18 trở lên.
- Chỉ cần [Go](https://go.dev/dl/) nếu muốn tự dựng file cài (xem bước 3).

## Bước 1 — Dựng server (khoảng 10 phút)

```bash
git clone <địa chỉ repo này> quanlycon
cd quanlycon/worker
npx wrangler login                       # mở trình duyệt đăng nhập Cloudflare
npx wrangler d1 create quanlycon-db      # in ra database_id
cp wrangler.example.toml wrangler.toml   # sửa database_id và VAPID_SUB (2 chỗ đánh dấu <<<)
npx wrangler d1 migrations apply quanlycon-db --remote
node ../scripts/tao-bi-mat.mjs           # in ra AGENT_TOKEN + VAPID_JWK — cất kỹ
npx wrangler secret put PIN              # mật khẩu bạn tự đặt để vào web
npx wrangler secret put AGENT_TOKEN
npx wrangler secret put VAPID_JWK
npx wrangler deploy                      # in ra địa chỉ web, vd https://quanlycon.<tên>.workers.dev
```

Mở địa chỉ đó trên điện thoại, nhập PIN. Trên iPhone: Chia sẻ → **Thêm vào MH chính**, mở từ biểu tượng
đó rồi bấm "Bật thông báo" (iPhone chỉ nhận thông báo khi mở từ màn hình chính).

Muốn xem thử giao diện trước khi làm: mở `<địa chỉ web>/?demo` (dữ liệu giả, không cần PIN).

## Bước 2 — Cài lên máy con

**Cách dễ:** tải `quanlycon.exe` ở mục Releases của repo, chép sang máy con, bấm đúp →
Windows hỏi "Windows protected your PC" thì bấm *More info → Run anyway* (file chưa ký số) →
bấm **Yes** và nhập mật khẩu Admin → dán địa chỉ server và AGENT_TOKEN → chờ dòng **CÀI XONG**.

Cài trên tài khoản của con: nếu máy **hỏi mật khẩu Admin** là đúng tài khoản thường.
Chỉ hiện nút Yes mà không hỏi mật khẩu ⇒ tài khoản con đang là Admin, sửa trước khi cài.

Cài xong, mở lại web: dòng "Chưa có máy nào kết nối" phải biến mất. Thử mở Facebook trên máy con để chắc là bị chặn.

Lệnh khác (cmd chạy *Run as administrator*): `quanlycon.exe cai <token> <server>` (cài lại / nâng cấp),
`quanlycon.exe go` (gỡ sạch, trả máy về như cũ), `quanlycon.exe xem` (xem trạng thái).

## Bước 3 (tuỳ chọn) — Tự dựng file cài

Tạo file `cau-hinh.env` ở thư mục gốc (đã nằm trong `.gitignore`):

```
SERVER=https://quanlycon.<tên>.workers.dev
CAU_HOI="Hỏi bố mẹ nếu muốn chơi."      # câu hiện cho con khi mở game bị chặn
TOKEN=<AGENT_TOKEN>                      # có dòng này thì ra thêm file bấm đúp là cài, không phải dán gì
```

rồi chạy `scripts/build-exe.sh`. Ra `dist/quanlycon.exe` và `dist/CaiDat-QuanLyCon.exe`.
**File CaiDat chứa mã của bạn — đừng gửi công khai, cài xong thì xoá.**
(Bản chép vào Program Files tự bị xoá mã, mã chỉ còn trong thư mục chỉ Admin đọc được.)

## Bước 4 (tuỳ chọn) — YouTube chỉ xem các kênh bạn chọn

Mặc định, giờ chặn là chặn hẳn YouTube. Muốn trong giờ chặn con vẫn xem được vài kênh học (giờ giải lao vẫn xem thoải mái):
đăng extension trong thư mục `extension/` lên Chrome Web Store (phí 5 USD một lần, để chế độ *Unlisted*),
ghi mã của nó vào `EXT_CHROME` trong `worker/wrangler.toml` rồi deploy. Máy con (bản ≥ 1.4.0) tự ép cài extension vào Chrome/Edge.
Các bước chi tiết: [`extension/HUONG-DAN-DANG.md`](extension/HUONG-DAN-DANG.md).

Sau đó thêm kênh trên web điện thoại, mục **YouTube: kênh được xem trong giờ học**: dán link kênh
(`youtube.com/@ten`, `/channel/UC…`) hoặc link một video của kênh đó.

Nếu extension chưa chạy trong trình duyệt nào thì giờ chặn trình duyệt đó vẫn chặn hẳn YouTube. Máy con chỉ mở YouTube
cho trình duyệt nào mà extension vừa báo về trong 10 phút gần nhất.

## Cách chặn

1. Chính sách Chrome/Edge (`URLBlocklist`), chặn cả tên miền con.
2. File `hosts` — lớp dự phòng cho app khác.
3. Luôn bật: tắt DNS-over-HTTPS, tắt ẩn danh/InPrivate, tắt tài khoản Guest (không thì lách được).
4. Game/app: quét tiến trình mỗi 2 giây, khớp tên `.exe` thì đóng và hiện câu nhắn cho con.
5. Khoá cài: phần mềm mới chạy ngoài `Windows`/`Program Files` bị tạm dừng, bố mẹ duyệt trên điện thoại (quá 2 phút thì đóng).

6. YouTube theo kênh (nếu bật bước 4): extension xem video đang mở thuộc kênh nào (trình phát + oEmbed). Kênh không có trong danh sách thì dừng video và che trang.
   Trong giờ chặn còn ẩn đề xuất, tìm kiếm, Shorts; YouTube Music và các trang xem YouTube "thay thế" (Invidious, Piped…) luôn bị chặn.
   Chặn được cả video nhúng trên trang khác (`youtube.com/embed`, `youtube-nocookie.com`). Đã mở YouTube cho Chrome/Edge thì trong giờ chặn
   máy con đóng các trình duyệt khác (Firefox, Cốc Cốc, Brave, Opera…), vì chúng không có extension. Extension bị ép cài thì con không gỡ hay tắt được.
   Máy con cũng tắt DevTools và chế độ nhà phát triển của trình duyệt.

Giờ lấy từ server, con đổi đồng hồ Windows không lách được.
Danh sách nền tảng (tên miền + tên .exe) nằm ở `worker/src/danhmuc.js` — thêm vào rồi `npx wrangler deploy`,
không phải cài lại máy con. Luật giờ có **hai bản phải sửa cùng lúc**: `worker/src/luat.js` ↔ `agent/lich.go`.

## Lỗ đã biết

- Trang web bị chặn hiện trang lỗi của Chrome, không thêm chữ riêng được.
- Phần mềm `.msi` cài theo người dùng không bị khoá cài bắt. Tiến trình lạ chạy được ~1 giây trước khi bị dừng.
- Con chép game vào thư mục một app đã được duyệt thì lọt.
- Trình duyệt khác Chrome/Edge (Firefox, Cốc Cốc…) chỉ chặn bằng file hosts — nên thêm chúng vào danh sách app bị chặn.
- YouTube theo kênh: extension từ Chrome Web Store mà con tự cài (vd. trình chạy userscript) có thể gỡ lớp che; công cụ dòng lệnh (`curl`)
  giả được tín hiệu "extension đang chạy" (chỉ có hại khi extension thật không chạy); trình duyệt lạ dạng portable phải nhờ khoá cài bắt.
  Thẻ gợi ý video bên trong trình phát vẫn hiện, nhưng bấm vào video kênh khác thì vẫn bị chặn.

## Kiểm thử

```bash
cd worker && npm test          # gồm luật extension/luat-kenh.js
cd agent && go test ./... && GOOS=windows GOARCH=amd64 go vet ./...
```

Phần mềm còn mới (dựng 11/2026, chạy thật trên một máy). Dùng cho gia đình mình, tự chịu trách nhiệm.
