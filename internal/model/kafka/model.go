package kafka

import (
	"time"
)

type Item struct {
	Title       string     `json:"title"`
	PubDate     *time.Time `json:"pubDate"`
	Description string     `json:"description"`
	FullText    string     `json:"fullText"`
	Name        string     `json:"name"`
	Link        string     `json:"link"`
	MD5         string     `json:"md5"`
	Enclosure   string     `json:"enclosure"`
	Category    string     `json:"category"`
	Changed     bool       `json:"changed"`
}
