package entity

type Guild struct {
	ID   string `datastore:"-" goon:"id"`
	Name string `datastore:",noindex"`
}
