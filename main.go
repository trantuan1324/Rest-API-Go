package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type TodoItem struct {
	Id          int        `json:"id" gorm:"column:id;"`
	Title       string     `json:"title" gorm:"column:title;"`
	Description string     `json:"description" gorm:"column:description;"`
	Status      string     `json:"status" gorm:"column:status;"`
	CreatedAt   *time.Time `json:"created_at" gorm:"column:created_at;"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" gorm:"column:updated_at;"`
}

func (TodoItem) TableName() string {
	return "todo_items"
}

type TodoItemCreation struct {
	Id          int    `json:"-" gorm:"column:id;"`
	Title       string `json:"title" gorm:"column:title;"`
	Description string `json:"description" gorm:"column:description;"`
}

func (TodoItemCreation) TableName() string {
	return "todo_items"
}

type TodoItemUpdate struct {
	Title       *string `json:"title" gorm:"column:title;"`
	Description *string `json:"description" gorm:"column:description;"`
	Status      string  `json:"status" gorm:"column:status;"`
}

func (TodoItemUpdate) TableName() string {
	return "todo_items"
}

func main() {
	dsn := os.Getenv("DB_CNN_STR")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	r := gin.Default()

	/*
		CRUD
		POST /v1/items (create a new item)
		GET /v1/items(list items) v1/items?page=1
		GET /v1/items/:id (get item detail by id)
		(PUT || PATCH) /v1/items/:id (update an item by id)
		DELETE /v1/items/:id (delete item by id)
	*/

	v1 := r.Group("/v1")

	{
		items := v1.Group("/items")
		{
			items.POST("", CreateItem(db))
			items.GET("", GetItems(db))
			items.GET("/:id", GetItemByID(db))
			items.PATCH("/:id", UpdateItemByID(db))
			items.DELETE("/:id", DeleteItemByID(db))
		}
	}

	r.GET("/ping", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{
			"message": "ping",
		})
	})

	if err := r.Run(":3000"); err != nil {
		log.Fatalln(err)
	}
}

type Paging struct {
	Page  int   `json:"page" form:"page"`
	Limit int   `json:"limit" form:"page"`
	Total int64 `json:"total" form:"page"`
}

func CreateItem(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var data TodoItemCreation
		if err := ctx.ShouldBind(&data); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := db.Create(&data).Error; err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"data": data.Id,
		})
	}
}

func GetItems(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var data Paging

		if err := ctx.ShouldBind(&data); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		var listItems []TodoItem

		if err := db.Order("id desc").Find(&listItems).Error; err != nil {
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"data": listItems,
		})
	}
}

func GetItemByID(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var data TodoItem

		id, err := strconv.Atoi(ctx.Param("id"))

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := db.First(&data, "id = ?", id).Error; err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"data": data,
		})
	}
}

func UpdateItemByID(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var data TodoItemUpdate

		id, err := strconv.Atoi(ctx.Param("id"))

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		if err := ctx.ShouldBind(&data); err != nil {
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}

		if err := db.Where("id = ?", id).Updates(&data).Error; err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"data": true,
		})

	}
}

func DeleteItemByID(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))

		if err != nil {
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}

		if err := db.Table(TodoItem{}.TableName()).Where("id = ?", id).Updates(map[string]interface{}{
			"status": "Deleted",
		}).Error; err != nil {
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": true,
		})
	}
}
