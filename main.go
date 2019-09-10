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
	"errors"
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
	
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
	<head>
		<title>GDK beta</title>	
		<link rel="stylesheet" href="https://unpkg.com/spectre.css/dist/spectre.min.css">
		<link rel="stylesheet" href="https://unpkg.com/spectre.css/dist/spectre-exp.min.css">
		<link rel="stylesheet" href="https://unpkg.com/spectre.css/dist/spectre-icons.min.css">
		<link href="https://cdn.bootcss.com/font-awesome/4.7.0/css/font-awesome.min.css" rel="stylesheet">
		<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
	</head>
	<body class="bg-gray">
	<div class="container">
	<div class="hero bg-gray">
		<div class="hero-body">
			<div class="login">
				<div class="empty">
					<div class="empty-icon"><i class="icon icon-3x icon-emoji"></i></div>
					<p class="empty-title h5">Welcome To Use Go-Deploy-Kit!</p>
					<p class="empty-subtitle">Start to enjoy the kit!</p>
					<div class="column">
						<form  enctype="multipart/form-data" 
							style="margin-left: auto; margin-right: auto; max-width: 600px;"  
							class="form-horizontal" 
							action="/upload" 
							method="POST" >
							<div class="form-group">
								<div class="col-3">
									<label class="form-label">Jar:</label>
								</div>
								<div class="col-9">
									<input class="form-input" name="file" type="file">
								</div>
							</div>
							<div class="form-group">
								<div class="col-3">
									<label class="form-label">Shell:</label>
								</div>
								<div class="col-9">
									<input class="form-input" name="shell" type="text">
								</div>
							</div>
							<div class="form-group">
								<div class="col-3">
									<label class="form-label">Argument:</label>
								</div>
								<div class="col-9">
									<textarea class="form-input" name="argument"></textarea>
								</div>
							</div>
							<button class="btn btn-success" type="submit"><i class="icon icon-upload"></i> Submit</button>
						</form>
					</div>
				</div>
			</div>
		</div>
	</div>
	
	</div>
	</body>
</html>
		`)
	
}

func Fail(w http.ResponseWriter, err error){
	fmt.Fprintf(w, fmt.Sprintf(`{ "errno": -1, "msg": " Error: %s"}`,  err.Error() ))
}

func Success(w http.ResponseWriter, data string){
	fmt.Fprintf(w, fmt.Sprintf(`{ "errno": 0, "msg": " %s"}`,  data))
}

// 接受webhook的请求处理函数
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	// get params
	body, _ := ioutil.ReadAll(r.Body)
	fmt.Println(fmt.Sprintf(`Webhook Body: %s`, body))
    r.Body.Close()
	var command DeployCommand
	// 将参数中的指令转换成 结构体
	if err := json.Unmarshal(body, &command); err != nil {
		Fail(w, err)
		return
	} 	
	// 执行shell脚本，输出结果
	if output, err := RunScriptFile(command); err != nil {
		Fail(w, err)
		return
	}else{
		Success(w, output)
	}
}

// 接受webhook的请求处理函数
func DeployHandler(w http.ResponseWriter, r *http.Request) {
	// get params
	var command DeployCommand
	// 将参数中的指令转换成 结构体
	if err := json.Unmarshal([]byte(`{"script":"deploy.sh", "argument":""}`), &command); err != nil {
		Fail(w, err)
		return
	} 	
	// 执行shell脚本，输出结果
	if output, err := RunScriptFile(command); err != nil {
		Fail(w, err)
		return
	}else{
		Success(w, output)
	}
}

// 上传文件的函数
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	// 
	if method := r.Method; method == "GET" {
		Fail(w, errors.New("Only support Post"))
		return
	}
    r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		Fail(w, err)
		return
	}
	defer file.Close()
	// 打开文件流，默认使用覆盖模式，同名的文件会被覆盖
	f, err := os.OpenFile(UPLOAD_DIR + handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		Fail(w, err)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	if shell := r.PostFormValue("shell"); shell != ""{
		argument := r.PostFormValue("argument")
		// 执行脚本
		command := & DeployCommand {
			Script: shell,
			Argument: argument,
		}
		if output, err := RunScriptFile( *command); err != nil {
			Fail(w, err)
			return
		}else {
			Success(w, output)
			return;
		}
	}

	Success(w, "upload success")
}

func main() {
	http.HandleFunc("/", IndexHandler)
	http.HandleFunc("/webhook/run/", WebhookHandler)
	http.HandleFunc("/webhook/deploy", DeployHandler)
	http.HandleFunc("/upload", UploadHandler)
	fmt.Println("Server startup at http://localhost:8000")
	if err := http.ListenAndServe("0.0.0.0:8000", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	
}