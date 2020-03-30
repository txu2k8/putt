package config

import "fmt"

// define const values
const (
	Level int = 3
)

// LoadConfig from json
func LoadConfig() int {
	fmt.Println("Load config ...")
	return 1
}
