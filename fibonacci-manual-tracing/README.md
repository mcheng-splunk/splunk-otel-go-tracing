## Golang Manual Instrumentation with Splunk Otel-go

The example is based on the Open Telemetry Golang documentation Fibonacci project. We are using the application but with some modification.

-	The traces will be sent to our O11y platform instead of the trace.txt file.

https://opentelemetry.io/docs/instrumentation/go/getting-started/



## PreRequisite

The appliation has been implemented and tested on   
`go version go1.17.8 darwin/amd64`   
`splunk otel collect 0.45.0`

## To Run

- git clone the `repo`
- go mod download
- go run .