Dynamic tcp proxy
=======================

Building
------------

Install:

    go get gopkg.in/redis.v3
    go get github.com/fsouza/go-dockerclient

---------
Running redis-tubes:

    go run redis-tubes.go


To change where tcp connections get proxied:

    > redis-cli
    127.0.0.1:6379> sadd 'localhost:9999' 'localhost:80'
    127.0.0.1:6379> srem 'localhost:9999' 'localhost:80'

The redis tubes program looks at to redis to see where incoming connections should get proxied to.  It looks at the key named after the local connection address (default localhost:9999) it listens on, and grabs a random address:port from that set.  To change where connections could get proxied to simply add or remove address:port values to that set.


---------
Running docker-tubes:

    go run docker-tubes.go -i ImageName:Tag

docker-tubes will connect to the docker instance your environment is configured for and look for containers running the image specified.  If that image has exposed ports it will automatically start proxying connections to it.  docker-tubes will automatically notice when a container is started or stopped that matches the image.
