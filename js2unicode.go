package main

import (
	"path"
	"io/ioutil"
	"strings"
	"flag"
	"fmt"
	"os"
	"strconv"
	"errors"
	"path/filepath"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
	_ "strings"
)

var (
	g_path = flag.String("path", "", "path to source file/directory")
	g_encoding = flag.String("encoding", "utf8", "encoding of source files, only support utf8 and gbk now")
	g_ext = flag.String("ext", "js", "extensions of file need be converted, default is js, you can also use multi ext by specific js,ts,css,coffee")
	g_out = flag.String("out", "", "output directory of result files(default is source dir/source file's parent dir(put in same name directories))")
)

func show_usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: %s -path=<path> [-out=<output_directory>] [-encoding=<encoding>] [-ext=<ext>]\n",
		os.Args[0])
	fmt.Fprintf(os.Stderr,
		"Flags:\n")
	flag.PrintDefaults()
}

func StrToUnicode(source string) (string, error) {
	result := ""
	for _, c := range(source) {
		if (c > 126 || c < 32) && (c!=10 && c!=9 && c!=13)  {
			quoted := strconv.QuoteToASCII(string(c))
			textUnquoted := quoted[1 : len(quoted)-1]
			result += textUnquoted
		} else {
			result += string(c)
		}
	}
	// textQuoted := strconv.QuoteToASCII(source)
	// textUnquoted := textQuoted[1 : len(textQuoted)-1]
	return result, nil
}

func IsDirectory(path string) bool {
    fileInfo, err := os.Stat(path)
    if err != nil{
      return false
    }
    return fileInfo.IsDir()
}

func IsFile(path string) bool {
	fileInfo, err := os.Stat(path)
    if err != nil{
      return false
    }
    return !fileInfo.IsDir()
}

func ListFilesInPath(path string) ([]string, error) {
	if IsFile(path) {
		return []string{path}, nil
	}
	var result []string
	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
			if ( f == nil ) {return err}
			if f.IsDir() {return nil}
			result = append(result, path)
			return nil
	})
	if err != nil {
		return []string{}, nil
	} else {
		return result, nil
	}
}

func AbsPathToRelativePath(absPath string, dirAbsPath string) (string, error) {
	// 绝对路径改成相对路径
	if !strings.HasPrefix(absPath, dirAbsPath) || len(absPath)<=len(dirAbsPath) {
		return "", errors.New("can't get relative path of file " + absPath + " in dir " + dirAbsPath)
	}
	subPath := absPath[len(dirAbsPath)+1:]
	return subPath, nil
}


func ExcludeIgnoredFiles(files []string, dirAbsPath string, outputDir string) ([]string, error) {
	// 跳过需要忽略的文件，这里暂时用相对路径中有.或/.的子目录，也需要跳过outputDir。参数files是文件绝对路径列表
	var result []string
	for _, file := range(files) {
		relativePath, err := AbsPathToRelativePath(file, dirAbsPath)
		if err != nil {
			return result, err
		}
		if strings.HasPrefix(relativePath, ".") || strings.HasPrefix(relativePath, string(filepath.Separator) + ".") {
			continue
		}
		if strings.HasPrefix(file, outputDir) {
			continue
		}
		result = append(result, file)
	}
	return result, nil
}

func matchOneExt(path string, exts []string) bool {
	// 判断文件路径是否至少满足exts中某一项
	for _, ext := range(exts) {
		// 不用filepath.Ext，是考虑ext可能是.min.js这类格式
		if strings.HasSuffix(path, "." + ext) && len(path)>(len(ext) + 1) {
			return true
		}
	}
	return false
}

func FilterFilesByExt(files []string, exts []string) ([]string, error) {
	// 过滤出满足ext后缀要求的文件格式
	var result []string
	for _, file := range(files) {
		if matchOneExt(file, exts) {
			result = append(result, file)
		} 
	}
	return result, nil
}

func SplitExtensions(fullExt string) ([]string, error) {
	// 把a,b,c格式的多个后缀字符串转成后缀数组
	exts := strings.Split(fullExt, ",")
	if len(exts) < 1 {
		return []string{}, errors.New("invalid ext string " + fullExt)
	}
	for _, ext := range(exts) {
		if len(ext) < 1 {
			return []string{}, errors.New("invalid ext string "+ fullExt)
		}
	}
	return exts, nil
}

func ReadFileAsEncoding(absPath string, enc string) (string, error) {
	f, err := os.Open(absPath)
    if err != nil {
        return "", err
    }
    defer f.Close()
	encLower := strings.ToLower(enc)
	var decoder *encoding.Decoder
	if encLower == "gbk" {
		decoder = simplifiedchinese.GBK.NewDecoder()
	} else if encLower == "utf8" || encLower == "utf-8" {
		decoder = unicode.UTF8.NewDecoder()
	}
    if decoder == nil {
        return "", errors.New("invalid encoding")
	}
	reader := transform.NewReader(f, decoder)
	d, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}
    return string(d), nil
}

func ReadFileByEncodingAndConvertToUnicode(absPath string, encoding string) (string, error) {
	// 按encoding读取文件内容并转换内容到unicode格式文本
	content, err := ReadFileAsEncoding(absPath, encoding)
	if err != nil {
		return "", err
	}
	return StrToUnicode(content)
}

func WriteResultFilesToOutputDir(contents []string, sourceFiles []string, sourceAbsDirPath string, outputDir string) error {
	// 按和输入目录的同样的子目录结构输出结果文件
	if len(contents) != len(sourceFiles) {
		return errors.New("source files and contents count not match")
	}
	var _, outputDirStatErr = os.Stat(outputDir)
	if os.IsNotExist(outputDirStatErr) {
		os.MkdirAll(outputDir, os.ModePerm)
	}
	for i := range(contents) {
		content := contents[i]
		sourceFile := sourceFiles[i]
		relativePath, err := AbsPathToRelativePath(sourceFile, sourceAbsDirPath)
		if err != nil {
			return err
		}
		outputFile := path.Join(outputDir, relativePath)
		ioutil.WriteFile(outputFile, []byte(content), 0644)
	}
	return nil
}

func main() {
	flag.Usage = show_usage
	flag.Parse()
	if *g_path=="" {
		show_usage()
		return
	}
	var sourceDir string
	var err error
	sourceDir, err = filepath.Abs(*g_path)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	if IsFile(sourceDir) {
		sourceDir = filepath.Dir(sourceDir)
	}
	var outputDir string
	if len(*g_out)<1 {
		outputDir = sourceDir
	} else {
		outputDir, err = filepath.Abs(*g_out)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			return
		}
		if IsFile(outputDir) {
			fmt.Printf("output directory must be directory, but got %s\n", outputDir)
			return
		}
	}
	fmt.Printf("source path: %s, encoding: %s, ext: %s, source dir: %s, output dir: %s\n", *g_path, *g_encoding, *g_ext, sourceDir, outputDir)
	var files []string
	var exts []string
	files, err = ListFilesInPath(sourceDir)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	files, err = ExcludeIgnoredFiles(files, sourceDir, outputDir)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	exts, err = SplitExtensions(*g_ext)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	files, err = FilterFilesByExt(files, exts)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	var contents []string
	for _, file := range(files) {
		converted, _ := ReadFileByEncodingAndConvertToUnicode(file, "utf8")
		contents = append(contents, converted)
	}
	err = WriteResultFilesToOutputDir(contents, files, sourceDir, outputDir)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	fmt.Printf("converted done. there are %d files with these extensions. \n", len(files)) 
}