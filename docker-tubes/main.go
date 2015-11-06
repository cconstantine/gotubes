package main

import (
  "flag"
  "math/rand"
  "errors"
  "sync"
  "github.com/cconstantine/gotubes/proxy"
  "github.com/cconstantine/gotubes/logging"
  "github.com/fsouza/go-dockerclient"
)
var image_name = flag.String("i", "training/webapp:latest", "image to run")

var remoteAddr = flag.String("r", "localhost",      "remote address")

var localAddr = flag.String("l",  "localhost:9999", "local address")
var verbose = flag.Bool("v", false, "display server actions")

func main() {

  flag.Parse()

  logging.Init(*verbose)

  container_ports := NewContainerPorts()

  s := &proxy.Server{
    LocalAddr: *localAddr,
    Verbose:    true,
    Ports: container_ports,
  }
  s.Run()
}

type ContainerPorts struct {
  container_ports map[string]string
  mutx  sync.Mutex
}

func NewContainerPorts() *ContainerPorts {
 container_ports := &ContainerPorts{container_ports: make(map[string]string)}


  client, _ := docker.NewClientFromEnv()

  go container_ports.listenForEvents(client)

  containers, _ := client.ListContainers(docker.ListContainersOptions{})

  for _, event := range containers {
     container, _ := client.InspectContainer(event.ID)
     container_ports.addContainer(container)
  }
  return container_ports
}

func (cp* ContainerPorts) GetRandomPort() (string, error) {
  defer cp.mutx.Unlock()
  cp.mutx.Lock()
  
  if len(cp.container_ports) == 0 {
    return "", errors.New("No ports available")
  }

  index := rand.Intn(len(cp.container_ports) )
  ret := ""
  for _, port := range cp.container_ports {
    if index == 0 {
      ret = port
      break

    }

    index -= 1
  }
  return ret, nil
}

func (container_ports *ContainerPorts) listenForEvents(client *docker.Client) {
  events_channel := make(chan *docker.APIEvents)
  client.AddEventListener(events_channel)

  for {
    event := <- events_channel

    if event.Status == "start" {
      container, _ := client.InspectContainer(event.ID)
      container_ports.addContainer(container)
    } else if event.Status == "die" {
      container, _ := client.InspectContainer(event.ID)

      container_ports.removeContainer(container)
    }
  }
}


func (cp* ContainerPorts) addContainer(container *docker.Container) {
  defer cp.mutx.Unlock()
  cp.mutx.Lock()

  if container.Config.Image != *image_name {
    return
  }

  for _,b := range container.NetworkSettings.Ports {
    port := *remoteAddr + ":" + b[0].HostPort
    cp.container_ports[container.ID] = port

    logging.Info.Printf("Adding a container %s: %s <-> %s\n", container.ID[:12], *localAddr, port)
  } 
  
}

func (cp* ContainerPorts) removeContainer(container *docker.Container) {
  defer cp.mutx.Unlock()
  cp.mutx.Lock()

  if container.Config.Image != *image_name {
    return
  }

  logging.Info.Printf("Removing container %s\n", container.ID[:12])
  
  delete(cp.container_ports, container.ID)
}
