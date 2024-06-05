package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/atotto/clipboard"
)

var appID string
var secret string

type WxImage struct {
	Url string `json:"url"`
}

func UpImage(token string, bodyBuf io.Reader) string {

	//upload
	req, err := http.NewRequest("POST", fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/media/uploadimg?access_token=%s", token), bodyBuf)
	req.Header.Add("Content-Type", "multipart/form-data")
	urlQuery := req.URL.Query()
	if err != nil {
		log.Fatalln("new request err:", err)
	}
	urlQuery.Add("access_token", token)

	req.URL.RawQuery = urlQuery.Encode()
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	defer res.Body.Close()
	jsonbody, err := io.ReadAll(res.Body)

	if err != nil {
		log.Fatalf("上传图片失败: %v", err)
	}

	var result WxImage
	err = json.Unmarshal(jsonbody, &result)
	if err != nil {
		log.Fatalf("解析上传图片返回结果失败: %v", err)
	}

	if result.Url == "" {
		fmt.Println(string(jsonbody))
		log.Fatalln("上传图片失败")
	}

	return result.Url
}

func GetBodyByClipboard() io.Reader {
	cmd := exec.Command("pngpaste", "-")
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatalln(err)
	}

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("image", "clipboard.png")
	if err != nil {
		log.Fatalf("error writing to buffer: %v", err)
	}

	_, err = io.Copy(fileWriter, bytes.NewBuffer(stdout))
	if err != nil {
		log.Fatalln("copy file err:", err)
	}

	bodyWriter.Close()
	return bodyBuf
}
func GetBodyByClipboard2() io.Reader {
	cmd := exec.Command("pngpaste", "/tmp/clipboard.png")
	err := cmd.Run()
	if err != nil {
		// 获取 pngpaste 命令的错误输出
		log.Fatalln(err)
	}

	return GetBodyByFile("/tmp/clipboard.png")
}

func GetBodyByFile(filename string) io.Reader {
	//打开文件
	fh, err := os.Open(filename)
	if err != nil {
		log.Fatalf("open file error: %v", err)
	}
	defer fh.Close()
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("image", filepath.Base(filename))
	if err != nil {
		log.Fatalf("error writing to buffer: %v", err)
	}

	_, err = io.Copy(fileWriter, fh)
	if err != nil {
		log.Fatalln("copy file err:", err)
	}

	bodyWriter.Close()
	return bodyBuf
}

func GetToken(appID, secret string) (token string) {

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, secret)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("request error: %v", err)
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("request error: %v", err)
		return
	}

	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("request error: %v", err)
		return
	}
	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Fatalf("unmarshal error: %v", err)
		return
	}
	
	if _, ok := result["access_token"];ok {
		
		return result["access_token"].(string)

	}

	log.Fatalf("get access_token err:$v",string(body))
	return

}

func main() {

	appID = os.Getenv("WX_APP_ID")
	secret = os.Getenv("WX_SECRET")

	if len(appID) == 0 || len(secret) == 0 {
		log.Fatalln("请设置环境变量 WX_APP_ID 和 WX_SECRET")
		return
	}

	token := GetToken(appID, secret)
	var bodyBuffer io.Reader

	if len(os.Args) >= 2 {
		filename := os.Args[1]
		bodyBuffer = GetBodyByFile(filename)
	} else {
		bodyBuffer = GetBodyByClipboard()

	}

	url := UpImage(token, bodyBuffer)
	if len(url) > 0 {
		fmt.Println("[Pic2wx SUCCESS]:")
		clipboard.WriteAll(url)
		fmt.Println("url已复制到剪切板")
	}
	fmt.Println(url)
}
