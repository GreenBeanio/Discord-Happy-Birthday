package main

// #region Imports
import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// #endregion Imports

// #region Variables

// Track the state of the bot
var silenced bool = false
var silenced_time time.Time

// Track the users who
var spoken []string

// Track current day
var current_day = time.Now().Day()

// Variables for storing discord information
var BotID string
var dg *discordgo.Session
var token string

// Set Variables for delay
var pause = 500 // Milliseconds to pause between birthday messages

// Variables to hold loaded information
var People []Person                        // Variable to hold people
var Discord_Credentials Discord_Credential // Variable to hold discord token

// Variable for tracking direct message sessions (Trying a map instead of a struct)
var DM_Sessions = map[string]*DM_Response{}

// #endregion Variables

// #region Structs

// Struct to hold the information about discord
type Discord_Credential struct {
	Token string `json:"token"`
}

// Struct to hold information about people
type Person struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Birthday  time.Time `json:"birthday"`
	Responses []string  `json:"responses"`
}

type DM_Response struct {
	Option   string
	Stage    int
	Guild    string
	TargetID string
	Response string
}

// #endregion Structs

// #region Main Code

// Initialize my variables (will load from a file in the future)
func init() {
	// Check if the files exist
	continue_or_not := both_exist("people.json", "discord.json")
	// Only continue if both files exist
	if continue_or_not {
		// Load and convert the discord file
		discord_json := load_json_file("discord.json")
		Discord_Credentials = load_json_to_discord(discord_json)
		token = Discord_Credentials.Token
		// Load and convert the people file
		people_json := load_json_file("people.json")
		People = load_json_to_person(people_json)
	} else {
		log.Print("No data files")
		os.Exit(1)
	}
}

// Main function
func main() {
	// Reset the people who have talked
	spoken = nil
	// Making a channel to keep the program running
	discord_channel := make(chan bool, 1)
	// Running the main discord function
	go main_discord(discord_channel)
	// Waiting for the channel
	<-discord_channel
	// Restart the main loop for the new day. Is this proper?
	main()
}

// #endregion Main Code

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

// Handling the messages from the server
func handle_server_messages(dg *discordgo.Session, message *discordgo.MessageCreate) {
	// Check if it is silenced and should be unsilenced
	if silenced == true && time.Now().After(silenced_time) { // Should really just be done with a timer or something else
		silenced = false
		dg.UpdateGameStatus(0, fmt.Sprintf("Birthday in %d days!", closest_birthday()))
	}
	// Listen to the user inputs
	if string([]rune(message.Content)[0]) == "!" { // If it is a command
		//Split for the keyword
		split_command := strings.Split(message.Content, " ")
		sent_command := split_command[0]
		// Switch based on the command
		switch sent_command {
		case "!help":
			dg.ChannelMessageSend(message.ChannelID,
				"```!help\t\tTo show more commands.\n!silence\tTo silence the bot for an hour.\n!talk\t\tTo un-silence the bot.\n!check\tCheck the time until un-silenced.```")
			return
		case "!silence":
			silenced = true
			silenced_time = time.Now().Add(time.Hour)
			dg.UpdateGameStatus(0, fmt.Sprintf("I have been silenced"))
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf(":sob: Thanks %s, now I will be silenced until %s :sob:", discord_id_format(message.Author.ID), silenced_time.Format("2006-01-02 03:04:05 MST")))
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
		case "!add_me":
			hand_dm_commands("add", dg, message)
			return
		case "!remove_me":
			hand_dm_commands("remove", dg, message)
			return
		case "!update_me":
			hand_dm_commands("edit", dg, message)
		default: // For an unknown command
			return
		}
	} else if known_user(message.Author.ID) && silenced == false && has_spoken(message.Author.ID) == false { // Check if the user is known
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
		select_num := rand.Intn(len(temp_p.Responses))
		dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%s, %s", discord_id_format(message.Author.ID), temp_p.Responses[select_num]))
	}
}

// Handling commands that interact require DMS
func hand_dm_commands(option string, dg *discordgo.Session, message *discordgo.MessageCreate) {
	// Get the user
	user, err := dg.UserChannelCreate(message.Author.ID)
	if err != nil {
		log.Print(fmt.Sprintf("Error with getting a user for a direct message\n%v", err))
		dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Sorry, %s there was an error in your request", discord_id_format(message.Author.ID)))
		return
	}
	// If the user doesn't have an open dm add them to the map
	if is_user_in_dm(message.Author.ID) {
		temp := DM_Response{Option: option, Stage: 0, Guild: message.GuildID}
		DM_Sessions[message.Author.ID] = &temp
		fmt.Println(DM_Sessions)
		dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("I sent you a DM %s", discord_id_format(message.Author.ID)))
		if option == "response" {
			dg.ChannelMessageSend(user.ID, fmt.Sprintf("Hello %s! Did you mean to add a response to a user?\n\"Yes\" or \"No\"", discord_id_format(message.Author.ID)))
		} // else if option == "add" {
		// 	dg.ChannelMessageSend(user.ID, fmt.Sprintf("Hello %s! What is your name?", discord_id_format(message.Author.ID)))
		// } else if option == "remove" {
		// 	dg.ChannelMessageSend(user.ID, fmt.Sprintf("Hello %s. Exactly type \"Delete %s\" to delete data about you.", discord_id_format(message.Author.ID), discord_id_format(message.Author.ID)))
		// }
	} else { // If the user does have an open dm remind them
		dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("You already have an open DM %s!", discord_id_format(message.Author.ID)))
	}
}

// Handling the messages from DMs
func handle_direct_messages(dg *discordgo.Session, message *discordgo.MessageCreate) {
	// If the user has an active DM
	if !is_user_in_dm(message.Author.ID) {
		switch DM_Sessions[message.Author.ID].Option {
		case "response":
			hand_dm_response(DM_Sessions[message.Author.ID].Stage, dg, message)
		case "add":
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Hello %s! You would like to add your information!", discord_id_format(message.Author.ID)))
			delete(DM_Sessions, message.Author.ID)
		case "remove":
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Hello %s! You would like to delete your information!", discord_id_format(message.Author.ID)))
			delete(DM_Sessions, message.Author.ID)
		case "edit":
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Hello %s! You would you like to edit your information!", discord_id_format(message.Author.ID)))
			delete(DM_Sessions, message.Author.ID)
		}
	} else {
		dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("You don't have any DMs active %s!", discord_id_format(message.Author.ID)))
	}
}

// Function to handle quitting DMs
func hand_dm_quit(complete bool, dg *discordgo.Session, message *discordgo.MessageCreate) {
	if complete {
		dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("See you next time %s!", discord_id_format(message.Author.ID)))
		delete(DM_Sessions, message.Author.ID)
	} else {
		dg.ChannelMessageSend(message.ChannelID, "No problem. Talk to you next time!")
		delete(DM_Sessions, message.Author.ID)
	}
}

// Handling the response command DM
func hand_dm_response(stage int, dg *discordgo.Session, message *discordgo.MessageCreate) {
	// If the user quits
	if message.Content == "Quit" {
		hand_dm_quit(false, dg, message)
	}
	switch stage {
	case 0: // Ask if they wanted to add a response to a user
		if message.Content == "Yes" {
			dg.ChannelMessageSend(message.ChannelID, "Great! What is the username of the user you'd like to add a response to?")
			DM_Sessions[message.Author.ID].Stage = 1
		} else if message.Content == "No" {
			hand_dm_quit(false, dg, message)
		} else {
			dg.ChannelMessageSend(message.ChannelID, "That input was incorrect!\n\"Yes\" or \"No\"")
		}
	case 1: // Ask for the user they wanted
		// Search the server for matching users
		users, err := dg.GuildMembersSearch(DM_Sessions[message.Author.ID].Guild, message.Content, 1)
		if err != nil {
			print("uh oh")
			return
		}
		// Get a list of valid users (discord shouldn't ever allow multiple of the same username)
		var valid_user *discordgo.Member
		for i := 0; i < len(users); i++ {
			// fmt.Println("------- Value\n", users[i])      // Value
			// fmt.Println("------- *Value\n", *users[i])    // Value
			// fmt.Println("------- &Pointer\n-", &users[i]) // Pointer
			if users[i].User.Username == message.Content {
				valid_user = users[i]
				break
			}
		}
		// Ask the user which user is correct if there are multiple somehow, but discord shouldn't ever allow that
		if valid_user != nil {
			// Check if the user is known
			if known_user(valid_user.User.ID) {
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("I found the user %s! Are they the correct user?\n\"Yes\" or \"No\"", discord_id_format(valid_user.User.ID)))
				DM_Sessions[message.Author.ID].Stage = 2
				DM_Sessions[message.Author.ID].TargetID = valid_user.User.ID
			} else { // If the user is unknown
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("I found the user %s! OH NO! They aren't registered! Would you like to try a different user?\n\"Yes\" or \"No\"", discord_id_format(valid_user.User.ID)))
				DM_Sessions[message.Author.ID].Stage = 5
			}
		} else {
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Sorry %s, I can't find anyone with that username in that server. Send me another username to try again!", discord_id_format(message.Author.ID)))
		}
	case 2: // Have them confirm the user
		if message.Content == "Yes" {
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Great! Tell me the response you would like to add to %s!", discord_id_format(DM_Sessions[message.Author.ID].TargetID)))
			DM_Sessions[message.Author.ID].Stage = 3
		} else if message.Content == "No" {
			dg.ChannelMessageSend(message.ChannelID, "No problem! Give me another username!")
			DM_Sessions[message.Author.ID].Stage = 1
		} else {
			dg.ChannelMessageSend(message.ChannelID, "That input was incorrect!\n\"Yes\" or \"No\"")
		}
	case 3: // Ask them to confirm the response
		dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Type \"Yes\" or \"No\" to confirm this is the response you want for %s\n%s", discord_id_format(DM_Sessions[message.Author.ID].TargetID), message.Content))
		DM_Sessions[message.Author.ID].Stage = 4
		DM_Sessions[message.Author.ID].Response = message.Content
	case 4: // Check if they confirmed or denied the response
		if message.Content == "Yes" {
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Great! I'll add the response to %s", discord_id_format(DM_Sessions[message.Author.ID].TargetID)))
			add_response(message.Author.ID, DM_Sessions[message.Author.ID].Response)
			hand_dm_quit(true, dg, message)
		} else if message.Content == "No" {
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("No problem! Tell me the response you would like to add to %s!", discord_id_format(DM_Sessions[message.Author.ID].TargetID)))
			DM_Sessions[message.Author.ID].Stage = 3
		} else {
			dg.ChannelMessageSend(message.ChannelID, "That input was incorrect!\n\"Yes\" or \"No\"")
		}
	case 5: // Confirming if they'd like to try another person if there desired person isn't registered
		if message.Content == "Yes" {
			dg.ChannelMessageSend(message.ChannelID, "Sorry for the inconvenience! What is the username of the new user you'd like to add a response to?")
			DM_Sessions[message.Author.ID].Stage = 1
		} else if message.Content == "No" {
			hand_dm_quit(false, dg, message)
		} else {
			dg.ChannelMessageSend(message.ChannelID, "That input was incorrect!\n\"Yes\" or \"No\"")
		}
	}
}

// Check if the user is known
func known_user(id string) bool {
	// Check each known person
	for i := 0; i < len(People); i++ {
		if People[i].Id == id {
			return true
		}
	}
	return false
}

// Get user
func get_user(id string) Person {
	// Check each known person
	for i := 0; i < len(People); i++ {
		if People[i].Id == id {
			return People[i]
		}
	}
	return Person{} // Not sure how to handle not having something to return
}

// Has spoken
func has_spoken(id string) bool {
	for i := 0; i < len(spoken); i++ {
		if spoken[i] == id {
			return true
		}
	}
	return false
}

// Track when it becomes a new day
func track_day(new_day chan bool, current_day int) {
	for time.Now().YearDay() == current_day {
		time.Sleep(1 * time.Minute)
	}
	new_day <- true
}

// Find when the closest birthday is (hopefully leap years don't mess it up much)
func closest_birthday() int {
	today_ := time.Now().YearDay()
	closest_birthday := 367
	// Check each known person
	for i := 0; i < len(People); i++ {
		temp_distace := 367
		// Get the days until their birthday
		person_day := People[i].Birthday.YearDay()
		if person_day >= today_ {
			// Create a version of their birthday this year
			temp_bd := time.Date(time.Now().Year(), People[i].Birthday.Month(), People[i].Birthday.Day(), 0, 0, 0, 0, time.UTC)
			temp_bd2 := temp_bd.Sub(time.Now())
			temp_distace = int(temp_bd2.Hours() / 24)
		} else if person_day < today_ {
			// Create a version of their birthday for next year
			temp_bd := time.Date(time.Now().Year()+1, People[i].Birthday.Month(), People[i].Birthday.Day(), 0, 0, 0, 0, time.UTC)
			temp_bd2 := temp_bd.Sub(time.Now())
			temp_distace = int(temp_bd2.Hours() / 24)
		}
		if temp_distace < closest_birthday {
			closest_birthday = temp_distace
		}
	}
	return closest_birthday
}

// Check if the user is part of a discord session
func is_user_in_dm(user string) bool {
	// Check if the user doesn't already have a dm open
	if _, ok := DM_Sessions[user]; !ok {
		return true
	} else {
		return false
	}
}

// Format the user id for discord messages
func discord_id_format(raw string) string {
	return fmt.Sprintf("<@%s>", raw)
}

// Add response to user
func add_response(sender_id string, new_response string) {
	// Shouldn't need to worry about the person being found, because that should've been determined earlier
	for i := 0; i < len(People); i++ {
		if People[i].Id == DM_Sessions[sender_id].TargetID {
			People[i].Responses = append(People[i].Responses, DM_Sessions[sender_id].Response)
			break
		}
	}
	save_to_json(true, false)
}

// #endregion Discord Code

// #region JSON Code

// Save the Json File
func save_json_file(file_name string, write_data []byte) {
	// Opening the file
	//file, err := os.Open(file_name)
	file, err := os.Create(file_name) // Using create because if I use Open it says access denied, creating will truncate/replace the old file
	if err != nil {
		log.Print(fmt.Sprintf("Error opening json file\n%v", err))
		os.Exit(5)
	}
	// Buffer to read from the file
	file_buffer := bufio.NewWriter(file)
	// Write the data to the file
	_, err2 := file_buffer.Write(write_data)
	if err2 != nil {
		log.Print(fmt.Sprintf("Error writing to json file\n%v", err2))
		os.Exit(5)
	}
	//Defer flushing and closing the file until the rest of the function is complete
	defer func() {
		// Flushing the buffer to finish writing the file
		err3 := file_buffer.Flush()
		if err3 != nil {
			log.Print(fmt.Sprintf("Error flushing the buffer to write the json file\n%v", err3))
			os.Exit(5)
		}
		// Close the file after the flush
		file.Close()
	}()
}

// Load the Json file
func load_json_file(file_name string) string {
	// Check if file exists
	file, err := os.Open(file_name)
	if err != nil {
		log.Print(fmt.Sprintf("Error opening json file\n%v", err))
		os.Exit(5)
	}
	//Defer closing the file until the function is complete
	defer file.Close()
	// String builder to concat the read strings together
	var read_string strings.Builder
	// Buffer to read from the file
	file_buffer := bufio.NewScanner(file)
	// Reading each line from the file
	for file_buffer.Scan() {
		read_string.WriteString(file_buffer.Text())
	}
	// Return the string
	return read_string.String()
}

// Saving a People to Json
func save_people_to_json(people_raw []Person) []byte {
	// Convert the struct into json
	json_person, err := json.MarshalIndent(people_raw, "", "    ")
	// Check that there isn't an error
	if err != nil {
		//panic(err)
		//log.Fatal(err)
		log.Print(fmt.Sprintf("Error with saving people to json\n%v", err))
		os.Exit(2)
	}
	// Returning the json in bytes
	return json_person
}

// Saving Credentials to Json
func save_discord_to_json(discord_raw Discord_Credential) []byte {
	// Convert the struct into json
	json_discord, err := json.MarshalIndent(discord_raw, "", "    ")
	// Check that there isn't an error
	if err != nil {
		log.Print(fmt.Sprintf("Error with saving discord credentials to json\n%v", err))
		os.Exit(2)
	}
	// Returning the json in bytes
	return json_discord
}

// Loading People from Json
func load_json_to_person(json_string string) []Person {
	// Creating a pointer to an empty person
	empty_people := &[]Person{}
	// Converting the string into bytes
	buffered_string := []byte(json_string)
	// Convert the json into a struct
	err := json.Unmarshal(buffered_string, empty_people)
	// Panic if there is an error
	if err != nil {
		log.Print(fmt.Sprintf("Error with loading people to json\n%v", err))
		os.Exit(3)
	}
	// Return the people object
	return *empty_people
}

// Loading Credentials from Json
func load_json_to_discord(json_string string) Discord_Credential {
	// Creating a pointer to an empty person
	empty_discord := &Discord_Credential{}
	// Converting the string into bytes
	buffered_string := []byte(json_string)
	// Convert the json into a struct
	err := json.Unmarshal(buffered_string, empty_discord)
	// Panic if there is an error
	if err != nil {
		log.Print(fmt.Sprintf("Error with loading discord credentials from json\n%v", err))
		os.Exit(3)
	}
	// Return the people object
	return *empty_discord
}

// Check if a file exists
func file_exists(file_name string) bool {
	// Discard the stat content and only care about the error
	_, err := os.Stat(file_name)
	// If the file doesn't exists
	if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}

// Check if both files exist
func both_exist(people_file string, credentials_file string) bool {
	// Variables to store results
	people_exist := true
	credentials_exist := true
	// Test if the people file exists
	if !file_exists(people_file) {
		// Create the file
		create_json_file(people_file)
		// Set boolean and send message
		people_exist = false
		log.Print("Fill in the people file")
		// Create an example file
		People = []Person{{Id: "Discord Id Number", Name: "Name they go by", Birthday: time.Date(2024, time.January, 29, 0, 0, 0, 0, time.UTC),
			Responses: []string{"Example Response", "Another Example Response", "As Many As You Want"}},
			{Id: "Discord Id Number", Name: "Name they go by", Birthday: time.Date(2024, time.January, 29, 0, 0, 0, 0, time.UTC),
				Responses: []string{"Example Response", "Another Example Response", "As Many As You Want"}}}
		save_to_json(true, false)
	}
	// Test if the credentials file exists
	if !file_exists(credentials_file) {
		// Create the file
		create_json_file(credentials_file)
		// Set boolean and send message
		credentials_exist = false
		log.Print("Fill in the discord file")
		// Create an example file
		Discord_Credentials = Discord_Credential{Token: "Your Discord Bot Token"}
		save_to_json(false, true)
	}
	// Result
	if (people_exist && credentials_exist) == true {
		return true
	} else {
		return false
	}
}

// Save to JSON
func save_to_json(people_ bool, discord_ bool) {
	if people_ {
		// Saving people to json file
		byte_people := save_people_to_json(People)
		save_json_file("people.json", byte_people)
	}
	if discord_ {
		// Saving discord to json file
		byte_discord := save_discord_to_json(Discord_Credentials)
		save_json_file("discord.json", byte_discord)
	}
}

// Create a new file
func create_json_file(file_name string) {
	file, err := os.Create(file_name)
	if err != nil {
		log.Print(fmt.Sprintf("Error when creating a new json file\n%v", err))
		os.Exit(2)
	}
	defer file.Close()
}

// #endregion JSON Code
