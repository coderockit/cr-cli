#/bin/sh
rm $GOPATH/bin/crlinux
GOOS=linux GOARCH=386 go build -o $GOPATH/bin/crlinux main.go
#scp $GOPATH/bin/crlinux nsivraj@www:/home/nsivraj/forge/bin/cr
