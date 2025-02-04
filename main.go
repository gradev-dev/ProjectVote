package main

import (
	"Planning_poker/app"
	"Planning_poker/app/logging"
	"Planning_poker/app/models"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
)

func main() {
	env, envErr := app.GetEnv()
	if envErr != nil {
		panic(envErr)
	}

	if err := logging.InitLogging(env); err != nil {
		panic(err)
	}

	// Tworzenie routera
	r := gin.New()
	r.Use(
		sloggin.New(slog.Default().WithGroup("server")),
		gin.Recovery(),
	)

	// Ładowanie szablonów HTML
	r.LoadHTMLGlob("templates/*")

	// Obsługa plików statycznych
	r.Static("/css", "./static/css")
	r.Static("/js", "./static/js")

	r.GET("/image-proxy", func(c *gin.Context) {
		app.ImageProxyHandler(c.Writer, c.Request)
	})

	// Strona główna (tworzenie pokoju)
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "create.html", gin.H{
			"AppName": env.App,
			"Ws":      fmt.Sprintf("ws://%v/ws", env.Url),
		})
	})

	// Strona dołączania do pokoju
	r.GET("/join/:roomId", func(c *gin.Context) {
		roomID := c.Param("roomId")
		roomExists, room := app.GetExistsRoom(roomID)

		if !roomExists {
			c.HTML(http.StatusOK, "error.html", gin.H{
				"AppName": env.App,
			})
			return
		}

		c.HTML(http.StatusOK, "join.html", gin.H{
			"AppName":  env.App,
			"RoomID":   roomID,
			"RoomName": room.Name,
			"Ws":       fmt.Sprintf("ws://%v/ws", env.Url),
		})
	})

	// Strona podsumowania do pokoju
	r.GET("/summary/:roomId", func(c *gin.Context) {
		roomID := c.Param("roomId")
		roomExists, room := app.GetExistsRoom(roomID)

		if !roomExists {
			c.HTML(http.StatusOK, "error.html", gin.H{
				"AppName": env.App,
			})
			return
		}

		c.HTML(http.StatusOK, "summary.html", gin.H{
			"AppName":  env.App,
			"RoomID":   roomID,
			"RoomName": room.Name,
			"Ws":       fmt.Sprintf("ws://%v/ws", env.Url),
		})
	})

	r.GET("/history/:roomId", func(c *gin.Context) {
		roomID := c.Param("roomId")
		summary := c.Query("summary")

		roomExists, room := app.GetExistsRoom(roomID)
		if !roomExists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
			return
		}

		if room.RoomHistory == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Room history is empty"})
			return
		}

		if summary != "" {
			roomHistories := []models.RoomHistory{}
			for _, h := range room.RoomHistory {
				taskKey := h.Task
				if taskKey != "" {
					jiraURL := fmt.Sprintf("%s/rest/api/2/issue/%s?fields=id", env.JiraUrl, taskKey)
					req, err := http.NewRequest("GET", jiraURL, nil)
					if err != nil {
						continue
					}
					req.Header.Set("Authorization", "Bearer "+env.JiraAPIToken)
					req.Header.Set("Content-Type", "application/json")

					client := &http.Client{}
					resp, err := client.Do(req)
					if err != nil || resp.StatusCode != http.StatusOK {
						continue
					}
					defer resp.Body.Close()

					var task models.JiraResponse
					if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
						continue
					}

					h.JiraTaskUrl = fmt.Sprintf("%s/secure/CreateWorklog!default.jspa?id=%s", env.JiraUrl, task.Id)
				}

				roomHistories = append(roomHistories, h)
			}

			c.JSON(http.StatusOK, roomHistories)
			return
		}

		c.JSON(http.StatusOK, room.RoomHistory)
	})

	// Strona głosowania
	r.GET("/voting/:roomId", func(c *gin.Context) {
		roomID := c.Param("roomId")
		roomExists, room := app.GetExistsRoom(roomID)

		if !roomExists {
			c.HTML(http.StatusOK, "error.html", gin.H{
				"AppName": env.App,
				"RoomID":  roomID,
			})
			return
		}

		c.HTML(http.StatusOK, "voting.html", gin.H{
			"AppName":    env.App,
			"RoomID":     roomID,
			"RoomName":   room.Name,
			"RoomMethod": room.RoomMethod,
			"Ws":         fmt.Sprintf("ws://%v/ws", env.Url),
		})
	})

	healthcheck := r.Group("/healthcheck")
	{
		healthcheck.GET("/liveness", app.GetLiveness)
		healthcheck.GET("/status", app.GetStatus)
		healthcheck.GET("/readiness", app.GetReadiness)
	}

	tasks := r.Group("/tasks")
	{
		tasks.GET("/search/:taskKey", app.GetTask)
		tasks.GET("/detail/:taskKey", app.GetTaskDetails)
		tasks.PUT("/save", app.SaveTask)
	}

	// Obsługa WebSocket
	r.GET("/ws", func(c *gin.Context) {
		app.HandleWebSocket(c.Writer, c.Request)
	})

	go startCron()
	// Start serwera
	log.Println("Starting server on :4009")
	if err := r.Run(":4009"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func startCron() {
	ticker := time.NewTicker(45 * time.Minute)
	defer ticker.Stop()

	for {
		<-ticker.C
		app.CleanEmptyRooms()
	}
}
