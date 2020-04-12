# Go-Deploy-Kit

是一个用 golang 开发的 用于部署一些简单项目的工具

使用方法

- 设置2个环境变量
  - [x] GDK_UPLOAD_DIR 存放上传的项目目录地址，最好是一个绝对路径，并且以 `/` 结尾
  - [x] GDK_SCRIPT_DIR 存放 shell 脚本的目录

- 上传项目文件的方法

  `curl -F 'file=@./api.jar' http://ip:8000/upload -v` 其中的 `'file=@./api.jar'` file 后面的部分填写需要上传的本地文件，`ip` 为启动服务的地址

- 执行 shell 的方法

  `curl -H "Content-Type:application/json" -X POST -d '{"script":"test.sh","argument":"aaa"}' http://ip:8000/webhook/run/deploy -i`  其中 `{\"script\":\"test.sh\",\"argument\":\"aaa\"}` 中 script 是对应的shell文件名，要求是带后缀的，比如 `test.sh` 或者 `test.py` 等等，`argument` 是后面需要执行带有的参数，直接以字符串的方式并接起来，比如 `-Max=100 -Min=1`,最后会执行 `test.sh -Max=100 -Min=1` 当然需要在shell中自己解析这些参数信息。


## Reference

- https://www.kimsereylam.com/angular/2019/04/26/angular-webpack-proxy-dev-server.html#proxy-dev-server

# RoadMap

## v0.2.0

- Feature

[ ] Add a ui for it, with Angular8
[ ] Support websoket