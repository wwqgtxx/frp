cd %~dp0
set CGO_ENABLED=0
set GOARCH=amd64
set GOOS=linux
go build -trimpath -ldflags "-s -w" -o bin/frpc-linux-amd64 ./cmd/frpc
go build -trimpath -ldflags "-s -w" -o bin/frps-linux-amd64 ./cmd/frps
set GOARCH=386
set GOOS=linux
go build -trimpath -ldflags "-s -w" -o bin/frpc-linux-386 ./cmd/frpc
go build -trimpath -ldflags "-s -w" -o bin/frps-linux-386 ./cmd/frps
set GOARCH=amd64
set GOOS=windows
go build -trimpath -ldflags "-s -w" -o bin/frpc-windows-amd64.exe ./cmd/frpc
go build -trimpath -ldflags "-s -w" -o bin/frps-windows-amd64.exe ./cmd/frps
set GOARCH=386
set GOOS=windows
go build -trimpath -ldflags "-s -w" -o bin/frpc-windows-386.exe ./cmd/frpc
go build -trimpath -ldflags "-s -w" -o bin/frps-windows-386.exe ./cmd/frps
pause