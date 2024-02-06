package main

// #region Imports
import (
	"time" // For usings times and dates

	"github.com/bwmarrin/discordgo" // For the underlying discord bot
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
