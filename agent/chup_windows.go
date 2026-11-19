//go:build windows

package main

// Ban 4 — phan Windows API cua chup man hinh.
//   chayTrongPhien: dich vu (SYSTEM, phien 0) chay `quanlycon.exe <lenh>` trong phien cua nguoi dang ngoi may
//                   bang token cua ho (WTSQueryUserToken + CreateProcessAsUser), doc stdout qua ong dan.
//   chayNhin:       chay trong phien cua con — chup man hinh bang GDI, doc cua so dang dung, so giay ranh.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	user32                  = windows.NewLazySystemDLL("user32.dll")
	gdi32                   = windows.NewLazySystemDLL("gdi32.dll")
	kernel32                = windows.NewLazySystemDLL("kernel32.dll")
	pGetDC                  = user32.NewProc("GetDC")
	pReleaseDC              = user32.NewProc("ReleaseDC")
	pGetSystemMetrics       = user32.NewProc("GetSystemMetrics")
	pGetForegroundWindow    = user32.NewProc("GetForegroundWindow")
	pGetWindowTextW         = user32.NewProc("GetWindowTextW")
	pGetWindowThreadProcId  = user32.NewProc("GetWindowThreadProcessId")
	pGetLastInputInfo       = user32.NewProc("GetLastInputInfo")
	pSetProcessDPIAware     = user32.NewProc("SetProcessDPIAware")
	pCreateCompatibleDC     = gdi32.NewProc("CreateCompatibleDC")
	pCreateCompatibleBitmap = gdi32.NewProc("CreateCompatibleBitmap")
	pSelectObject           = gdi32.NewProc("SelectObject")
	pBitBlt                 = gdi32.NewProc("BitBlt")
	pGetDIBits              = gdi32.NewProc("GetDIBits")
	pDeleteObject           = gdi32.NewProc("DeleteObject")
	pDeleteDC               = gdi32.NewProc("DeleteDC")
	pGetTickCount           = kernel32.NewProc("GetTickCount")
)

// ---------- phia dich vu ----------
func chayTrongPhien(lenh string, gioiHan time.Duration) ([]byte, error) {
	sid := windows.WTSGetActiveConsoleSessionId()
	if sid == 0xFFFFFFFF {
		return nil, fmt.Errorf("không có phiên nào trên màn hình")
	}
	var tok windows.Token
	if err := windows.WTSQueryUserToken(sid, &tok); err != nil {
		return nil, fmt.Errorf("chưa ai đăng nhập: %w", err)
	}
	defer tok.Close()
	var env *uint16
	if windows.CreateEnvironmentBlock(&env, tok, false) == nil {
		defer windows.DestroyEnvironmentBlock(env)
	}
	sa := windows.SecurityAttributes{Length: uint32(unsafe.Sizeof(windows.SecurityAttributes{})), InheritHandle: 1}
	var r, w windows.Handle
	if err := windows.CreatePipe(&r, &w, &sa, 0); err != nil {
		return nil, err
	}
	_ = windows.SetHandleInformation(r, windows.HANDLE_FLAG_INHERIT, 0)
	desk, _ := windows.UTF16PtrFromString(`winsta0\default`)
	si := windows.StartupInfo{
		Cb: uint32(unsafe.Sizeof(windows.StartupInfo{})), Desktop: desk,
		Flags: windows.STARTF_USESTDHANDLES | windows.STARTF_USESHOWWINDOW, ShowWindow: windows.SW_HIDE,
		StdOutput: w, StdErr: w,
	}
	exe, _ := os.Executable()
	cmd, _ := windows.UTF16PtrFromString(`"` + exe + `" ` + lenh)
	thuMucChay, _ := windows.UTF16PtrFromString(filepath.Dir(exe))
	var pi windows.ProcessInformation
	err := windows.CreateProcessAsUser(tok, nil, cmd, nil, nil, true,
		windows.CREATE_NO_WINDOW|windows.CREATE_UNICODE_ENVIRONMENT, env, thuMucChay, &si, &pi)
	windows.CloseHandle(w)
	if err != nil {
		windows.CloseHandle(r)
		return nil, fmt.Errorf("CreateProcessAsUser: %w", err)
	}
	defer windows.CloseHandle(pi.Process)
	defer windows.CloseHandle(pi.Thread)
	f := os.NewFile(uintptr(r), "ong-nhin")
	defer f.Close()
	type kq struct {
		b   []byte
		err error
	}
	c := make(chan kq, 1)
	go func() { b, err := io.ReadAll(io.LimitReader(f, 8<<20)); c <- kq{b, err} }()
	select {
	case k := <-c:
		windows.WaitForSingleObject(pi.Process, 5000)
		return k.b, k.err
	case <-time.After(gioiHan):
		_ = windows.TerminateProcess(pi.Process, 1)
		return nil, fmt.Errorf("quá %v", gioiHan)
	}
}

// ---------- phia phien cua con ----------
type bitmapInfoHeader struct {
	Size          uint32
	Width         int32
	Height        int32
	Planes        uint16
	BitCount      uint16
	Compression   uint32
	SizeImage     uint32
	XPelsPerMeter int32
	YPelsPerMeter int32
	ClrUsed       uint32
	ClrImportant  uint32
}

type bitmapInfo struct {
	Header bitmapInfoHeader
	Colors [1]uint32
}

func cuaSoDangDung() (exe, tieuDe string) {
	hwnd, _, _ := pGetForegroundWindow.Call()
	if hwnd == 0 {
		return "", ""
	}
	buf := make([]uint16, 512)
	n, _, _ := pGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 512)
	tieuDe = windows.UTF16ToString(buf[:n])
	var pid uint32
	pGetWindowThreadProcId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid); err == nil {
		p := make([]uint16, 1024)
		sz := uint32(len(p))
		if windows.QueryFullProcessImageName(h, 0, &p[0], &sz) == nil {
			exe = filepath.Base(windows.UTF16ToString(p[:sz]))
		}
		windows.CloseHandle(h)
	}
	return
}

func giayRanh() int {
	var li struct{ Size, Time uint32 }
	li.Size = 8
	if r, _, _ := pGetLastInputInfo.Call(uintptr(unsafe.Pointer(&li))); r == 0 {
		return 0
	}
	tick, _, _ := pGetTickCount.Call()
	return int((uint32(tick) - li.Time) / 1000)
}

func chupManHinh() (image.Image, error) {
	pSetProcessDPIAware.Call()
	x, _, _ := pGetSystemMetrics.Call(76) // SM_XVIRTUALSCREEN
	y, _, _ := pGetSystemMetrics.Call(77)
	wr, _, _ := pGetSystemMetrics.Call(78)
	hr, _, _ := pGetSystemMetrics.Call(79)
	w, h := int32(wr), int32(hr)
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("không đọc được kích thước màn hình")
	}
	hdc, _, _ := pGetDC.Call(0)
	if hdc == 0 {
		return nil, fmt.Errorf("GetDC lỗi")
	}
	defer pReleaseDC.Call(0, hdc)
	mem, _, _ := pCreateCompatibleDC.Call(hdc)
	defer pDeleteDC.Call(mem)
	bmp, _, _ := pCreateCompatibleBitmap.Call(hdc, uintptr(w), uintptr(h))
	if bmp == 0 {
		return nil, fmt.Errorf("CreateCompatibleBitmap lỗi")
	}
	defer pDeleteObject.Call(bmp)
	cu, _, _ := pSelectObject.Call(mem, bmp)
	const SRCCOPY, CAPTUREBLT = 0x00CC0020, 0x40000000
	ok, _, _ := pBitBlt.Call(mem, 0, 0, uintptr(w), uintptr(h), hdc, uintptr(int32(x)), uintptr(int32(y)), SRCCOPY|CAPTUREBLT)
	pSelectObject.Call(mem, cu)
	if ok == 0 {
		return nil, fmt.Errorf("BitBlt lỗi (màn hình đang khoá?)")
	}
	bi := bitmapInfo{Header: bitmapInfoHeader{Size: 40, Width: w, Height: -h, Planes: 1, BitCount: 32}}
	buf := make([]byte, int(w)*int(h)*4)
	n, _, _ := pGetDIBits.Call(mem, bmp, 0, uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&bi)), 0)
	if n == 0 {
		return nil, fmt.Errorf("GetDIBits lỗi")
	}
	img := image.NewRGBA(image.Rect(0, 0, int(w), int(h)))
	for i := 0; i < len(buf); i += 4 { // BGRA -> RGBA
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = buf[i+2], buf[i+1], buf[i], 255
	}
	return img, nil
}

func chayNhin(coAnh bool) error {
	exe, td := cuaSoDangDung()
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()
	b, _ := json.Marshal(NhinThay{Exe: exe, TieuDe: td, Ranh: giayRanh()})
	out.Write(b)
	out.WriteByte('\n')
	if !coAnh {
		return nil
	}
	img, err := chupManHinh()
	if err != nil {
		return err
	}
	return jpeg.Encode(out, thuNho(img, 1600), &jpeg.Options{Quality: 60})
}
