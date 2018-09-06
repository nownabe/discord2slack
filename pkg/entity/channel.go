package entity

import "google.golang.org/appengine/datastore"

type Channel struct {
	ID            string         `datastore:"-" goon:"id"`
	GuildKey      *datastore.Key `datastore:"-" goon:"parent"`
	Name          string
	LastMessageID string `datastore:",noindex"`
}
