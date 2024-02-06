package main

// #region Imports
import (
	"log" // For logging information
	"os"  // For operating with the operating system for files
)

// #endregion Imports

// #region Main Code

// Initialize Files
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
