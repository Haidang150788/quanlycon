# Extension "YouTube theo kênh" — đăng lên kho và bật

Trong giờ chặn, YouTube trên máy con chỉ mở được video của các kênh bạn chọn trên web điện thoại.
Giờ giải lao thì xem thoải mái. Không cài extension thì mọi thứ vẫn như cũ: giờ chặn là chặn hẳn YouTube.

## Vì sao phải đăng lên kho

Chrome và Edge trên Windows **không vào domain công ty** chỉ cho chính sách ép cài
(`ExtensionInstallForcelist`) những extension có trên kho chính hãng: Chrome Web Store cho Chrome,
Edge Add-ons cho Edge. Extension tự để trên máy hay trên server riêng sẽ bị trình duyệt bỏ qua.
Để chế độ **Không công khai (Unlisted / Hidden)** thì người ngoài không tìm thấy, chỉ ai có link mới thấy.

## Bước 1 — Đóng gói

```bash
scripts/dong-goi-extension.sh        # chạy bộ kiểm luật rồi ra dist/quanlycon-youtube-<phiên bản>.zip
```

## Bước 2a — Chrome Web Store (phí 5 USD một lần)

1. Mở <https://chrome.google.com/webstore/devconsole>, đăng nhập một tài khoản Google
   (cần bật xác minh 2 bước), đồng ý điều khoản, trả phí đăng ký 5 USD.
2. **Thêm mặt hàng mới** → tải file `.zip` ở bước 1 lên.
3. Thẻ **Store listing**: ngôn ngữ Tiếng Việt, danh mục *Family* (hoặc *Tools*), mô tả lấy dòng `description`
   trong `manifest.json`. Ảnh: `extension/cua-hang/anh-1280x800.png` (ảnh chụp màn hình) và
   `extension/cua-hang/o-quang-cao-440x280.png` (ô quảng cáo nhỏ).
4. Thẻ **Privacy practices**:
   - *Single purpose*: "Trong giờ học chỉ cho xem video của các kênh YouTube do bố mẹ chọn."
   - `alarms`: hỏi phần mềm quản lý trên máy mỗi 30 giây xem đang giờ học hay giờ giải lao.
   - `storage`: nhớ danh sách kênh để vẫn chặn đúng khi phần mềm trên máy tạm không trả lời.
   - Host `http://127.0.0.1:47123/*`: phần mềm quản lý chạy trên chính máy đó, không đi ra mạng ngoài.
   - Host `https://www.youtube.com/*`: tra kênh của một video (oEmbed) và chặn/cho xem trên trang YouTube.
   - Dữ liệu: không thu thập, không gửi đi đâu. Đánh dấu ba ô cam kết.
   - *Privacy policy URL*: link tới file `extension/RIENG-TU.md` trên GitHub của repo bạn.
5. Thẻ **Distribution**: *Visibility* = **Unlisted**.
6. **Submit for review**. Thường duyệt xong trong 1–3 ngày, có khi lâu hơn.
7. Duyệt xong, mã extension (32 chữ a–p) nằm trong địa chỉ trang của nó và trong trang devconsole.

## Bước 2b — Edge (tuỳ chọn, miễn phí)

Cách nhanh: thử cho Edge dùng luôn bản trên Chrome Web Store (bước 3, dòng `EXT_EDGE` có `;https://clients2...`).
Edge không nhận cũng không sao: máy con không nghe thấy extension trong Edge thì Edge vẫn chặn hẳn YouTube như trước.

Cách chính thức: đăng ký miễn phí tại <https://partner.microsoft.com/dashboard/microsoftedge>
bằng tài khoản Microsoft, tải cùng file `.zip`, *Visibility* = **Hidden**, gửi duyệt (tới khoảng 7 ngày làm việc).
Edge Add-ons cấp một mã **khác** mã bên Chrome.

## Bước 3 — Báo mã cho server

Sửa `worker/wrangler.toml`, mục `[vars]`:

```toml
EXT_CHROME = "<mã trên Chrome Web Store>"
EXT_EDGE = "<mã trên Edge Add-ons>"
# hoặc Edge dùng chung bản Chrome Web Store:
# EXT_EDGE = "<mã Chrome>;https://clients2.google.com/service/update2/crx"
```

rồi `cd worker && npm test && npx wrangler deploy`. Máy con phải chạy phần mềm **≥ 1.4.0**.

## Bước 4 — Kiểm trên máy con

1. Đợi 1–2 phút (máy con hỏi server mỗi 30 giây rồi ghi chính sách). Mở `chrome://policy` →
   bấm *Reload policies*: phải thấy `ExtensionInstallForcelist` có mã ở trên, trạng thái OK.
2. `chrome://extensions`: extension có biểu tượng toà nhà (do tổ chức quản lý), không có nút gỡ/tắt.
3. Trên web điện thoại, dòng tên máy phải hiện `Extension YouTube: Chrome ✓` (Edge ✓ nếu có).
4. Trong giờ chặn: mở một video của kênh đã chọn → xem được; mở trang chủ YouTube → hiện danh sách kênh được xem.

## Cập nhật extension

Tăng `version` trong `manifest.json`, đóng gói lại, tải lên mục cũ trên devconsole (*Package → Upload new package*).
Duyệt xong, trình duyệt máy con tự cập nhật trong vài giờ.
