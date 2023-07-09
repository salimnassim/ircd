package ircd

type Message struct {
	Raw     string
	Tags    map[string]interface{}
	Prefix  string
	Command string
	Params  []string
}
