package ircd

func handleMode(s *server, c clienter, m message) {
	if !m.isTargetChannel() {
		handleModeClient(s, c, m)
		return
	}

	if m.isTargetChannel() {
		handleModeChannel(s, c, m)
		return
	}
}
