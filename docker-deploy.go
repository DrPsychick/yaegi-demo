///usr/bin/env yaegi run -syscall "$0" "$@"; exit
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"io"
	"os"
)

func tarFromDockerfile() io.Reader {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(zw)

	fileinfo, _ := os.Stat("Dockerfile")
	th, _ := tar.FileInfoHeader(fileinfo, "Dockerfile")
	tw.WriteHeader(th)

	file, err := os.Open("Dockerfile")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// write file contents
	if _, err = io.Copy(tw, file); err != nil {
		panic(err)
	}
	tw.Close()
	zw.Close()

	br := bytes.NewReader(buf.Bytes())

	// TEST: write a .tgz to check if it's ok
	//out, _ := os.Create("Docker.tgz")
	//defer out.Close()
	//if _, err = io.Copy(out, br); err != nil {
	//	panic(err)
	//}
	//br.Seek(0,0)

	return br
}

func tarFromFile() io.Reader {
	file, err := os.Open("Docker.tgz")
	if err != nil {
		panic(err)
	}
	return file
}

func main() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	// it works!
	// proof: curl --unix-socket /var/run/docker.sock -H "Content-Type: application/x-tar" --data-binary @Docker.tgz -XPOST http://1.40/build
	buildArgs := make(map[string]*string)
	res, err := cli.ImageBuild(context.Background(), tarFromDockerfile(), types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{"doo"},
		NoCache:    true,
		Remove:     true,
		BuildArgs:  buildArgs,
	})
	if err != nil {
		fmt.Println(err)
	} else {
		defer res.Body.Close()
	}

	var body bytes.Buffer
	b := make([]byte, 1024)
	for {
		n, _ := res.Body.Read(b)
		if n == 0 {
			break
		}
		body.Write(b[:n])
	}
	fmt.Printf("Building finished\n%s", string(body.Bytes()))

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{true, true, true, true, "", "", 0, filters.Args{}})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}
}
