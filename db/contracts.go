package db

type Server struct {
	ID     int
	Name   string
	URL    string
	Client string
	Topics []Topic
	Subs   []Sub
}

type Topic struct {
	ID        int
	ServerID  int
	TopicName string
}

type Sub struct {
	ID       int
	ServerID int
	SubName  string
}
