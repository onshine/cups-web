package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// minimalPDF 是一段最小可被 pdf.Open 解析的 PDF 字节流，用于 smoke test。
// 结构：%PDF-1.4 + 1 个 Pages 对象 + 1 个 Page 对象 + xref + trailer + %%EOF。
var minimalPDF = []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Count 1 /Kids [3 0 R] >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>
endobj
xref
0 4
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000111 00000 n 
trailer
<< /Size 4 /Root 1 0 R >>
startxref
174
%%EOF
`)

func writeTempFile(t *testing.T, content []byte, name string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return path
}

func TestFileHasPDFHeader(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    bool
	}{
		{"valid PDF", minimalPDF, true},
		{"fake text", []byte("hello world this is not a pdf"), false},
		{"too short", []byte("%PD"), false},
		{"empty", []byte{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempFile(t, tt.content, "t.bin")
			got := fileHasPDFHeader(path)
			if got != tt.want {
				t.Errorf("fileHasPDFHeader(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestFileHasPDFHeader_MissingFile(t *testing.T) {
	if fileHasPDFHeader("/nonexistent/path/xxx.pdf") {
		t.Error("expected false for missing file")
	}
}

func TestDiagnosePDF_DoesNotPanic(t *testing.T) {
	// 1. 合法 PDF
	validPath := writeTempFile(t, minimalPDF, "valid.pdf")
	diagnosePDF(validPath)

	// 2. 截断的损坏 PDF
	corruptedPath := writeTempFile(t, []byte("%PDF-1.4\nthis is corrupted"), "bad.pdf")
	diagnosePDF(corruptedPath)

	// 3. 根本不存在的文件
	diagnosePDF("/this/does/not/exist.pdf")

	// 4. 空文件
	emptyPath := writeTempFile(t, nil, "empty.pdf")
	diagnosePDF(emptyPath)
}

// TestNormalizePDF_Passthrough 验证：当 gs / libreoffice 都不可用（或对最小 PDF
// 处理失败）时，normalizePDF 不会报错，Method 为 "passthrough"，输出路径等于输入路径。
// 本用例在有 gs 的开发机上会走 ghostscript 路径；结果兼容两种情况。
func TestNormalizePDF_MinimalPDF(t *testing.T) {
	path := writeTempFile(t, minimalPDF, "mini.pdf")
	ctx := context.Background()
	res, err := normalizePDF(ctx, path)
	if err != nil {
		t.Fatalf("normalizePDF err: %v", err)
	}
	if res == nil {
		t.Fatal("res is nil")
	}
	switch res.Method {
	case "ghostscript", "libreoffice":
		// 真实标准化路径：输出文件必须存在
		if st, err := os.Stat(res.OutputPath); err != nil || st.Size() == 0 {
			t.Errorf("normalized output missing or empty: %v", err)
		}
		if res.Cleanup != nil {
			res.Cleanup()
		}
	case "passthrough":
		if res.OutputPath != path {
			t.Errorf("passthrough OutputPath = %q, want %q", res.OutputPath, path)
		}
	default:
		t.Errorf("unexpected Method=%q", res.Method)
	}
}

func TestNormalizePDF_MissingFile(t *testing.T) {
	_, err := normalizePDF(context.Background(), "/definitely/missing/xxx.pdf")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
