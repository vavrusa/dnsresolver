package main

// http://mervine.net/json2struct

type Job struct {
	Id int `json:"id"`
	Domain   string   `json:"domain"`
	Results  []string `json:"results"`
	Duration int      `json:"duration"` // in ms
	Error    string   `json:"error"`
}
