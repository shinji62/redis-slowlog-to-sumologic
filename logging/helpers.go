package logging

var (
	// LevelsInSlice levels in string format
	LevelsInSlice = [...]string{"TRACE", "INFO", "WARNING", "ERROR"}
)

// SliceToLevels convert levels in slice to a LogLevel
func SliceToLevels(sl []string) int {
	lvl := 0
	for _, s := range sl {
		switch s {
		case LevelsInSlice[0]:
			lvl = lvl | LogTrace
		case LevelsInSlice[1]:
			lvl = lvl | LogInfo
		case LevelsInSlice[2]:
			lvl = lvl | LogWarning
		case LevelsInSlice[3]:
			lvl = lvl | LogError
		}
	}
	return lvl
}
