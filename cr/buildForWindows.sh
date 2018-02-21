#/bin/sh
rm $GOPATH/bin/cr.exe
GOOS=windows GOARCH=386 go build -o $GOPATH/bin/cr.exe main.go
scp $GOPATH/bin/cr.exe nsivraj@zytecoinc:/home/nsivraj/forge/bin
