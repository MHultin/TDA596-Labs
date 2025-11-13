

## Key Features

* HTTP server
  - HTTP GET
  - HTTP POST
* Proxy for handling the connection between client and server.
  - Can only handle GET


## How To Use

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

## Credits

This software uses the following open source packages:


## Support

## You may also like...

## License





