package main

// #region Imports
import (
	"fmt"  // For formatting strings
	"log"  // For logging information
	"os"   // For operating with the operating system for files
	"time" // For usings times and dates

	"github.com/bwmarrin/discordgo" // For the underlying discord bot
)

// #endregion Imports

// #region Discord Code

func main_discord(done chan bool) {
	// Get the current day
	current_day := time.Now().YearDay()
	// Making a channel to keep track of the switching day
	new_day := make(chan bool, 1)
	// Create the day tracker
	go track_day(new_day, current_day)
	// Getting the closest birthday
	nearest_birthday := closest_birthday()
	// Attempting to connect the token
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Print(fmt.Sprintf("Can't connect to the discord token\n%v", err))
		os.Exit(4)
	}
	// Attempting to get the userID of the bot
	user, err := dg.User("@me")
	if err != nil {
		log.Print(fmt.Sprintf("Can't get the discord bots user ID\n%v", err))
		os.Exit(4)
	}
	BotID = user.ID
	// Creating the handler to scan the messages
	dg.AddHandler(Happy_Birthday)
	// Add intents
	dg.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentsGuildMembers
	// Attempting to open the connection to the bot
	err = dg.Open()
	if err != nil {
		log.Print(fmt.Sprintf("Can't establish a connection with the bot\n%v", err))
		os.Exit(4)
	}
	// Add a custom name
	dg.UserUpdate("The Birthday Friend", "") // Seems to only work randomly?
	// Add the custom activity
	dg.UpdateGameStatus(0, fmt.Sprintf("Birthday in %d days!", nearest_birthday))
	// Defer closing the connection until the main loop is done
	defer dg.Close()
	// Using a channel to keep the discord bot running, and restart every day
	switch {
	case <-new_day:
		done <- true
	}
}

// Discord message function
func Happy_Birthday(dg *discordgo.Session, message *discordgo.MessageCreate) {
	/* Debug: Message Info
	fmt.Println("\n Author: ",message.Author.ID)
	fmt.Println("\n Channel: ",message.ChannelID)
	fmt.Println("\n Message: ",message.Content)
	fmt.Println("\n Server: ",message.GuildID) */
	// Don't do anything if it's the bot itself or another bot
	if message.Author.ID == BotID || message.Author.Bot { // Don't listen to yourself or other bots silly
		return
	}
	if message.GuildID != "" {
		handle_server_messages(dg, message)
	} else {
		handle_direct_messages(dg, message)
	}
}

// #endregion Discord Code
