package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// normalizePDFResult 描述一次 PDF 标准化的产物与元信息。
//
// OutputPath  标准化后的 PDF 绝对路径；passthrough 模式下即输入原路径
// Cleanup     临时目录清理函数（若无临时产物则为 nil）
// Method      实际走了哪条路径："ghostscript" / "libreoffice" / "passthrough"
// Warnings    降级原因、失败子步骤等诊断信息（仅用于日志，不参与控制流）
type normalizePDFResult struct {
	OutputPath string
	Cleanup    func()
	Method     string
	Warnings   []string
}

// normalizePDF 把任意 PDF 转成更兼容的版本：
//   1. 优先调用 Ghostscript `pdfwrite`，统一降级到 PDF 1.4 并强制嵌入所有字体
//      —— 解决 `UniGB-UCS2-H` 等外部 CMap + 字体未嵌入导致的打印乱码
//   2. Ghostscript 不可用或失败时，fallback 到 LibreOffice headless 重新导出
//   3. 两者都不可用时回退为 passthrough（原样返回），保证兜底打印链路不中断
//
// 失败不会返回 error（除非输入文件本身不可读），调用方只需关注 result.Method
// 来决定是否使用标准化后的产物。
func normalizePDF(ctx context.Context, inputPath string) (*normalizePDFResult, error) {
	if _, err := os.Stat(inputPath); err != nil {
		return nil, fmt.Errorf("normalizePDF: stat input: %w", err)
	}

	// 给标准化一个相对宽松的超时，避免个别 PDF 卡死整条打印链路。
	nctx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	res := &normalizePDFResult{Method: "passthrough", OutputPath: inputPath}

	// 1. Ghostscript 优先
	if path, cleanup, err := runGhostscriptNormalize(nctx, inputPath); err == nil {
		res.Method = "ghostscript"
		res.OutputPath = path
		res.Cleanup = cleanup
		log.Printf("[pdf-normalize] method=ghostscript in=%s out=%s", filepath.Base(inputPath), filepath.Base(path))
		return res, nil
	} else {
		res.Warnings = append(res.Warnings, "ghostscript: "+err.Error())
		log.Printf("[pdf-normalize] ghostscript unavailable/failed: %v", err)
	}

	// 2. LibreOffice fallback
	if path, cleanup, err := convertPDFViaLibreOffice(nctx, inputPath); err == nil {
		res.Method = "libreoffice"
		res.OutputPath = path
		res.Cleanup = cleanup
		log.Printf("[pdf-normalize] method=libreoffice in=%s out=%s", filepath.Base(inputPath), filepath.Base(path))
		return res, nil
	} else {
		res.Warnings = append(res.Warnings, "libreoffice: "+err.Error())
		log.Printf("[pdf-normalize] libreoffice fallback failed: %v", err)
	}

	// 3. passthrough
	log.Printf("[pdf-normalize] method=passthrough in=%s warnings=%v", filepath.Base(inputPath), res.Warnings)
	return res, nil
}

// runGhostscriptNormalize 调用 `gs` 将 PDF 重写为兼容性更好的 1.4 版本并嵌入所有字体。
// 若 gs 二进制不在 PATH 中，返回明确的错误以便上层降级。
func runGhostscriptNormalize(ctx context.Context, inputPath string) (string, func(), error) {
	gsBin, err := exec.LookPath("gs")
	if err != nil {
		return "", nil, fmt.Errorf("gs not found in PATH: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "pdf-normalize-gs-")
	if err != nil {
		return "", nil, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }
	outPath := filepath.Join(tmpDir, "normalized.pdf")

	start := time.Now()
	args := []string{
		"-dNOPAUSE", "-dBATCH", "-dQUIET", "-dSAFER",
		"-sDEVICE=pdfwrite",
		"-dCompatibilityLevel=1.4",
		"-dPDFSETTINGS=/prepress",
		"-dEmbedAllFonts=true",
		"-dSubsetFonts=true",
		"-dDetectDuplicateImages=true",
		"-dCompressFonts=true",
		"-dAutoRotatePages=/None",
		"-sOutputFile=" + outPath,
		inputPath,
	}
	cmd := exec.CommandContext(ctx, gsBin, args...)
	cmd.Env = append(os.Environ(), "LANG=C.UTF-8", "LC_ALL=C.UTF-8")
	out, err := cmd.CombinedOutput()
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("gs pdfwrite failed: %w - %s", err, strings.TrimSpace(string(out)))
	}
	if st, err := os.Stat(outPath); err != nil || st.Size() == 0 {
		cleanup()
		return "", nil, fmt.Errorf("gs pdfwrite produced empty output: %v", err)
	}
	log.Printf("[pdf-normalize] ghostscript elapsed=%s out=%s", time.Since(start).Round(time.Millisecond), filepath.Base(outPath))
	return outPath, cleanup, nil
}

// fileHasPDFHeader 检查文件前若干字节是否以 "%PDF-" 开头。
// 用于在 passthrough 路径上做最基本的健全性校验。
func fileHasPDFHeader(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	buf := make([]byte, 8)
	n, _ := f.Read(buf)
	if n < 5 {
		return false
	}
	return strings.HasPrefix(string(buf[:n]), "%PDF-")
}

// diagnosePDF 对 PDF 做一次轻量诊断（仅日志），发现可能导致打印乱码的特征时输出警告。
// 不 panic、不返回 error；任何异常都被吞掉，保证主流程可用。
func diagnosePDF(path string) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[pdf-diag] panic recovered: %v", r)
		}
	}()

	f, err := os.Open(path)
	if err != nil {
		log.Printf("[pdf-diag] open failed: %v", err)
		return
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		log.Printf("[pdf-diag] stat failed: %v", err)
		return
	}

	// 读 header（前 1KB）与 tail（后 8KB）即可覆盖版本号 / linearization / 外部 CMap 标志
	headBuf := make([]byte, 1024)
	hn, _ := f.ReadAt(headBuf, 0)
	head := string(headBuf[:hn])

	tailSize := int64(8192)
	if st.Size() < tailSize {
		tailSize = st.Size()
	}
	tailBuf := make([]byte, tailSize)
	_, _ = f.ReadAt(tailBuf, st.Size()-tailSize)
	tail := string(tailBuf)

	// 版本号
	version := "unknown"
	if m := regexp.MustCompile(`%PDF-(\d+\.\d+)`).FindStringSubmatch(head); len(m) == 2 {
		version = m[1]
	}

	linearized := strings.Contains(head, "/Linearized")

	// 粗略嗅探全文里的可疑 CMap / 未嵌入字体。scanner 只读前 512KB，避免在超大 PDF 上过度开销。
	suspiciousEncodings := map[string]bool{}
	if _, err := f.Seek(0, 0); err == nil {
		lr := &boundedReader{r: f, remaining: 512 * 1024}
		scanner := bufio.NewScanner(lr)
		scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)
		patterns := []string{"UniGB-UCS2-H", "UniCNS-UCS2-H", "UniJIS-UCS2-H", "UniKS-UCS2-H", "GBK-EUC-H", "GB-EUC-H"}
		for scanner.Scan() {
			line := scanner.Text()
			for _, p := range patterns {
				if strings.Contains(line, p) {
					suspiciousEncodings[p] = true
				}
			}
		}
	}

	enc := make([]string, 0, len(suspiciousEncodings))
	for k := range suspiciousEncodings {
		enc = append(enc, k)
	}

	log.Printf("[pdf-diag] path=%s size=%d version=%s linearized=%v suspiciousCMaps=%v tailOK=%v",
		filepath.Base(path), st.Size(), version, linearized, enc, strings.Contains(tail, "%%EOF"))
}

// methodOf 安全地取 normalizePDFResult.Method，result 为 nil 时返回 "none"。
// 主要用于日志格式化，避免 nil 解引用 panic。
func methodOf(r *normalizePDFResult) string {
	if r == nil {
		return "none"
	}
	return r.Method
}

// boundedReader 限制从底层 Reader 读取的最大字节数，避免 diagnosePDF 在超大文件上耗时过长。
type boundedReader struct {
	r         interface{ Read(p []byte) (int, error) }
	remaining int64
}

func (b *boundedReader) Read(p []byte) (int, error) {
	if b.remaining <= 0 {
		return 0, fmt.Errorf("EOF-bounded")
	}
	if int64(len(p)) > b.remaining {
		p = p[:b.remaining]
	}
	n, err := b.r.Read(p)
	b.remaining -= int64(n)
	return n, err
}
