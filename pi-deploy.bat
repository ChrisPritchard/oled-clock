go env -w GOOS=linux GOARCH=arm64 CGO_ENABLED=0
go build -o oo ./cmd/display/
go env -u GOOS GOARCH CGO_ENABLED

ssh pizero2w "ps aux | grep ./oo$ | cut -d' ' -f8 | xargs kill"
scp .\oo pizero2w:
ssh pizero2w "chmod +x oo && ./oo"
del oo