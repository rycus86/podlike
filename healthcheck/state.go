package healthcheck

const (
	StateUnknown   = iota
	StateStarting  = iota
	StateUnhealthy = iota
	StateHealthy   = iota
)

var (
	stateNames = map[int]string{
		StateUnknown:   "unknown",
		StateStarting:  "starting",
		StateUnhealthy: "unhealthy",
		StateHealthy:   "healthy",
	}

	currentStates = map[string]int{}
)

func getCurrentState() int {
	state := StateHealthy

	for _, s := range currentStates {
		if s < state {
			state = s
		}
	}

	return state
}

func getStateName(state int) string {
	return stateNames[state]
}

func NameToValue(state string) int {
	for value, name := range stateNames {
		if name == state {
			return value
		}
	}

	return StateUnknown
}

func Initialize(component string, state int) {
	currentStates[component] = state
}

func GetState() string {
	return getStateName(getCurrentState())
}

func SetState(component string, state int) {
	// only initialized components can set their state
	if _, ok := currentStates[component]; ok {
		currentStates[component] = state
	}
}
