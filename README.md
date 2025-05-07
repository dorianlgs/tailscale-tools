# tailscale-tools

## Module creation
```bash
go mod init github.com/shujink0/tailscale-tools
```

## Build
```bash
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w"
tailscale-tools.exe --port 80 --host incamfus.cifu.dev --apachev 2.4.62.1
```
