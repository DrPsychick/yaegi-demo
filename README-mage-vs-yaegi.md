## `mage` vs `yaegi`


### Install

mage:
```shell
go get -u github.com/magefile/mage
```

yaegi:
```shell
go get -u github.com/traefik/yaegi/cmd/yaegi
```

### How to run
The first line of your `.go` file

mage:
```shell
// +build mage

# ^^ an empty line after is required!
```

yaegi:
```shell
///usr/bin/env yaegi run "$0" "$@"; exit
```

### Entrypoint

mage:
* Any exported function without arguments that returns nothing or an error
* see https://magefile.org/magefiles/
```go
func Build() {}

func Install() error {}
```

yaegi:
* Runs the `main` function, just as if it was compiled
```go
func main() {}
```

## `mage` vs `yaegi`
### Mage
* **+** multiple targets within one file
* **+** more stable than `yaegi`
* **+** larger community
* **-** cannot be compiled

#### Use cases
* replacement for `Makefile`


### Yaegi
* **+** can be compiled "as-is" (`go build conf_update.go`)
* **-** not as stable as `mage` (`docker-deploy.go`, using libraries does NOT work with yaegi)

#### Use cases
* small, standalone Go scripts, that can optionally be compiled
  * as a replacement for `shell` scripts
* for use with a service:
  * `yaegi` 28M vs. `conf_update` 2.3M compiled...
  * only really useful, if 
    * used independent of a build and deploy pipeline
    * there is a use case for changing the scripts "live"/independent of a deploy
    * e.g. 
      * testing a client library during implementation (instead of running your service, you run the script)
      * executing parts of your code/library to test it
