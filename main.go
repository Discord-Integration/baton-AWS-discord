package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

// Main bot process
func main() {

	// load ENV file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get discord token from ENV file
	token := os.Getenv("DISCORD_TOKEN")

	// Init bot and catch error
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}

	dg.AddHandler(messageCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		panic(err)
	}

	fmt.Println("Bot is now running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

// Message handling function
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore messages from bot himself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// If content has no information find the last chat message
	if m.Content == "" {
		chanMsgs, err := s.ChannelMessages(m.ChannelID, 1, "", "", m.ID)
		if err != nil {
			fmt.Println("error getting message: ", err)
			return
		}
		m.Content = chanMsgs[0].Content
		m.Attachments = chanMsgs[0].Attachments

	}

	// Sends results of user search to discord
	if strings.Contains(m.Content, "search") {

		// Split the user out of the entire string
		parts := strings.Split(m.Content, "search: ")
		query := parts[1]

		search := awsBatonUserSearch(query)

		if len(search) != 0 {
			fmt.Println(awsBatonUserSearch(query)[1])
			// If user found notify discord
			s.ChannelMessageSend(m.ChannelID, query+" was found with a resource type of "+search[1]+"!")
		} else {
			// If user not found notify discord
			s.ChannelMessageSend(m.ChannelID, query+" not found!")
		}
	} else if strings.Contains(m.Content, "total") { // Sends total count of users to discord

		// Convert int to string
		numUsers := strconv.FormatInt(int64(total()), 10)

		// Notify num users back to channel
		s.ChannelMessageSend(m.ChannelID, numUsers+" users!")

	} else if m.Content == "help" { // Sends list of commands to discord
		s.ChannelMessageSend(m.ChannelID, "'search: <USER>' Checks is user/role exists.")
		s.ChannelMessageSend(m.ChannelID, "'total' Returns total number of users.")

	}

}

// Format of DisplayName within the baton json response
type AWSResources struct {
	Resources []struct {
		Resource struct {
			DisplayName string `json:"displayName"`
		} `json:"resource"`
	} `json:"resources"`
}

// format of displayname and resource type within the baton json response
type AWSResourcesCombined struct {
	Resources []struct {
		Resource struct {
			DisplayName string `json:"displayName"`
			ID          struct {
				ResourceType string `json:"resourceType"`
			} `json:"id"`
		} `json:"resource"`
	} `json:"resources"`
}

func awsBatonUserSearch(user string) []string {
	// run the `baton resources` command and capture the JSON output
	out, err := exec.Command("baton", "resources", "-o", "json").Output()
	if err != nil {
		log.Fatal(err)
	}

	// convert the output to a string
	response := string(out)

	// parse the JSON input
	var resourcesCombined AWSResourcesCombined
	err = json.Unmarshal([]byte(response), &resourcesCombined)
	if err != nil {
		fmt.Println(err)
	}

	// iterate over the resources in the struct and return the display name and resource type if the display name matches the given user
	for _, r := range resourcesCombined.Resources {
		if r.Resource.DisplayName == user {
			return []string{r.Resource.DisplayName, r.Resource.ID.ResourceType}
		}
	}

	// if the user was not found, return an empty slice
	return []string{}
}

// Returns total count of all users
func total() int {

	_, err := exec.Command("baton-aws").Output()

	if err != nil {
		log.Fatal(err)
	}

	out, err := exec.Command("baton", "resources", "-o", "json").Output()

	if err != nil {
		log.Fatal(err)
	}

	// Convert Baton AWS response into a string
	response := string(out)

	// Process json
	var resources AWSResources
	err = json.Unmarshal([]byte(response), &resources)
	if err != nil {
		fmt.Println(err)
	}

	// Return total count of users
	return len(resources.Resources)
}
