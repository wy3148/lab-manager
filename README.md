# lab-manager

##Before build, to resolve the /vendor issue,

running following two commands

```
1. 
rm -rf $GOPATH/src/github.com/docker/docker/vendor/github.com/docker/go-connections

2.
go get github.com/pkg/errors

3. go build
```
