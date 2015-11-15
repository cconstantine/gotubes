package main

import (
  "flag"
  "math/rand"
  "errors"
	"os"
  "sync"
  "github.com/cconstantine/gotubes/proxy"
  "github.com/cconstantine/gotubes/logging"
  "github.com/fsouza/go-dockerclient"
)
var image_name = flag.String("i", "training/webapp:latest", "Proxy connections to Image")

var proxyPort = flag.String("p",  "5000", "Port to proxy")

var verbose = flag.Bool("v", false, "display server actions")

func main() {
  flag.Parse()

  logging.Init(*verbose)

  container_ports := NewContainerPorts()

  s := &proxy.Server{
    LocalAddr: "0.0.0.0:9999",
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


  client, err := docker.NewClientFromEnv()
	if err != nil {
		logging.Error.Printf(err.Error())
		os.Exit(1)
	}

  go container_ports.listenForEvents(client)

  containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		logging.Error.Printf(err.Error())
		os.Exit(1)
	}

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

	connection_string := container.NetworkSettings.IPAddress + ":" + *proxyPort
	cp.container_ports[container.ID] = connection_string

	logging.Info.Printf(
		"Adding a container %s: %s <-> %s\n",
		container.ID[:12], "0.0.0.0:9999", connection_string)
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
