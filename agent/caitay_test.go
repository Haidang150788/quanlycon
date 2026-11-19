package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestXoaTokenTrongFile(t *testing.T) {
	cu := tokenNhung
	defer func() { tokenNhung = cu }()
	tokenNhung = "QLCTOKEN=abc123def456"
	if layTokenNhung() != "abc123def456" {
		t.Fatal("khong doc duoc token nhung")
	}
	p := filepath.Join(t.TempDir(), "x.exe")
	goc := []byte("dau..QLCTOKEN=abc123def456..duoi")
	_ = os.WriteFile(p, goc, 0755)
	if err := xoaTokenTrongFile(p); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(p)
	if bytes.Contains(b, []byte("abc123")) || len(b) != len(goc) {
		t.Fatalf("van con token hoac sai do dai: %q", b)
	}
	tokenNhung = ""
	if layTokenNhung() != "" {
		t.Fatal("rong phai ra rong")
	}
}
