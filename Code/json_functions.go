package main

// #region Imports
import (
	"bufio"         // For writing to files
	"encoding/json" // For marshalling (encoding) and unmarshalling (decoding) json
	"errors"        // For handling errors
	"fmt"           // For formatting strings
	"log"           // For logging information
	"os"            // For operating with the operating system for files
	"strings"       // For doing string operations
	"time"          // For usings times and dates
)

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
