### Use `docker` library to build & push docker images

```
go mod vendor
./docker-deploy.go
```


#### Initialize `go.mod`
* comment out the shebang from the .go file, so that `go fmt` succeeds
```
go mod init
go get github.com/...
# -> generates the go.mod
```
* now you can add the shebang again to the .go file