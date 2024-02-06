package main

// #region Imports
import (
	"fmt"       // For formatting strings
	"log"       // For logging information
	"math"      // For doing math functions
	"math/rand" // For getting random numbers (not cryptographic)
	"os"        // For operating with the operating system for files
	"strings"   // For doing string operations
	"time"      // For usings times and dates

	"github.com/bwmarrin/discordgo" // For the underlying discord bot
)

// #endregion Imports

// #region Server Commands

// Handling the messages from the server
func handle_server_messages(dg *discordgo.Session, message *discordgo.MessageCreate) {
	// Check if it is silenced and should be unsilenced
	if silenced == true && time.Now().After(silenced_time) { // Should really just be done with a timer or something else
		silenced = false
		dg.UpdateGameStatus(0, fmt.Sprintf("Birthday in %d days!", closest_birthday()))
	}
	// Listen to the user inputs
	if string([]rune(message.Content)[0]) == "!" { // If it is a command always listen, no matter if the bot is silenced or they've already spoken
		//Split for the keyword
		split_command := strings.Split(message.Content, " ")
		sent_command := split_command[0]
		// Switch based on the command
		switch sent_command {
		case "!help":
			dg.ChannelMessageSend(message.ChannelID,
				"```!help\t\t  To show more commands.\n!silence\t   To silence the bot for an hour.\n!talk\t\t  To un-silence the bot.\n"+
					"!check\t\t Check the time until un-silenced.\n!response\t  To add a response to another user.\n"+
					"!add\t\t   To register yourself as a user.\n"+
					"!remove\t\tTo remove yourself as a user.\n!update\t\tTo update your user information.```")
			return
		case "!silence":
			silenced = true
			silenced_time = time.Now().Add(time.Hour)
			dg.UpdateGameStatus(0, fmt.Sprintf("I have been silenced"))
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf(":sob: Thanks %s, now I will be silenced until %s :sob:",
				discord_id_format(message.Author.ID), silenced_time.Format("2006-01-02 03:04:05 MST")))
			return
		case "!talk":
			silenced = false
			silenced_time = time.Now()
			dg.UpdateGameStatus(0, fmt.Sprintf("Birthday in %d days!", closest_birthday()))
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Thank you %s! I can speak again!", discord_id_format(message.Author.ID)))
			return
		case "!check":
			if silenced {
				between := silenced_time.Sub(time.Now())
				minutes_, seconds_ := math.Modf(between.Minutes())
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf(":pensive: %s, sadly I can't speak until %s! That's %d:%d away! :pensive:",
					discord_id_format(message.Author.ID), silenced_time.Format("2006-01-02 03:04:05 MST"), int(minutes_), int(seconds_*60)))
			} else {
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%s, I can already speak you silly goose!", discord_id_format(message.Author.ID)))
			}
			return
		case "!quit":
			save_to_json(true, true)
			log.Print("Saved and Shut Down as Planned")
			os.Exit(0)
		case "!response":
			hand_dm_commands("response", dg, message)
			return
		case "!add":
			hand_dm_commands("add", dg, message)
			return
		case "!remove":
			hand_dm_commands("remove", dg, message)
			return
		case "!update":
			hand_dm_commands("edit", dg, message)
			return
		default: // For an unknown command
			return
		}
	} else if has_spoken(message.Author.ID) == true { // Don't let them speak again if they've spoken
		return
	} else if known_user(message.Author.ID) && silenced == false && has_spoken(message.Author.ID) == false { // If the user is known
		spoken = append(spoken, message.Author.ID)
	} else if silenced == false && has_spoken(message.Author.ID) == false { // If the user isn't known
		dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("I don't know you %s :rage:", discord_id_format(message.Author.ID)))
		spoken = append(spoken, message.Author.ID)
		return
	}
	// Send result if it's a known person and it's not a command
	// Set age
	age := 0
	// Get person
	temp_p := get_user(message.Author.ID)
	// Check for birthday
	u_y, u_m, u_d := temp_p.Birthday.Date()
	t_y, t_m, t_d := time.Now().Date()
	if u_m == t_m && u_d == t_d {
		age = t_y - u_y
	}
	// Check if it's the birthday or not
	if age > 0 {
		// Wish them a very happy birthday
		for i := 0; i < age; i++ {
			if i != age-1 {
				message_text := fmt.Sprintf("Happy Birthday %s! Congratulations on having been %d!", temp_p.Name, i+1)
				dg.ChannelMessageSend(message.ChannelID, message_text)
				time.Sleep(time.Duration(pause) * time.Millisecond)
			} else {
				message_text := fmt.Sprintf("Happy Birthday %s!!! Congratulations on turning %d!", temp_p.Name, i+1)
				dg.ChannelMessageSend(message.ChannelID, message_text)
				time.Sleep(time.Duration(pause) * time.Millisecond)
			}
		}
	} else { // Tell them one of their random quips
		// Get a random number
		if len(temp_p.Responses) > 0 {
			select_num := rand.Intn(len(temp_p.Responses))
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%s, %s", discord_id_format(message.Author.ID), temp_p.Responses[select_num]))
		} else {
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%s, I'm so sorry. You have no responses :sob:", discord_id_format(message.Author.ID)))
		}
	}
}

// #endregion Server Commands
