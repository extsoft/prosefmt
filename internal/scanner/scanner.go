package scanner

import (
	"io"
	"os"
	"path/filepath"
)

const maxScanBytes = 32 * 1024

func Scan(paths []string) ([]string, map[string]string, error) {
	var out []string
	skipped := make(map[string]string)
	seen := make(map[string]bool)
	for _, root := range paths {
		info, err := os.Stat(root)
		if err != nil {
			return nil, nil, err
		}
		if info.Mode().IsRegular() {
			abs, _ := filepath.Abs(root)
			if seen[abs] {
				continue
			}
			seen[abs] = true
			ok, reason := isTextFileWithReason(root)
			if ok {
				out = append(out, root)
			} else if reason != "" {
				skipped[root] = reason
			}
			continue
		}
		if info.IsDir() {
			err := filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !fi.Mode().IsRegular() {
					return nil
				}
				abs, _ := filepath.Abs(p)
				if seen[abs] {
					return nil
				}
				seen[abs] = true
				ok, reason := isTextFileWithReason(p)
				if ok {
					out = append(out, p)
				} else if reason != "" {
					skipped[p] = reason
				}
				return nil
			})
			if err != nil {
				return nil, nil, err
			}
		}
	}
	return out, skipped, nil
}

func isTextFile(path string) bool {
	ok, _ := isTextFileWithReason(path)
	return ok
}

func isTextFileWithReason(path string) (bool, string) {
	f, err := os.Open(path)
	if err != nil {
		return false, ""
	}
	defer f.Close()
	buf := make([]byte, maxScanBytes)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, ""
	}
	buf = buf[:n]
	if n == 0 {
		return true, ""
	}
	for _, b := range buf {
		if b == 0 {
			return false, "null byte"
		}
		if b <= 0x08 || b == 0x0B || b == 0x0C || (b >= 0x0E && b <= 0x1F) || b == 0x7F {
			return false, "binary or control characters"
		}
	}
	return true, ""
}
