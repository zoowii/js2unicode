package main

import (
	"fmt"
	"testing"
	_ "os"
	"path/filepath"
)

func TestConvertString(t *testing.T) {
	p, _ := filepath.Abs(".")
	files, _ := ListFilesInPath(p)
	fmt.Printf("%v\n", files)
	relPath, _ := AbsPathToRelativePath("C:/abc/def/k.jpg", "C:/abc")
	fmt.Printf("relative path: %s\n", relPath)
	fmt.Printf("file sep: %s\n", string(filepath.Separator))
	remainingFiles, _ := ExcludeIgnoredFiles(files, p, "./out")
	fmt.Printf("file exclude ignored files: %v\n", remainingFiles)
	exts, _ := SplitExtensions("js,ts,css,coffee")
	fmt.Printf("exts: %v\n", exts)
	extFiles, _ := FilterFilesByExt(remainingFiles, exts)
	fmt.Printf("files with exts: %v\n", extFiles)
	for _, file := range(extFiles) {
		converted, _ := ReadFileByEncodingAndConvertToUnicode(file, "utf8")
		fmt.Printf("converted: %s\n", converted)
	}
	t1, _ := StrToUnicode("中文\"abc\"测试")
	fmt.Println(t1, len(t1))
}