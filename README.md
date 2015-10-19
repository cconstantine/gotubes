Dynamic tcp proxy
=======================

Building
------------

Install:

    go get gopkg.in/redis.v3

Run:

    go run proxy.go

Also this other thing
----------

Make sure your local reds db has a target for where you're proxying from:

    > redis-cli
    127.0.0.1:6379> sadd 'localhost:9999' 'localhost:80'