package ircd

import "fmt"

type notice interface {
	format() string
}

type noticeAuth struct {
	client  string
	message string
}

func (n noticeAuth) format() string {
	return fmt.Sprintf(
		"NOTICE %s :AUTH :*** %s",
		n.client, n.message,
	)
}
