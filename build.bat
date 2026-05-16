go env -w GOOS=linux GOARCH=arm64 CGO_ENABLED=0
go build -o oo .
go env -u GOOS GOARCH CGO_ENABLED

scp .\oo pizero2w:
ssh pizero2w "chmod +x oo && ./oo"
del oo