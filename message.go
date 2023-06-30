package ircd

type Message struct {
	CommandType MessageType
	Raw         string
	Tags        map[string]interface{}
	Prefix      string
	Command     string
	Params      []string
}
