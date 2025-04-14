package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// ip2region 数据库下载地址
	ip2regionURL = "https://raw.githubusercontent.com/lionsoul2014/ip2region/master/data/ip2region.xdb"
	// 目标目录
	targetDir = "bin/data/ip2region"
)

func main() {
	// 创建目标目录
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		return
	}

	// 创建一个自定义的 Transport，跳过证书验证
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// 下载文件
	resp, err := client.Get(ip2regionURL)
	if err != nil {
		fmt.Printf("下载失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("下载失败: HTTP状态码 %d\n", resp.StatusCode)
		return
	}

	// 创建目标文件
	targetFile := filepath.Join(targetDir, "ip2region.xdb")
	out, err := os.Create(targetFile)
	if err != nil {
		fmt.Printf("创建文件失败: %v\n", err)
		return
	}
	defer out.Close()

	// 复制内容
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("保存文件失败: %v\n", err)
		return
	}

	fmt.Printf("成功下载 ip2region 数据库到: %s\n", targetFile)
}
