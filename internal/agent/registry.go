package agent

var AllAgents []Agent

func Register(a Agent) {
	AllAgents = append(AllAgents, a)
}
