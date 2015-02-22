package main

// http://mervine.net/json2struct

type Job struct {
	Domain   string   `json:"domain"`
	Results  []string `json:"results"`
	Duration int      `json:"duration"` // in ms
	Error    string   `json:"error"`
	Security string   `json:"security"`
}
