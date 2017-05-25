To build osx client
CGO_ENABLED=0 go build -a -installsuffix cgo -o fabric-discovery-linux-client .

To build linux build
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o fabric-discovery-linux-client .


Run with
./fabric-discovery-client :3500 service2 localhost:8085