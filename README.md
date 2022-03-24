# Overview

This repository aims to achieve the following objectives
- Explore how the [Splunk Otel Go](https://github.com/signalfx/splunk-otel-go) works
- Understand how to perform auto and manual instrumentation for Golang application

# Todo Application

The todo application aims to provide a simple Golang Todo application (CRUD) with a MYSQL backend. 

Modifications are made **directly** in the Golang code for instrumentation to happen. 

Take note of the mainly of the 

- `DeleteItem` function
  - The application is modified to use `"github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql"` instead of the original `"github.com/jinzhu/gorm"` library

```
		//For Otel manual instrumentation
		log.Info("id is ", id)
		res, err := db_trace.Exec("DELETE FROM todo_item_models WHERE id=?", id)
		log.Println("res is ", res)
		if err == nil {

			count, err := res.RowsAffected()
			if err == nil {
				/* check count and return true/false */
			}
			log.Println(count, " is DELETED")

		} else {
			log.Println("err ", err)
		}

		// original function using gorm library.
		// log.WithFields(log.Fields{"Id": id}).Info("Deleting TodoItem")
		// todo := &TodoItemModel{}
		// db.First(&todo, id)
		// db.Delete(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": true}`)
```



- `main` function
  - The mux router is modified to generate traces using the `"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"` library
```
	//router := splunkhttprouter.New()
	router := mux.NewRouter()
```


# Fibonacci Application

The Fibonacci application aims to let user understand how to perform manual instrumentation in Golang.

The application is based on the [example](https://opentelemetry.io/docs/instrumentation/go/getting-started/). However, after going through the example, we explore how we will `integrate` the example with our Splunk-Otel-Go to send all the manual traces our the Splunk Observability Cloud instead of the `traces.txt` file in the example.