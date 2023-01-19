package api

type ok struct{}

type notOk struct{}

// Param ...
type Param struct {
	Key      string
	Required bool
}

// Status ...
type Status string

// Statuses ...
const (
	StatusOK         Status = "200"
	StatusBadRequest Status = "400"
)

// Route ...
type Route struct {
	ID     string
	IDD    func(...any) any
	Path   string
	Desc   string
	Body   *any
	Params []Param
	Res    map[Status]any
}

// Group ...
type Group struct {
	Name   string
	Prefix string
	Routes []Route
}
