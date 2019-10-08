package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"time"

	mux "github.com/gorilla/mux"
)

// 定义一个运行shell脚本的命令结构体，用于和 json 格式转换
type DeployCommand struct {
	Script   string `json:"script"`
	Argument string `json:"argument"`
}

// 定义2个环境变量
var (
	UPLOAD_DIR  = os.Getenv("GDK_UPLOAD_DIR")
	SCRIPT_DIR  = os.Getenv("GDK_SCRIPT_DIR")
	ErrorLogger *log.Logger
	InfoLogger  *log.Logger
)

// 初始化这些变量
func init() {
	if UPLOAD_DIR == "" {
		UPLOAD_DIR = "./uploads/"
	}
	if SCRIPT_DIR == "" {
		SCRIPT_DIR = "./shells/"
	}

	errorFile, err := os.OpenFile("errors.txt",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	infoFile, err := os.OpenFile("infos.txt",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open error log file:", err)
	}

	ErrorLogger = log.New(io.MultiWriter(errorFile, os.Stderr),
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	InfoLogger = log.New(io.MultiWriter(infoFile, os.Stdout),
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

// 执行脚本的函数，返回结果和错误
func RunScriptFile(command DeployCommand) (string, error) {
	cmdStr := SCRIPT_DIR + command.Script + " " + command.Argument
	cmd := exec.Command("/bin/bash", "-c", cmdStr)

	if output, err := cmd.Output(); err != nil {
		return "", err
	} else {
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

func Fail(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, fmt.Sprintf(`{ "errno": -1, "msg": " Error: %s"}`, err.Error()))
}

func Success(w http.ResponseWriter, data string) {
	fmt.Fprintf(w, fmt.Sprintf(`{ "errno": 0, "msg": " %s"}`, data))
}

// 接受webhook的请求处理函数
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	InfoLogger.Println(fmt.Sprintf(`Filename is: %s`, filename))

	// get params
	body, _ := ioutil.ReadAll(r.Body)
	InfoLogger.Println(fmt.Sprintf(`Webhook Body: %s`, body))
	// 输出到日志文件和终端中
	r.Body.Close()
	var command DeployCommand
	// 将参数中的指令转换成 结构体
	if err := json.Unmarshal(body, &command); err != nil {
		ErrorLogger.Println(err)
		Fail(w, err)
		return
	}
	// 执行shell脚本，输出结果
	if output, err := RunScriptFile(command); err != nil {
		ErrorLogger.Println(err)
		Fail(w, err)
		return
	} else {
		Success(w, output)
	}
}

// 接受webhook的请求处理函数
func DeployHandler(w http.ResponseWriter, r *http.Request) {
	InfoLogger.Println(`Deploy Webhook Execute`)
	// get params
	var command DeployCommand
	// 将参数中的指令转换成 结构体
	if err := json.Unmarshal([]byte(`{"script":"deploy.sh", "argument":""}`), &command); err != nil {
		ErrorLogger.Println(err)
		Fail(w, err)
		return
	}
	// 执行shell脚本，输出结果
	if output, err := RunScriptFile(command); err != nil {
		ErrorLogger.Println(err)
		Fail(w, err)
		return
	} else {
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
	f, err := os.OpenFile(UPLOAD_DIR+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		Fail(w, err)
		return
	}
	defer f.Close()
	io.Copy(f, file)

	if shell := r.PostFormValue("shell"); shell != "" {
		argument := r.PostFormValue("argument")
		// 执行脚本
		command := &DeployCommand{
			Script:   shell,
			Argument: argument,
		}
		if output, err := RunScriptFile(*command); err != nil {
			ErrorLogger.Println(err)
			Fail(w, err)
			return
		} else {
			Success(w, output)
			return
		}
	}

	Success(w, "upload success")
}

func main() {

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandler)
	r.HandleFunc("/webhook/run/", WebhookHandler)
	r.HandleFunc("/webhook/deploy", DeployHandler)
	r.HandleFunc("/webhook/shell/{filename}", WebhookHandler)
	r.HandleFunc("/upload", UploadHandler)
	fmt.Println("Server startup at http://localhost:8000")

	srv := &http.Server{
		Addr:         "0.0.0.0:8000",
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)

}
