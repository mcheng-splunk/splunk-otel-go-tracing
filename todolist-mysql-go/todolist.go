package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	// _ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"

	//_ "github.com/jinzhu/gorm/dialects/mysql"

	"github.com/rs/cors"
	"github.com/signalfx/splunk-otel-go/distro"

	// Using Splunksql tracing library
	"github.com/signalfx/splunk-otel-go/instrumentation/database/sql/splunksql"
	_ "github.com/signalfx/splunk-otel-go/instrumentation/github.com/go-sql-driver/mysql/splunkmysql"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

var db, _ = gorm.Open("mysql", "root:root@/todolist?charset=utf8&parseTime=True&loc=Local")

// For Otel manual instrumentation
var db_trace, _ = splunksql.Open("mysql", "root:root@/todolist?charset=utf8&parseTime=True&loc=Local")

type TodoItemModel struct {
	Id          int `gorm:"primary_key"`
	Description string
	Completed   bool
}

func CreateItem(w http.ResponseWriter, r *http.Request) {
	description := r.FormValue("description")
	log.WithFields(log.Fields{"description": description}).Info("Add new TodoItem. Saving to database.")
	todo := &TodoItemModel{Description: description, Completed: false}
	db.Create(&todo)
	result := db.Last(&todo)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

}

func UpdateItem(w http.ResponseWriter, r *http.Request) {
	// Get URL parameter from mux
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	log.Info("id is ", id)
	// Test if the TodoItem exist in DB
	err := GetItemByID(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": false, "error": "Record Not Found"}`)
	} else {
		completed, _ := strconv.ParseBool(r.FormValue("completed"))
		log.WithFields(log.Fields{"Id": id, "Completed": completed}).Info("Updating TodoItem")
		todo := &TodoItemModel{}
		db.First(&todo, id)
		todo.Completed = completed
		db.Save(&todo)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"updated": true}`)
	}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {

	// Get URL parameter from mux
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	// Test if the TodoItem exist in DB
	err := GetItemByID(id)
	if err == false {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"deleted": false, "error": "Record Not Found"}`)
	} else {
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
	}
}

func GetItemByID(Id int) bool {
	todo := &TodoItemModel{}
	result := db.First(&todo, Id)
	if result.Error != nil {
		log.Warn("TodoItem not found in database")
		return false
	}
	return true
}

func GetCompletedItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get completed TodoItems")
	completedTodoItems := GetTodoItems(true)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(completedTodoItems)
}

func GetIncompleteItems(w http.ResponseWriter, r *http.Request) {
	log.Info("Get Incomplete TodoItems")
	IncompleteTodoItems := GetTodoItems(false)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(IncompleteTodoItems)
}

func GetTodoItems(completed bool) interface{} {
	var todos []TodoItemModel
	TodoItems := db.Where("completed = ?", completed).Find(&todos).Value
	return TodoItems
}

func Healthz(w http.ResponseWriter, r *http.Request) {
	log.Info("API Health is OK")
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, `{"alive", true}`)
}

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetReportCaller(true)

	// Configured to send our trace to Otel Agent residing in our network.
	os.Setenv("OTEL_RESOURCE_ATTRIBUTES", "service.name=go-crud,service.version=1.0.0,deployment.environment=development")
	os.Setenv("OTEL_LOG_LEVEL", "debug")
	os.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", "http://192.168.20.34:14268/api/traces")
	os.Setenv("OTEL_TRACES_EXPORTER", "jaeger-thrift-splunk")

	// If we are going to send our trace directly to O11y
	//os.Setenv("OTEL_EXPORTER_JAEGER_ENDPOINT", "https://ingest.us1.signalfx.com/v2/trace")
	//os.Setenv("SPLUNK_ACCESS_TOKEN", "O11y_token")
}

func main() {

	sdk, err := distro.Run()
	if err != nil {
		panic(err)
	}

	// Ensure all spans are flushed before the application exits.
	defer func() {
		if err := sdk.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}()

	OTEL_LOG_LEVEL := os.Getenv("OTEL_LOG_LEVEL")
	OTEL_EXPORTER_OTLP_TRACES_ENDPOINT := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	fmt.Printf("Display environment variables OTEL_LOG_LEVEL: %s, OTEL_EXPORTER_OTLP_TRACES_ENDPOINT: %s\n", OTEL_LOG_LEVEL, OTEL_EXPORTER_OTLP_TRACES_ENDPOINT)

	defer db.Close()
	defer db_trace.Close()
	db.Debug().DropTableIfExists(&TodoItemModel{})
	db.Debug().AutoMigrate(&TodoItemModel{})

	log.Info("Starting TodoList API server")

	//router := splunkhttprouter.New()
	router := mux.NewRouter()

	// For Otel manual instrumentation
	router.Use(otelmux.Middleware("my-server"))
	router.HandleFunc("/healthz", Healthz).Methods("GET")
	router.HandleFunc("/todo-completed", GetCompletedItems).Methods("GET")
	router.HandleFunc("/todo-incomplete", GetIncompleteItems).Methods("GET")
	router.HandleFunc("/todo", CreateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", UpdateItem).Methods("POST")
	router.HandleFunc("/todo/{id}", DeleteItem).Methods("DELETE")

	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./ui/static/")))

	handler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "POST", "DELETE", "PATCH", "OPTIONS"},
	}).Handler(router)

	http.ListenAndServe(":8000", handler)

}
