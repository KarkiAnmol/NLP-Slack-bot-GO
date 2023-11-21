package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/krognol/go-wolfram"
	"github.com/shomali11/slacker"
	"github.com/tidwall/gjson"
	witai "github.com/wit-ai/wit-go/v2"
)

var wolframClient *wolfram.Client

// printCommandEvents is a goroutine that prints command events for debugging purposes.
func printCommandEvents(analyticsChannel <-chan *slacker.CommandEvent) {
	for event := range analyticsChannel {
		fmt.Println("Command Events	")
		fmt.Println(event.Timestamp)
		fmt.Println(event.Command)
		fmt.Println(event.Parameters)
		fmt.Println(event.Event)
		fmt.Println()
	}
}

func main() {
	// Load environment variables from the .env file
	godotenv.Load(".env")

	// Create a new Slacker bot instance with Slack API tokens
	bot := slacker.NewClient(os.Getenv("SLACK_BOT_TOKEN"), os.Getenv("SLACK_APP_TOKEN"))

	// Create Wit.ai client and Wolfram Alpha client
	client := witai.NewClient(os.Getenv("WIT_AI_TOKEN"))
	wolframClient = &wolfram.Client{AppID: os.Getenv("WOLFRAM_APP_ID")}

	// Start a goroutine to print command events for debugging
	go printCommandEvents(bot.CommandEvents())

	// Define a command for querying Wolfram Alpha
	bot.Command("query - <message>", &slacker.CommandDefinition{
		Description: "Send any question to Wolfram Alpha",
		Examples:    []string{"who was the president of Nepal in 2017?"},
		Handler: func(bc slacker.BotContext, r slacker.Request, w slacker.ResponseWriter) {
			// Extract the message parameter from the command
			query := r.Param("message")
			fmt.Println(query)

			// Use Wit.ai to parse the message and extract relevant information
			msg, _ := client.Parse(&witai.MessageRequest{
				Query: query,
			})

			// Extract the value from Wit.ai response
			data, _ := json.MarshalIndent(msg, "", "    ")
			rough := string(data[:])
			value := gjson.Get(rough, "entities.wit$wolfram_search_query:wolfram_search_query.0.value")
			answer := value.String()

			// Query Wolfram Alpha for a spoken answer
			res, err := wolframClient.GetSpokentAnswerQuery(answer, wolfram.Metric, 1000)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(value)

			// Reply to the user in the Slack channel
			w.Reply(res)
		},
	})

	// Set up a context and listen for incoming events
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := bot.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}
}
