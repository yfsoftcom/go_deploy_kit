package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"fmt"
	"os"
	"io"
	"os/exec"
)

// 定义一个运行shell脚本的命令结构体，用于和 json 格式转换
type DeployCommand struct {
    Script string `json:"script"`
    Argument  string    `json:"argument"`
}

// 定义2个环境变量
var (
	UPLOAD_DIR = os.Getenv("GDK_UPLOAD_DIR")
	SCRIPT_DIR = os.Getenv("GDK_SCRIPT_DIR")
)

// 初始化这些变量
func init() {
	if UPLOAD_DIR == "" {
		UPLOAD_DIR = "./uploads/"
	}
	if SCRIPT_DIR == "" {
		SCRIPT_DIR = "./shells/"
	}
}

// 执行脚本的函数，返回结果和错误
func RunScriptFile(command DeployCommand) (string, error){
	cmdStr := SCRIPT_DIR + command.Script + " " + command.Argument
    cmd := exec.Command("/bin/bash", "-c", cmdStr)

    if output, err := cmd.Output();  err != nil {
        return "", err
	}else{
		return string(output), nil
	}	
}

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{ \"errno\": 0}" )
}

// 接受webhook的请求处理函数
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	// get params
	body, _ := ioutil.ReadAll(r.Body)
    r.Body.Close()
	var command DeployCommand
	// 将参数中的指令转换成 结构体
	if err := json.Unmarshal(body, &command); err != nil {
		fmt.Fprintf(w, "{ \"errno\": -1, \"msg\": \" Error:" + err.Error() +"\"}")
		return
	} 	
	// 执行shell脚本，输出结果
	if output, err := RunScriptFile(command); err != nil {
		fmt.Fprintf(w, "{ \"errno\": -1, \"msg\": \"" + err.Error() + "\"}" )
	}else{
		fmt.Fprintf(w, "{ \"errno\": 0, \"msg\": \"" + output + "\"}" )
	}
}

// 上传文件的函数
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// 
    r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintf(w, "{ \"errno\": -1, \"msg\": \" Error:" + err.Error() +"\"}")
		return
	}
	defer file.Close()
	// 打开文件流，默认使用覆盖模式，同名的文件会被覆盖
	f, err := os.OpenFile(UPLOAD_DIR + handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Fprintf(w, "{ \"errno\": -1, \"msg\": \" Error:" + err.Error() +"\"}")
		return
	}
	defer f.Close()
	io.Copy(f, file)
	fmt.Fprintf(w, "{ \"errno\":0 , \"msg\":\"upload success\"}" )
}

func main() {
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/webhook/run/", WebhookHandler)
	http.HandleFunc("/upload", UploadHandler)
	fmt.Println("Server startup at http://localhost:8000")
	if err := http.ListenAndServe("0.0.0.0:8000", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	
}