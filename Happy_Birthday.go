package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Variables for storing information
var BotID string
var dg *discordgo.Session
var token = "Your Discord Bot Token" // Your discord bot token

// Set Varaibles for delay
var name = "Name"         // Name of the birthday holder
var age = 99              // Age of the birthday
var birthday = "01-31"    //Their birthday MM-DD
var pause = 500           // Milliseconds
var birthday_ID = ""      //Put in a ID to use ID instead of command
var command = "!birthday" //Change the command if you want
// Main function
func main() {
	// Attempting to connect the token
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal(err)
		return
	}
	// Attempting to get the userID of the bot
	user, err := dg.User("@me")
	if err != nil {
		log.Fatal(err)
		return
	}
	BotID = user.ID
	// Creaitng the handler to scan the messages
	dg.AddHandler(Happy_Birthday)
	// Attempting to open the conenction to the bot
	err = dg.Open()
	if err != nil {
		log.Fatal(err)
		return
	}
	for {
		//Keeping the program running
		defer dg.Close()
	}
}

// Message funciton
func Happy_Birthday(dg *discordgo.Session, message *discordgo.MessageCreate) {
	/* Debug: Message Info
	fmt.Print("\n Author: ",message.Author.ID)
	fmt.Print("\n Channel: ",message.ChannelID)
	fmt.Print("\n Message: ",message.Content)
	fmt.Print("\n Server: ",message.GuildID) */
	// Variable to determine result
	send_message := false
	// Listen to the command
	if message.Author.ID == BotID { //Don't listen to yourself silly
		send_message = false
	} else if birthday_ID == "" { // If not using the user ID
		if strings.ToLower(message.Content) == command { //If the command is correct
			send_message = true
		}
	} else if message.Author.ID == birthday_ID { // If using the user ID and it matches
		send_message = true
	} else { // If nothing is correct
		send_message = false
	}
	// If send message is true check the date
	if send_message == true {
		// Getting and formatting time
		time_now := time.Now()
		time_format := fmt.Sprintf(time_now.Format("01-02"))
		// If it is the birthday then start the birthday wishes
		if time_format == birthday {
			for i := 0; i < age; i++ {
				message_text := fmt.Sprintf("%d: Happy Birthday %s!!!", i+1, name)
				dg.ChannelMessageSend(message.ChannelID, message_text)
				time.Sleep(time.Duration(pause) * time.Millisecond)
			}
		}
	} else { // If not correct return doing nothing
		return
	}
}
