package main

type Catalog struct {
	Repositories []string `json:"repositories"`
}

type TagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}
