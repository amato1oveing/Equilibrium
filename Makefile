# makefile文件
# Created Time: 2022年10月03日 星期一 00时03分00秒

APP = Equilibrium
all:build windows linux macos
build:
	go build -o $(APP)
windows:
        GOOS=windows GOARCH=amd64 go build -o ${APP}.exe cmd.go
linux:
        GOOS=linux GOARCH=amd64 go build -o ${APP} cmd.go
macos:
		GOOS=darwin GOARCH=amd64 go build -o ${APP} cmd.go