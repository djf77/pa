package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

//全局定义
var (
	reImg         = `https?://[^"]+?(\.(jpg)|(png))`
	chanImageUrls chan string
	waitGroup     sync.WaitGroup
	chanTask      chan string
)

//异常处理
func HandleError(err error, why string) {
	if err != nil {
		fmt.Println(why, err)
	}
}

//获取页面
func GetPageStr(url string) (pageStr string) {
	resp, err := http.Get(url)
	HandleError(err, "http.Get url")
	defer resp.Body.Close()
	pageBytes, err := ioutil.ReadAll(resp.Body)
	HandleError(err, "ioutil.ReadAll")
	pageStr = string(pageBytes)
	return pageStr
}

//筛选出图片链接（当前页）
func GetImgs(url string) (urls []string) {
	pageStr := GetPageStr(url)
	re := regexp.MustCompile(reImg)
	//输出全部
	results := re.FindAllStringSubmatch(pageStr, -1)
	fmt.Printf("共找到%d条结果\n", len(results))
	//遍历
	for _, result := range results {
		url := result[0]
		urls = append(urls, url)
	}
	return
}

//爬图片链接到管道
//传的整页链接
func getImgUrls(url string) {
	urls := GetImgs(url)
	//遍历,进管道
	for _, url := range urls {
		chanImageUrls <- url
	}
	//标识当前协程完成
	//监控协程
	chanTask <- url
	waitGroup.Done()
}

//任务统计协程
func Check() {
	var count int
	for {
		url := <-chanTask
		fmt.Printf("%s 完成了爬取任务\n", url)
		count++
		if count == 8 {
			close(chanImageUrls)
			break
		}
	}
	waitGroup.Done()
}

//下载1
func DownloadFile(url string, filename string) (ok bool) {
	resp, err := http.Get(url)
	HandleError(err, "http.get.url")
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	HandleError(err, "resp.body")
	filename = "D:/Gowork/pa/Img/" + filename
	//写出
	err = ioutil.WriteFile(filename, bytes, 0)
	if err != nil {
		return false
	} else {
		return true
	}
}

//截取名字
func GetFilenameFromUrl(url string) (filename string) {
	lastIndex := strings.LastIndex(url, "/")
	filename = url[lastIndex+1:]
	return
}

//下载2
func DownloadImg() {
	for url := range chanImageUrls {
		filename := GetFilenameFromUrl(url)
		ok := DownloadFile(url, filename)
		if ok {
			fmt.Printf("%s 下载成功\n", filename)
		} else {
			fmt.Printf("%s 下载失败\n", filename)
		}
	}
	waitGroup.Done()
}

func main() {
	//初始化
	chanImageUrls = make(chan string, 100000)
	chanTask = make(chan string, 8)
	//爬虫协程
	for i := 1; i < 9; i++ {
		waitGroup.Add(1)
		go getImgUrls("https://tieba.baidu.com/p/6423847316?pn=" + strconv.Itoa(i))
	}
	//任务统计
	waitGroup.Add(1)
	go Check()
	//下载
	for i := 0; i < 4; i++ {
		waitGroup.Add(1)
		go DownloadImg()
	}
	waitGroup.Wait()
}
