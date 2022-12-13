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

	token := os.Getenv("DISCORD_TOKEN")
	fmt.Println("token" + token)

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

		if awsBatonUserSearch(query) {
			// If user found notify discord
			s.ChannelMessageSend(m.ChannelID, query+" was found!")
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

// Format of user data within the baton json response
type AWSResources struct {
	Resources []struct {
		Resource struct {
			DisplayName string `json:"displayName"`
		} `json:"resource"`
	} `json:"resources"`
}

// Returns a bool if "user" was found or not
func awsBatonUserSearch(user string) bool {

	// run terminal baton resources and capture json output
	out, err := exec.Command("baton", "resources", "-o", "json").Output()

	// catch error if process above produces one
	if err != nil {
		log.Fatal(err)
	}

	// convert baton response into a string
	response := string(out)

	// process json
	var resources AWSResources
	err = json.Unmarshal([]byte(response), &resources)
	if err != nil {
		fmt.Println(err)
	}

	// check if user exists in json output
	for _, r := range resources.Resources {
		if r.Resource.DisplayName == user {
			return true
		}
	}

	// If user not found, return false
	return false
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

	// process json
	var resources AWSResources
	err = json.Unmarshal([]byte(response), &resources)
	if err != nil {
		fmt.Println(err)
	}

	return len(resources.Resources)
}
