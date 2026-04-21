package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"REST_Server/taskstore"

	"github.com/gin-gonic/gin"
)

type TaskServer struct {
	store *taskstore.TaskDB
}

func NewTaskServer() *TaskServer {
	store := taskstore.New()

	return &TaskServer{store: store}
}

func (ts *TaskServer) createTaskHandler(c *gin.Context) {
	type RequestTask struct {
		Text string    `json:"text"`
		Due  time.Time `json:"due"`
	}

	var rt RequestTask
	if err := c.ShouldBindJSON(&rt); err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	id := ts.store.CreateTask(rt.Text, rt.Due)
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (ts *TaskServer) getTaskHandler(c *gin.Context) {
	ids := c.Param("id")
	id, err := strconv.Atoi(ids)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	task, err := ts.store.GetTask(id)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		log.Println(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": task.Id, "text": task.Text, "due": task.Due})
}

func (ts *TaskServer) getAllTasksHandler(c *gin.Context) {
	allTasks, err := ts.store.GetAllTasks()
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}
	c.JSON(http.StatusOK, allTasks)
}

func (ts *TaskServer) deleteAllTasksHandler(c *gin.Context) {
	msg, err := ts.store.DeleteAllTasks()
	if err != nil {
		c.String(http.StatusPreconditionFailed, err.Error())
		log.Println(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": msg})
}

func (ts *TaskServer) deleteTaskHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		log.Println(err)
		return
	}

	msg, err := ts.store.DeleteTask(id)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		log.Println(err)
		return
	}
	c.String(http.StatusOK, msg)
}

func (ts *TaskServer) dueHandler(c *gin.Context) {
	badRequestError := func() {
		c.String(http.StatusBadRequest, "expect /due/<year>/<month>/<day>, got %v", c.FullPath())
	}

	year, err := strconv.Atoi(c.Param("year"))
	if err != nil {
		badRequestError()
		return
	}

	month, err := strconv.Atoi(c.Param("month"))
	if err != nil || month < int(time.January) || month > int(time.December) {
		badRequestError()
		return
	}

	day, err := strconv.Atoi(c.Param("day"))
	if err != nil {
		badRequestError()
		return
	}

	tasks, err := ts.store.GetTasksByDueDate(year, time.Month(month), day)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		log.Println(err)
		return
	}
	c.JSON(http.StatusOK, tasks)
}

func main() {
	router := gin.Default()
	server := NewTaskServer()
	defer server.store.Close()

	router.POST("/task/", server.createTaskHandler)
	router.GET("/task/", server.getAllTasksHandler)
	router.DELETE("/task/", server.deleteAllTasksHandler)
	router.GET("/task/:id", server.getTaskHandler)
	router.DELETE("/task/:id", server.deleteTaskHandler)
	router.GET("/due/:year/:month/:day", server.dueHandler)

	router.Run("localhost:3030")
}
