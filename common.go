package main

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"github.com/bodgit/sevenzip"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

// url转换
func parseURL(urlStr string) *url.URL {
	link, err := url.Parse(urlStr)
	if err != nil {
		fyne.LogError("Could not parse URL", err)
	}

	return link
}

// 路径是否合法
func isValidPath(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	cleanAbsPath := filepath.Clean(absPath)
	cleanPath := filepath.Clean(path)
	return cleanAbsPath == cleanPath && dirExist(path)
}

// 字符串是否是数字
func isNumeric(str string) bool {
	_, err := strconv.ParseFloat(str, 64)
	return err == nil
}

// 格式化文件大小
func formatFileSize(size int64) string {
	const (
		KB = 1 << 10
		MB = 1 << 20
		GB = 1 << 30
	)
	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d Bytes", size)
	}
}

// 获取文件大小
func getFileSize(url string) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("服务器返回错误： %v", resp.Status)
	}

	return resp.ContentLength, nil
}

// 下载文件块
func downloadChunk(url string, start, end int64) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		return "", fmt.Errorf("服务器不支持分块下载：%v", resp.Status)
	}

	// 创建一个临时文件用于保存下载的文件块
	chunkPath := fmt.Sprintf("chunk_%d_%d.tmp", start, end)
	chunkFile, err := os.Create(chunkPath)
	if err != nil {
		return "", err
	}
	defer chunkFile.Close()

	_, err = io.Copy(chunkFile, resp.Body)
	if err != nil {
		return "", err
	}

	return chunkPath, nil
}

// 合并文件块
func mergeChunk(chunkPath string, output *os.File) error {
	chunkFile, err := os.Open(chunkPath)
	if err != nil {
		return err
	}
	defer chunkFile.Close()

	_, err = io.Copy(output, chunkFile)
	if err != nil {
		return err
	}

	return nil
}

// 获取文件名
func getFileName(fileURL string) string {
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		fmt.Println("Failed to parse URL:", err)
		return ""
	}
	filename := path.Base(parsedURL.Path)
	return filename
}

// 7z解压缩
func UnCompress7z(filePath, targetDir string) {
	r, err := sevenzip.OpenReader(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	for _, file := range r.File {
		rc, err := file.Open()
		if err != nil {
			log.Fatal(err)
		}
		defer rc.Close()
		fp := path.Join(targetDir, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(fp, os.ModePerm)
		} else {
			outputFile, err := os.Create(fp)
			if err != nil {
				log.Fatal(err)
			}
			defer outputFile.Close()
			_, err = io.Copy(outputFile, rc)
		}
	}
}

// 计算文件SHA1
func sumFileSHA1(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("无法打开文件:", err)
		return ""
	}
	defer file.Close()
	hash := sha1.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		fmt.Println("读取文件错误:", err)
		return ""
	}
	hashValue := hash.Sum(nil)
	hashString := hex.EncodeToString(hashValue)
	return strings.ToUpper(hashString)
}

// 检查文件是否存在
func fileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			fmt.Println("发生错误:", err)
		}
		return false
	}
	return true
}
func getString(v binding.String) string {
	gv, _ := v.Get()
	return gv
}
func getBool(v binding.Bool) bool {
	gv, _ := v.Get()
	return gv
}
func getStringList(v binding.StringList) []string {
	gv, _ := v.Get()
	return gv
}
func isProcessExist(appName string) bool {
	appary := make(map[string]int)
	cmd := exec.Command("cmd", "/C", "tasklist")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	output, _ := cmd.Output()
	n := strings.Index(string(output), "System")
	if n == -1 {
		fmt.Println("no find")
		os.Exit(1)
	}
	data := string(output)[n:]
	fields := strings.Fields(data)
	for k, v := range fields {
		if v == appName {
			appary[appName], _ = strconv.Atoi(fields[k+1])

			return true
		}
	}
	return false
}
func alertInfo(message string, win fyne.Window) {
	cnf := dialog.NewInformation("提示", message, win)
	cnf.SetDismissText("关闭")
	cnf.Show()
}
func alertConfirm(message string, callback func(bool), win fyne.Window) {
	cnf := dialog.NewConfirm("确认", message, func(b bool) {
		callback(b)
	}, win)
	cnf.SetDismissText("关闭")
	cnf.SetConfirmText("确认")
	cnf.Show()
}
func dirExist(dir string) bool {
	// 打开当前目录
	dirHandle, err := os.Open(dir)
	if err != nil {
		return false
	}
	defer dirHandle.Close()
	return true
}

// 获取系统信息
func getInfo() SysInfo {
	goarch := "x64"
	goos := "win"
	if runtime.GOARCH == "amd64" {
		goarch = "x64"
	} else if runtime.GOARCH == "386" {
		goarch = "x86"
	}
	if runtime.GOOS == "darwin" {
		goos = "mac"
	} else if runtime.GOOS == "windows" {
		goos = "win"
	}
	return SysInfo{goarch, goos}
}
func getMapKeys(m map[string]GithubRelease) []string {
	keys := make([]string, len(m))
	i := 0
	for key := range m {
		keys[i] = key
		i++
	}
	return keys
}
