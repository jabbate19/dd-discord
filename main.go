package main

import (
	"os"
	"fmt"
	"net/http"
	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"syscall"
	"os/signal"
	"strings"
)

var (
	token = os.Getenv("TOKEN")
	opsChannel = os.Getenv("OPS_CHANNEL")
)

type Org struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Monitor struct {
	ID          string `json:"id"`
	LastUpdated string `json:"last_updated"`
	EventType   string `json:"event_type"`
	Title       string `json:"title"`
	Date        string `json:"date"`
	Org         Org    `json:"org"`
	Body        string `json:"body"`
}

func main() {
	

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session:", err)
		return
	}

	dg.AddHandler(ready)

	// Open a connection to the Discord gateway
	err = dg.Open()
	if err != nil {
		fmt.Println("Error opening connection:", err)
		return
	}

	// Create a channel to signal bot shutdown
	shutdown := make(chan struct{})

	// Run the bot in a separate Goroutine
	go func() {
		// Add your event handlers and bot logic here

		// Wait for shutdown signal
		<-shutdown

		// Close the Discord session before exiting
		dg.Close()
	}()

	// Wait for a termination signal (e.g., Ctrl+C)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Signal the bot to shut down
	close(shutdown)

}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	r := gin.Default()

	r.POST("/", func(c *gin.Context) {
		var monitor Monitor
		var color int

		// Parse JSON data from request body into the monitor struct
		if err := c.ShouldBindJSON(&monitor); err != nil {
			fmt.Println("Parse Error: " + err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		monitor.Body = strings.Replace(monitor.Body, "%%%", "", -1)
		monitor.Body = strings.Replace(monitor.Body, "- - -", "", -1)
		monitor.Body = strings.Replace(monitor.Body, "\n\n", "\n", -1)

		if strings.Contains(monitor.Title, "Triggered") {
			color = 15548997
		} else if strings.Contains(monitor.Title, "Warn") {
			color = 16776960
		} else if strings.Contains(monitor.Title, "Recovered") {
			color = 5763719
		}

		_, err := s.ChannelMessageSendComplex(opsChannel, &discordgo.MessageSend{
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "Datadog Alerts",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:  monitor.Title,
							Value: monitor.Body,
						},
					},
					Color: color,
				},
			},
		})

		if err != nil {
			fmt.Println("Send Error: " + err.Error())
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Monitor data parsed successfully"})
	})

	r.Run(":8080")
}