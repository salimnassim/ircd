package ircd

type message struct {
	Raw     string
	Tags    map[string]any
	Prefix  string
	Command string
	Params  []string
}
