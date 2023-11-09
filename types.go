package kvorm

type DbConfig struct {
	Host      string
	Port      int
	User      string
	Password  string
	Name      string
	PollCount int32
}
