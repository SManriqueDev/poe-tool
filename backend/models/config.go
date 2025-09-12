package models

type Config struct {
	PoeSessid         string `json:"poesessid"`
	AccountName       string `json:"accountName"`
	league            string `json:"league"`
	automationEnabled bool   `json:"automationEnabled"`
	delay             int    `json:"delay"`
}
