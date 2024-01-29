package main

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

// Main function
func main() {
	// Attempting to connect the token
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Print("Can't connect to the discord token\n", err)
		os.Exit(4)
	}
	// Attempting to get the userID of the bot
	user, err := dg.User("@me")
	if err != nil {
		log.Print("Can't get the discord bots user ID\n", err)
		os.Exit(4)
	}
	BotID = user.ID
	// Creating the handler to scan the messages
	dg.AddHandler(Happy_Birthday)
	// Attempting to open the connection to the bot
	err = dg.Open()
	if err != nil {
		log.Print("Can't establish a connection with the bot\n", err)
		os.Exit(4)
	}
	// Defer closing the connection until the main loop is done
	defer dg.Close()
	// Keep the code running indefinitely
	for {
	}
}

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

var People []Person                        // Variable to hold people
var Discord_Credentials Discord_Credential // Variable to hold discord token

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

// Message function
func Happy_Birthday(dg *discordgo.Session, message *discordgo.MessageCreate) {
	/* Debug: Message Info
	fmt.Print("\n Author: ",message.Author.ID)
	fmt.Print("\n Channel: ",message.ChannelID)
	fmt.Print("\n Message: ",message.Content)
	fmt.Print("\n Server: ",message.GuildID) */
	// Variable to determine result
	send_message := false
	age := 0
	message_ := message.Content // Have to save this separate for the rune conversion
	// Listen to the command
	if message.Author.ID == BotID { // Don't listen to yourself silly
		send_message = false
	} else if message.Author.Bot { // Don't listen to other bots silly head
		send_message = false
	} else if string([]rune(message_)[0]) == "!" { // If it is a command
		//Split for the keyword
		split_command := strings.Split(message.Content, " ")
		sent_command := split_command[0]
		// Switch based on the command
		switch sent_command {
		case "!help":
			dg.ChannelMessageSend(message.ChannelID,
				"```!help\t\tTo show more commands.\n!silence\tTo silence the bot for an hour.\n!talk\t\tTo un-silence the bot.\n!check\tCheck the time until un-silenced.```")
		case "!silence":
			silenced = true
			silenced_time = time.Now().Add(time.Hour)
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf(":sob: I will be silenced until %s :sob:", silenced_time.Format("2006-01-02 03:04:05 MST")))
		case "!talk":
			silenced = false
			silenced_time = time.Now()
			dg.ChannelMessageSend(message.ChannelID, "I can speak again!")
		case "!check":
			if silenced {
				between := silenced_time.Sub(time.Now())
				minutes_, seconds_ := math.Modf(between.Minutes())
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf(":pensive: I can't speak until %s! That's %d:%d away! :pensive:",
					silenced_time.Format("2006-01-02 03:04:05 MST"), int(minutes_), int(seconds_*60)))
			} else {
				dg.ChannelMessageSend(message.ChannelID, "I can speak now!")
			}
		}
	} else if known_user(message.Author.ID) && silenced == false && has_spoken(message.Author.ID) == false { // Check if the user is known
		send_message = true
		spoken = append(spoken, message.Author.ID)
	} else if silenced == false && has_spoken(message.Author.ID) == false { // If nothing is correct
		dg.ChannelMessageSend(message.ChannelID, "I don't know you :rage:")
		spoken = append(spoken, message.Author.ID)
	}
	// If send message is true check the date
	if send_message == true {
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
			dg.ChannelMessageSend(message.ChannelID, temp_p.Responses[select_num])
		}
	} else { // If not correct return doing nothing
		return
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

// Saving a People to Json
func save_people_to_json(people_raw []Person) []byte {
	// Convert the struct into json
	json_person, err := json.Marshal(people_raw)
	// Check that there isn't an error
	if err != nil {
		//panic(err)
		//log.Fatal(err)
		log.Print("Error with saving people to json\n", err)
		os.Exit(2)
	}
	// Returning the json in bytes
	return json_person
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
		log.Print("Error with loading people to json\n", err)
		os.Exit(3)
	}
	// Return the people object
	return *empty_people
}

// Load the Json file
func load_json_file(file_name string) string {
	// Check if file exists
	file, err := os.Open(file_name)
	if err != nil {
		log.Print("Error opening json file\n", err)
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

// Save the Json File
func save_json_file(file_name string, write_data []byte) {
	// Opening the file
	file, err := os.Open(file_name)
	if err != nil {
		log.Print("Error opening json file\n", err)
		os.Exit(5)
	}
	//Defer closing the file until the function is complete
	defer file.Close()
	// Buffer to read from the file
	file_buffer := bufio.NewWriter(file)
	// Defer flushing the writer until the function is complete
	defer file_buffer.Flush()
	// Write the data to the file
	file_buffer.Write(write_data)
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

// Saving a People to Json
func save_discord_to_json(discord_raw Discord_Credential) []byte {
	// Convert the struct into json
	json_discord, err := json.Marshal(discord_raw)
	// Check that there isn't an error
	if err != nil {
		log.Print("Error with saving discord credentials to json\n", err)
		os.Exit(2)
	}
	// Returning the json in bytes
	return json_discord
}

// Loading People from Json
func load_json_to_discord(json_string string) Discord_Credential {
	// Creating a pointer to an empty person
	empty_discord := &Discord_Credential{}
	// Converting the string into bytes
	buffered_string := []byte(json_string)
	// Convert the json into a struct
	err := json.Unmarshal(buffered_string, empty_discord)
	// Panic if there is an error
	if err != nil {
		log.Print("Error with loading discord credentials from json\n", err)
		os.Exit(3)
	}
	// Return the people object
	return *empty_discord
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
