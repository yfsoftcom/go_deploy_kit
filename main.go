/*
 The comment for the package .
**/
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
	"path/filepath"
	"strconv"
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
	uploadDir   = os.Getenv("GDK_UPLOAD_DIR")
	scriptDir   = os.Getenv("GDK_SCRIPT_DIR")
	port        = os.Getenv("PORT")
	timeout     = os.Getenv("TIMEOUT")
	ErrorLogger *log.Logger
	InfoLogger  *log.Logger
	htmlDir     = "ui/build"
)

// --------- Support SPA -------
type spaHandler struct {
	staticPath string
	indexPath  string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	path = filepath.Join(h.staticPath, path)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

// 初始化这些变量
func init() {
	if uploadDir == "" {
		uploadDir = "./uploads/"
	}
	if scriptDir == "" {
		scriptDir = "./shells/"
	}
	if port == "" {
		port = "8000"
	}
	if timeout == "" {
		timeout = "60"
	}
	errorFile, err := os.OpenFile("errors.txt",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("failed to open error log file:", err)
	}

	infoFile, err := os.OpenFile("infos.txt",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("failed to open error log file:", err)
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
	cmdStr := scriptDir + command.Script + " " + command.Argument
	cmd := exec.Command("/bin/bash", "-c", cmdStr)

	if output, err := cmd.Output(); err != nil {
		return "", err
	} else {
		return string(output), nil
	}
}

func fail(w http.ResponseWriter, err error) {
	fmt.Fprintf(w, fmt.Sprintf(`{ "errno": -1, "msg": " Error: %s"}`, err.Error()))
}

func success(w http.ResponseWriter, data string) {
	fmt.Fprintf(w, fmt.Sprintf(`{ "errno": 0, "msg": " %s"}`, data))
}

// 接受webhook的请求处理函数
func ApiHandler(w http.ResponseWriter, r *http.Request) {
	success(w, "ok")
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
		fail(w, err)
		return
	}
	// 执行shell脚本，输出结果
	if output, err := RunScriptFile(command); err != nil {
		ErrorLogger.Println(err)
		fail(w, err)
		return
	} else {
		success(w, output)
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
		fail(w, err)
		return
	}
	// 执行shell脚本，输出结果
	if output, err := RunScriptFile(command); err != nil {
		ErrorLogger.Println(err)
		fail(w, err)
		return
	} else {
		success(w, output)
	}
}

// 上传文件的函数
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	//
	if method := r.Method; method == "GET" {
		fail(w, errors.New("Only support Post"))
		return
	}
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		fail(w, err)
		return
	}
	defer file.Close()
	// 打开文件流，默认使用覆盖模式，同名的文件会被覆盖
	f, err := os.OpenFile(uploadDir+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fail(w, err)
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
			fail(w, err)
			return
		} else {
			success(w, output)
			return
		}
	}

	success(w, "upload success")
}

func main() {

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/api/run", ApiHandler)
	r.HandleFunc("/webhook/run/", WebhookHandler)
	r.HandleFunc("/webhook/deploy", DeployHandler)
	r.HandleFunc("/webhook/shell/{filename}", WebhookHandler)
	r.HandleFunc("/upload", UploadHandler)

	// r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(local))))
	r.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	spa := spaHandler{staticPath: htmlDir, indexPath: "/"}
	r.PathPrefix("/").Handler(spa)

	fmt.Printf("Server startup at http://localhost:%s\n", port)

	timeoutNum, _ := strconv.Atoi(timeout)
	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		WriteTimeout: time.Duration(timeoutNum) * time.Second,
		ReadTimeout:  time.Duration(timeoutNum) * time.Second,
		IdleTimeout:  time.Duration(timeoutNum*4) * time.Second,
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
