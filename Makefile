all: build

install:
	go mod download
deploy: 
	curl -H Content-Type:application/json -X POST \
	-d '{"script":"test.sh","argument":"alpha"}' \
	http://localhost:8000/webhook/shell/test.sh -i
upload: 
	curl -F 'file=@./test.txt' http://localhost:8000/upload -v
start: 
	env GDK_UPLOAD_DIR=./uploads/ GDK_SCRIPT_DIR=./shells/ go run main.go
build: 
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/gdk main.go

zip:
	zip -r gdk.zip bin/gdk 