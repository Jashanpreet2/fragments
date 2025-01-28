package main

import (
	"encoding/csv"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// func generateCsvFile() {
// 	user1 := "user1@email.com"
// 	user1password := "password1"
// 	user1hash, err := bcrypt.GenerateFromPassword([]byte(user1password), bcrypt.DefaultCost)
// 	if err != nil {
// 		fmt.Printf("Failed to hash user1's password!")
// 	}

// 	user2 := "user2@email.com"
// 	user2password := "password2"
// 	user2hash, err := bcrypt.GenerateFromPassword([]byte(user2password), bcrypt.DefaultCost)
// 	if err != nil {
// 		fmt.Printf("Failed to hash user2's password!")
// 	}

// 	data := [][]string{
// 		{"username", "password"},
// 		{user1, string(user1hash)},
// 		{user2, string(user2hash)},
// 	}

// 	file, err := os.Create("testProfiless.csv")
// 	if err != nil {
// 		log.Fatal("Failed to create test profiles csv")
// 	}
// 	defer file.Close()

// 	w := csv.NewWriter(file)
// 	defer w.Flush()

// 	err = w.WriteAll(data)
// 	if err != nil {
// 		log.Fatal("Failed to save test profiles")
// 	}
// }

func AuthenticateTestProfile(filepath string, username string, password string) bool {
	file, err := os.Open("testProfiles.csv")
	if err != nil {
		log.Fatal("Failed to open test profiles csv")
	}

	r := csv.NewReader(file)
	data, err := r.ReadAll()
	if err != nil {
		log.Fatal("Failed to read data from test profiles csv")
	}

	// If there are less than 2 records, it means we have no profiles in the csv file
	if len(data) < 2 {
		return false
	}

	for i := 1; i < len(data); i++ {
		if data[i][0] == username {
			err = bcrypt.CompareHashAndPassword([]byte(data[i][1]), []byte(password))
			if err != nil {
				return false
			} else {
				return true
			}
		}
	}

	// No user records found
	return false
}
