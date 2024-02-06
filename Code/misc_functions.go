package main

// #region Imports
import (
	"fmt"  // For formatting strings
	"time" // For usings times and dates
)

// #endregion Imports

// #region Misc Functions

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

// Get person index
func get_user_index(id string) int {
	// Check each known person
	for i := 0; i < len(People); i++ {
		if People[i].Id == id {
			return i
		}
	}
	return 0 // Possibly not the best thing, but I don't expect it to happen
}

// Get user
func get_user(id string) Person {
	return People[get_user_index(id)]
}

// Edit person
func edit_person(user_id string, new_person Person) {
	index := get_user_index(user_id)
	changed := false
	if People[index].Name != new_person.Name {
		People[index].Name = new_person.Name
		changed = true
	}
	if People[index].Birthday != new_person.Birthday {
		People[index].Birthday = new_person.Birthday
		changed = true
	}
	if changed {
		save_to_json(true, false)
	}
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

// #endregion Misc Functions
