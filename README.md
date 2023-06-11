# Protomesh


## Building example AWS Lambda

To build the example AWS Lambda that handles the `PingPong.DoPing` gRPC call use the following commands.

```bash
cd examples/lambda
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main .
rm -rf main.zip
zip main.zip main
```