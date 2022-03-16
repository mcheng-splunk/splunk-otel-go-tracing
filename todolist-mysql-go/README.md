# Golang Auto Instrumentation with Splunk Otel-go

The example is a simple Todo CRUD application using MYSQL as a backend database. 

The application was originally written using the following 2 libraries   
`Gorilla mux`   
`Gorm`

We will proceed to make modification to generate the traces and send to O11y.

## Otel Supported Libraries

When instrumenting any Golang application, we need to know what are the Golang libraries the application is using. The following list of Golang libraries are Otel ready.

https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation   
https://github.com/signalfx/splunk-otel-go#library-instrumentation

***Note*** that Gom is not within the list of Otel readied library, hence the application will be modified to use the splunkmysql library


## PreRequisite

The appliation has been implemented and tested on   
`go version go1.17.8 darwin/amd64`   
`splunk otel collect 0.45.0`


## To Run

- git clone the `repo`
- go mod download
- cd todolsit-mysql-go/db
- docker-compose up
- go run todolist.go