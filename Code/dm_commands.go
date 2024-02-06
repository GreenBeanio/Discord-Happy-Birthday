package main

// #region Imports

import (
	"fmt"     // For formatting strings
	"log"     // For logging information
	"strconv" // For converting strings to ints and vice versa
	"time"    // For usings times and dates

	"github.com/bwmarrin/discordgo" // For the underlying discord bot
)

// #endregion Imports

// #region Direct Message Code

// Handling commands that interact require DMs
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
		default_message := fmt.Sprintf("Hello %s! You can type \"Quit\" at any time to stop your request.", discord_id_format(message.Author.ID))
		// These have a bit of repeated code, but it is what is is
		switch option {
		case "response":
			temp := DM_Response{Option: option, Stage: 0, Guild: message.GuildID, LastUpdate: time.Now()}
			DM_Sessions[message.Author.ID] = &temp
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("I sent you a DM %s", discord_id_format(message.Author.ID)))
			dg.ChannelMessageSend(user.ID, fmt.Sprintf("%s\nDid you mean to add a response to a user?\n\"Yes\" or \"No\"", default_message))
		case "add":
			if known_user(message.Author.ID) {
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%s you silly goose! You're already registered!", discord_id_format(message.Author.ID)))
			} else {
				temp := DM_Response{Option: option, Stage: 0, Guild: message.GuildID, LastUpdate: time.Now(), Person: Person{Id: message.Author.ID}}
				DM_Sessions[message.Author.ID] = &temp
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("I sent you a DM %s", discord_id_format(message.Author.ID)))
				dg.ChannelMessageSend(user.ID, fmt.Sprintf("%s\nDo you wish to add yourself as a user?\n\"Yes\" or \"No\"", default_message))
			}
		case "remove":
			if known_user(message.Author.ID) {
				temp := DM_Response{Option: option, Stage: 0, Guild: message.GuildID, LastUpdate: time.Now()}
				DM_Sessions[message.Author.ID] = &temp
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("I sent you a DM %s", discord_id_format(message.Author.ID)))
				dg.ChannelMessageSend(user.ID, fmt.Sprintf("%s\nDo you want to remove yourself as a user?\n\"Yes\" or \"No\"", default_message))
			} else {
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%s you silly goose! You're not even registered!", discord_id_format(message.Author.ID)))
			}
		case "edit":
			if known_user(message.Author.ID) {
				temp := DM_Response{Option: option, Stage: 0, Guild: message.GuildID, LastUpdate: time.Now(), Person: get_user(message.Author.ID)}
				DM_Sessions[message.Author.ID] = &temp
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("I sent you a DM %s", discord_id_format(message.Author.ID)))
				dg.ChannelMessageSend(user.ID, "Choose which action you'd like to do..\n0: Quit\n1: Edit Name\n2: Edit Birthday\n3: Save Changes")
			} else {
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%s you silly goose! You're not even registered!", discord_id_format(message.Author.ID)))
			}
		}
	} else { // If the user does have an open dm remind them
		dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("You already have an open DM %s!", discord_id_format(message.Author.ID)))
	}
}

// Handling the messages from DMs
func handle_direct_messages(dg *discordgo.Session, message *discordgo.MessageCreate) {
	// If the user has an active DM
	if !is_user_in_dm(message.Author.ID) {
		// Update the DM time
		DM_Sessions[message.Author.ID].LastUpdate = time.Now()
		// Select which command is being used
		switch DM_Sessions[message.Author.ID].Option {
		case "response":
			hand_dm_response(DM_Sessions[message.Author.ID].Stage, dg, message)
		case "add":
			hand_dm_add(DM_Sessions[message.Author.ID].Stage, dg, message)
		case "remove":
			hand_dm_remove(DM_Sessions[message.Author.ID].Stage, dg, message)
		case "edit":
			hand_dm_edit(DM_Sessions[message.Author.ID].Stage, dg, message)
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

// Function to delete a person (and their dm)
func hand_person_delete(dg *discordgo.Session, message *discordgo.MessageCreate) {
	// Delete the DM from the map
	delete(DM_Sessions, message.Author.ID)
	// Delete the struct from the array, actually a slice, but whatever
	for i := 0; i < len(People); i++ {
		if People[i].Id == message.Author.ID {
			People = append(People[:i], People[i+1:]...)
			break
		}
	}
	dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("No problem %s. I deleted your account! You can make one again at any time if you desire!", discord_id_format(message.Author.ID)))
}

// Handling the response command DM
func hand_dm_response(stage int, dg *discordgo.Session, message *discordgo.MessageCreate) {
	// If the user quits
	if message.Content == "Quit" {
		hand_dm_quit(true, dg, message)
		return
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
		// Ask the user which user is correct
		if valid_user != nil {
			// First make sure it isn't themselves, we don't want people making their own responses that's not fun!
			if message.Author.ID == valid_user.User.ID {
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("%s, You silly goose! That's you! You can't add a response to yourself! That's no fun! Send me another username to try again!", discord_id_format(message.Author.ID)))
			} else if known_user(valid_user.User.ID) { // Check if the user is known
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

// Handling the add_me command DM
func hand_dm_add(stage int, dg *discordgo.Session, message *discordgo.MessageCreate) {
	// If the user quits
	if message.Content == "Quit" {
		hand_dm_quit(true, dg, message)
		return
	}
	switch stage {
	case 0: // Confirm that they want to add themselves
		if message.Content == "Yes" {
			// Get a list of the unfilled aspects
			add_len := 0
			add_string := "\n0: Quit"
			add_result := make(map[int]string)
			add_result[add_len] = "Quit"
			if DM_Sessions[message.Author.ID].Person.Name == "" {
				add_len = add_len + 1
				add_string = add_string + fmt.Sprintf("\n%d: Name", add_len)
				add_result[add_len] = "Name"
			}
			if DM_Sessions[message.Author.ID].Person.Birthday.IsZero() {
				add_len = add_len + 1
				add_string = add_string + fmt.Sprintf("\n%d: Birthday", add_len)
				add_result[add_len] = "Birthday"
			}
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Great! What aspect would you like to add first?%s", add_string))
			DM_Sessions[message.Author.ID].Add_Cat = add_result
			DM_Sessions[message.Author.ID].Stage = 1
			DM_Sessions[message.Author.ID].Response = add_string
		} else if message.Content == "No" {
			hand_dm_quit(false, dg, message)
		} else {
			dg.ChannelMessageSend(message.ChannelID, "That input was incorrect!\n\"Yes\" or \"No\"")
		}
	case 1: // Get which aspect to add
		// Get the first letter only of the response
		var valid_response bool
		response := string([]rune(message.Content)[0])
		response_int, err := strconv.Atoi(response)
		if err != nil {
			valid_response = false
		} else {
			// Check if it's an allowed int
			if response_int <= len(DM_Sessions[message.Author.ID].Add_Cat) {
				// Check if it's 0, as it'll always be quitting
				if response_int == 0 {
					hand_dm_quit(false, dg, message)
					return
				} else {
					valid_response = true
				}
			} else {
				valid_response = false
			}
		}
		// If it's a valid response
		if valid_response {
			DM_Sessions[message.Author.ID].Response = DM_Sessions[message.Author.ID].Add_Cat[response_int]
			DM_Sessions[message.Author.ID].Add_Cat = make(map[int]string)
			DM_Sessions[message.Author.ID].Stage = 2
			switch DM_Sessions[message.Author.ID].Response {
			case "Name":
				dg.ChannelMessageSend(message.ChannelID, "Great! Please give me your name.")
			case "Birthday":
				dg.ChannelMessageSend(message.ChannelID, "Great! Please give me your birthday in the format Year-Month-Day (such as 2000-12-30).")
			}
		} else {
			DM_Sessions[message.Author.ID].Stage = 0
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("That input was incorrect! Select one of the following to add.%s",
				DM_Sessions[message.Author.ID].Response))
		}
	case 2: // Get the users value for their chosen category
		switch DM_Sessions[message.Author.ID].Response {
		case "Name":
			DM_Sessions[message.Author.ID].Person.Name = message.Content
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Confirm you'd like to be called\n%s.\n\"Yes\" or \"No\"",
				DM_Sessions[message.Author.ID].Person.Name))
			DM_Sessions[message.Author.ID].Stage = 3
		case "Birthday":
			//Split the text up using the date format stated
			converted_date, err := time.Parse("2006-01-02", message.Content)
			if err != nil {
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Something wen't wrong with your date %s. Let's try that again!\nPlease give me your birthday in the format Year-Month-Day (such as 2000-12-30).", discord_id_format(message.Author.ID)))
			} else {
				DM_Sessions[message.Author.ID].Person.Birthday = converted_date
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Confirm that your birthday is\n%s.\n\"Yes\" or \"No\"",
					DM_Sessions[message.Author.ID].Person.Birthday))
				DM_Sessions[message.Author.ID].Stage = 3
			}
		}
	case 3: // Get the users to confirm their input
		if message.Content == "Yes" {
			// If we have all the inputs filled in
			if DM_Sessions[message.Author.ID].Person.Name != "" && !DM_Sessions[message.Author.ID].Person.Birthday.IsZero() {
				People = append(People, DM_Sessions[message.Author.ID].Person)
				save_to_json(true, false)
				dg.ChannelMessageSend(message.ChannelID, "Great! I've added you!")
				hand_dm_quit(true, dg, message)
			} else { // If we need to get another response
				switch DM_Sessions[message.Author.ID].Response {
				case "Name":
					dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Great! Your name is %s. Time to tell me your birthday in the format Year-Month-Day (such as 2000-12-30)!", DM_Sessions[message.Author.ID].Person.Name))
					DM_Sessions[message.Author.ID].Response = "Birthday"
					DM_Sessions[message.Author.ID].Stage = 2
				case "Birthday":
					dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Great! Your birthday is %s. Time to tell me your name!",
						DM_Sessions[message.Author.ID].Person.Birthday))
					DM_Sessions[message.Author.ID].Response = "Name"
					DM_Sessions[message.Author.ID].Stage = 2
				}
			}
		} else if message.Content == "No" {
			switch DM_Sessions[message.Author.ID].Response {
			case "Name":
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Oh no! I'm sorry about that %s. Let's try that again!\nPlease give me your name.", discord_id_format(message.Author.ID)))
			case "Birthday":
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Oh no! I'm sorry about that %s. Let's try that again!\nPlease give me your birthday in the format Year-Month-Day (such as 2000-12-30).", discord_id_format(message.Author.ID)))
			}
			DM_Sessions[message.Author.ID].Stage = 2
		} else {
			dg.ChannelMessageSend(message.ChannelID, "That input was incorrect!\n\"Yes\" or \"No\"")
		}
	}
}

// Handling the remove_me command
func hand_dm_remove(stage int, dg *discordgo.Session, message *discordgo.MessageCreate) {
	// If the user quits
	if message.Content == "Quit" {
		hand_dm_quit(true, dg, message)
		return
	}
	switch stage {
	case 0: // Ask if they want to delete their information
		if message.Content == "Yes" {
			DM_Sessions[message.Author.ID].Stage = 1
			dg.ChannelMessageSend(message.ChannelID, "Are you absolutely sure? Your data can't be recovered.\n\"Yes\" or \"No\"")
		} else if message.Content == "No" {
			hand_dm_quit(false, dg, message)
		} else {
			dg.ChannelMessageSend(message.ChannelID, "That input was incorrect!\n\"Yes\" or \"No\"")
		}
	case 1:
		if message.Content == "Yes" {
			hand_person_delete(dg, message)
		} else if message.Content == "No" {
			dg.ChannelMessageSend(message.ChannelID, "Understood! You won't be deleted.\n\"Yes\" or \"No\"")
			hand_dm_quit(false, dg, message)
		} else {
			dg.ChannelMessageSend(message.ChannelID, "That input was incorrect!\n\"Yes\" or \"No\"")
		}
	}
}

// Handling the edit_me command
func hand_dm_edit(stage int, dg *discordgo.Session, message *discordgo.MessageCreate) {
	// If the user quits
	if message.Content == "Quit" {
		hand_dm_quit(true, dg, message)
		return
	}
	switch stage {
	case 0: // Get what they want to edit
		response := string([]rune(message.Content)[0])
		response_int, err := strconv.Atoi(response)
		if err != nil {
			dg.ChannelMessageSend(message.ChannelID, "That wasn't a valid choice! Choose which action you'd like to do.\n0: Quit\n1: Edit Name\n2: Edit Birthday\n3: Save Changes")
		} else {
			switch response_int {
			case 0:
				hand_dm_quit(false, dg, message)
			case 1:
				DM_Sessions[message.Author.ID].Stage = 1
				DM_Sessions[message.Author.ID].Response = "Name"
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Your current name is\n%s\nWhat would you like your new name to be?", DM_Sessions[message.Author.ID].Person.Name))
			case 2:
				DM_Sessions[message.Author.ID].Stage = 1
				DM_Sessions[message.Author.ID].Response = "Birthday"
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Your current birthday is\n%s\nWhat is your real birthday in the format Year-Month-Day (such as 2000-12-30)?", DM_Sessions[message.Author.ID].Person.Birthday))
			case 3:
				DM_Sessions[message.Author.ID].Stage = 1
				DM_Sessions[message.Author.ID].Response = "Save"
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Confirm that you really want to save and overwrite your previous information.\n\"Yes\" or \"No\""))
			default:
				dg.ChannelMessageSend(message.ChannelID, "That wasn't a valid choice! Choose which action you'd like to do.\n0: Quit\n1: Edit Name\n2: Edit Birthday\n3: Save Changes")
			}
		}
	case 1: // Get the new value
		switch DM_Sessions[message.Author.ID].Response {
		case "Name":
			DM_Sessions[message.Author.ID].Person.Name = message.Content
			dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Confirm you'd like to be called\n%s.\n\"Yes\" or \"No\"",
				DM_Sessions[message.Author.ID].Person.Name))
			DM_Sessions[message.Author.ID].Stage = 2
		case "Birthday":
			converted_date, err := time.Parse("2006-01-02", message.Content)
			if err != nil {
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Something wen't wrong with your date %s. Let's try that again!\nPlease give me your birthday in the format Year-Month-Day (such as 2000-12-30).", discord_id_format(message.Author.ID)))
			} else {
				DM_Sessions[message.Author.ID].Person.Birthday = converted_date
				dg.ChannelMessageSend(message.ChannelID, fmt.Sprintf("Confirm that your birthday is\n%s.\n\"Yes\" or \"No\"",
					DM_Sessions[message.Author.ID].Person.Birthday))
				DM_Sessions[message.Author.ID].Stage = 2
			}
		case "Save":
			switch message.Content {
			case "Yes":
				edit_person(message.Author.ID, DM_Sessions[message.Author.ID].Person)
				DM_Sessions[message.Author.ID].Stage = 0
				dg.ChannelMessageSend(message.ChannelID, "Your data has been updated! Choose which action you'd like to do.\n0: Quit\n1: Edit Name\n2: Edit Birthday\n3: Save Changes")
			case "No":
				DM_Sessions[message.Author.ID].Stage = 0
				dg.ChannelMessageSend(message.ChannelID, "The updates weren't saved! Choose which action you'd like to do.\n0: Quit\n1: Edit Name\n2: Edit Birthday\n3: Save Changes")
			default:
				dg.ChannelMessageSend(message.ChannelID, "That input was incorrect!\n\"Yes\" or \"No\"")
			}
		}
	case 2:
		switch message.Content {
		case "Yes":
			DM_Sessions[message.Author.ID].Stage = 0
			dg.ChannelMessageSend(message.ChannelID, "Recorded your change, but it hasn't been saved yet! Choose which action you'd like to do.\n0: Quit\n1: Edit Name\n2: Edit Birthday\n3: Save Changes")
		case "No":
			DM_Sessions[message.Author.ID].Stage = 0
			dg.ChannelMessageSend(message.ChannelID, "Disregarded your data! Choose which action you'd like to do.\n0: Quit\n1: Edit Name\n2: Edit Birthday\n3: Save Changes")
		default:
			dg.ChannelMessageSend(message.ChannelID, "That input was incorrect!\n\"Yes\" or \"No\"")
		}
	}
}

// #endregion Direct Message Code
