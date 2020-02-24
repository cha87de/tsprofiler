package models

// TSProfile contains the resulting statistical profile
type TSProfile struct {
	Name       string     `json:"name"`
	PeriodTree PeriodTree `json:"periodTree"`
	Phases     Phases     `json:"phases"`
	Settings   Settings   `json:"settings"`
}

// Likeliness returns how likely [0,1] the current value appears according to the root tx matrix
func (profile *TSProfile) Likeliness(oldstates [][]TSState, newstate []TSState) float32 {
	likelinessSum := float32(0)
	likelinessCount := 0

	metricPhases := profile.PeriodTree.Root.TxMatrix
	for _, phaseTx := range metricPhases {

		fromStates := make([]TSState, 0)
		var toState TSState

		// find from
		for _, oldstateHistory := range oldstates {
			for _, s := range oldstateHistory {
				if s.Metric == phaseTx.Metric {
					fromStates = append(fromStates, s)
					break
				}
			}
		}

		// find to
		for _, s := range newstate {
			if s.Metric == phaseTx.Metric {
				toState = s
				break
			}
		}

		likeliness := phaseTx.Likeliness(fromStates, toState)
		likelinessSum += likeliness
		likelinessCount++
	}

	return likelinessSum / float32(likelinessCount)
}

// LikelinessPhase returns how likely [0,1] the current value appears according to the profile's phase
func (profile *TSProfile) LikelinessPhase(currentPhase int, oldstates [][]TSState, newstate []TSState) float32 {
	likelinessSum := float32(0)
	likelinessCount := 0

	metricPhases := profile.Phases.Phases[currentPhase]
	for _, phaseTx := range metricPhases {

		fromStates := make([]TSState, 0)
		var toState TSState

		// find from
		for _, oldstateHistory := range oldstates {
			for _, s := range oldstateHistory {
				if s.Metric == phaseTx.Metric {
					fromStates = append(fromStates, s)
					break
				}
			}
		}

		// find to
		for _, s := range newstate {
			if s.Metric == phaseTx.Metric {
				toState = s
				break
			}
		}

		likeliness := phaseTx.Likeliness(fromStates, toState)
		likelinessSum += likeliness
		likelinessCount++
	}

	return likelinessSum / float32(likelinessCount)
}
