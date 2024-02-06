package main

// #region Imports
import (
	"time" // For usings times and dates
)

// #endregion Imports

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

// / Structure to hold DM information
type DM_Response struct {
	Option     string
	Stage      int
	Guild      string
	TargetID   string
	Response   string
	LastUpdate time.Time
	Add_Cat    map[int]string
	Person     Person
}

// #endregion Structs
