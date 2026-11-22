# Chính sách riêng tư — extension "Quản lý con — YouTube theo kênh"

Extension này dùng kèm phần mềm "Quản lý máy tính của con" do bố mẹ tự cài trên máy của con.

- **Không thu thập, không lưu, không gửi** dữ liệu cá nhân hay lịch sử duyệt web đi bất cứ đâu.
- Extension chỉ đọc địa chỉ và thông tin kênh của trang YouTube đang mở **ngay trên máy** để quyết định cho xem hay che trang.
- Nó hỏi phần mềm quản lý chạy trên chính máy đó (`http://127.0.0.1:47123`) xem đang giờ học hay giờ giải lao
  và danh sách kênh được xem. Kết nối này không ra khỏi máy.
- Để biết một video thuộc kênh nào, extension có thể gọi `https://www.youtube.com/oembed` với mã video —
  đây là dịch vụ công khai của YouTube, không kèm thông tin đăng nhập.
- Danh sách kênh được nhớ trong bộ nhớ của extension trên máy (`chrome.storage.local`).

Mã nguồn công khai trong thư mục `extension/` của repo.
