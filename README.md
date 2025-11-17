

## Key Features

* HTTP server
  - HTTP GET
  - HTTP POST
* Proxy for handling the connection between client and server.
  - Can only handle GET

## How To Use (building)

From your command line:

```bash
# Clone this repository
$ git clone https://github.com/MHultin/TDA596-Labs.git

# Go into the repository
$ cd Lab1

# navigate to, build, and run the binary file for the proxy
$ cd proxy
$ go build proxy.go
$ ./proxy [port]

# navigate to, build, and run the binary file for the http_server
$ cd ../http_server
$ go build http_server.go
$ ./http_server [port]
```


## How To Use (without building)

From your command line:

```bash
# Clone this repository
$ git clone https://github.com/MHultin/TDA596-Labs.git

# Go into the repository
$ cd Lab1

# navigate to and run the proxy
$ cd proxy
$ go run proxy.go [port]

# navigate to and run server
$ cd server
$ go run http_server.go [port] 
```

> **Note**
An example of how to send a GET request using the proxy.
```bash
$ curl -X GET <server_ip>:<server_port>/<file> -x <proxy_ip>:<proxy_port>
```

## How to build with Docker
One alternative way to run both the proxy and server is using Docker.
In order to run with docker we have to build our images using the supplied docker files. 

```bash
# Build the http-server.
$ cd Lab1/server
$ docker build . -t http-server

# Build the proxt.
$ cd Lab1/proxy
$ docker build . -t proxy

# Run the http server
$ docker run -d -p <host port>:<container port> --env PORT=<container port> --name http-server http-server

# Run the proxy
$ docker run -d -p <host port>:<container port> --env PORT=<container port> --name proxy proxy
```

This is also the way we published our http-server to AWS. By pushing our image to ECR, we could host our images on ECS and reach them from the internet. 

## Description

This applications purpose is to create a server as well as a proxy server, to handle GET and POST requests. Both the server and the proxy will listen to the ports specified in the arguments on startup. A user could make GET requests to either the server or the proxy, but only POST requests through the server. When a GET request has been made through the proxy, the request will be forwarded to the server. This setup opens up multiple potentials for features, such as privacy, administration, or performance improvements.
