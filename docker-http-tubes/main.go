package main

import (
  "flag"
  "github.com/cconstantine/gotubes/logging"
  "github.com/cconstantine/gotubes/container-listener"
  "net/http"
  "net/http/httputil"
)

var image_name = flag.String("i", "training/webapp:latest", "Proxy connections to Image")
var proxyPort = flag.String("p",  "5000", "Port to proxy")
var verbose = flag.Bool("v", false, "display server actions")

func NewMultipleHostReverseProxy(container_provider containers.ContainerPorts) *httputil.ReverseProxy {
        director := func(req *http.Request) {
          target, _ := container_provider.GetRandomPort()

          container_provider.GetRandomPort()
          req.URL.Host = target
        }
        return &httputil.ReverseProxy{Director: director}
}

func main() {
  flag.Parse()

  logging.Init(*verbose)

  container_provider := containers.NewContainerPorts(*image_name, *proxyPort)

  proxy := NewMultipleHostReverseProxy(*container_provider)
  http.ListenAndServe(":9090", proxy)

}