// Copyright 2017 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Game-specific period timing.

package game

import "time"

const (
	TeleopGracePeriodSec     = 3
	TransitionDurationSec    = 10 // First 10 seconds of teleop when both hubs are active (transition period)
	ShiftDurationSec         = 25
	HubScoringGracePeriodSec = 3  // Grace period after hub deactivates to still count FUEL as active
	EndGameDurationSec       = 30 // Last 30 seconds of teleop when both hubs are active
)

var MatchTiming = struct {
	WarmupDurationSec           int
	AutoDurationSec             int
	PauseDurationSec            int
	TeleopDurationSec           int
	WarningRemainingDurationSec int
	TimeoutDurationSec          int
}{0, 20, 3, 140, 20, 0}

func GetDurationToAutoEnd() time.Duration {
	return time.Duration(MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec) * time.Second
}

func GetDurationToTeleopStart() time.Duration {
	return time.Duration(
		MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec,
	) * time.Second
}

func GetDurationToTeleopEnd() time.Duration {
	return time.Duration(
		MatchTiming.WarmupDurationSec+MatchTiming.AutoDurationSec+MatchTiming.PauseDurationSec+
			MatchTiming.TeleopDurationSec,
	) * time.Second
}

// GetCurrentShift returns which shift number (0-indexed) is currently active based on match time.
// Returns -1 if not in teleop period.
func GetCurrentShift(matchTimeSec float64) int {
	teleopStartSec := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)
	teleopEndSec := teleopStartSec + float64(MatchTiming.TeleopDurationSec)

	if matchTimeSec < teleopStartSec || matchTimeSec >= teleopEndSec {
		return -1
	}

	teleopElapsedSec := matchTimeSec - teleopStartSec
	return int(teleopElapsedSec / ShiftDurationSec)
}

// IsRedHubActive returns true if the red alliance's hub is currently active.
// During auto and pause, both hubs are active.
// During the first 10 seconds of teleop (transition period), both hubs are active.
// During the last 30 seconds of teleop (END GAME), both hubs are active.
// After the transition, the alliance that LOST auto has their hub active first.
// If Red won auto: Red is INACTIVE first, then alternates every 25 seconds.
// If Blue won auto or tie: Red is ACTIVE first, then alternates every 25 seconds.
func IsRedHubActive(matchTimeSec float64, redWonAuto bool) bool {
	teleopStartSec := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)
	teleopEndSec := teleopStartSec + float64(MatchTiming.TeleopDurationSec)
	transitionEndSec := teleopStartSec + float64(TransitionDurationSec)

	// During auto and pause, both hubs are active
	if matchTimeSec < teleopStartSec {
		return true
	}

	// During transition period (first 10 seconds of teleop), both hubs are active
	if matchTimeSec < transitionEndSec {
		return true
	}

	// During END GAME (last 30 seconds of teleop), both hubs are active
	if matchTimeSec >= teleopEndSec-EndGameDurationSec && matchTimeSec < teleopEndSec {
		return true
	}

	// After the match ends, hubs are not active
	if matchTimeSec >= teleopEndSec {
		return false
	}

	// Calculate which alternating shift we're in (after transition period)
	// Shift 0 = 10-35 sec into teleop, Shift 1 = 35-60 sec, etc.
	postTransitionSec := matchTimeSec - transitionEndSec
	if postTransitionSec < 0 {
		return false
	}
	shift := int(postTransitionSec / ShiftDurationSec)

	if redWonAuto {
		// Red won auto, so Red hub is INACTIVE for first alternating shift
		// Red is INACTIVE on even shifts (0, 2, 4...), ACTIVE on odd shifts (1, 3, 5...)
		return shift%2 == 1
	} else {
		// Blue won auto or tie, so Red hub is ACTIVE for first alternating shift
		// Red is ACTIVE on even shifts (0, 2, 4...), INACTIVE on odd shifts (1, 3, 5...)
		return shift%2 == 0
	}
}

// IsBlueHubActive returns true if the blue alliance's hub is currently active.
// During auto and pause, both hubs are active.
// During the first 10 seconds of teleop (transition period), both hubs are active.
// During the last 30 seconds of teleop (END GAME), both hubs are active.
// After the transition, the alliance that LOST auto has their hub active first.
// If Blue won auto: Blue is INACTIVE first, then alternates every 25 seconds.
// If Red won auto or tie: Blue is ACTIVE first, then alternates every 25 seconds.
func IsBlueHubActive(matchTimeSec float64, blueWonAuto bool) bool {
	teleopStartSec := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)
	teleopEndSec := teleopStartSec + float64(MatchTiming.TeleopDurationSec)
	transitionEndSec := teleopStartSec + float64(TransitionDurationSec)

	// During auto and pause, both hubs are active
	if matchTimeSec < teleopStartSec {
		return true
	}

	// During transition period (first 10 seconds of teleop), both hubs are active
	if matchTimeSec < transitionEndSec {
		return true
	}

	// During END GAME (last 30 seconds of teleop), both hubs are active
	if matchTimeSec >= teleopEndSec-EndGameDurationSec && matchTimeSec < teleopEndSec {
		return true
	}

	// After the match ends, hubs are not active
	if matchTimeSec >= teleopEndSec {
		return false
	}

	// Calculate which alternating shift we're in (after transition period)
	// Shift 0 = 10-35 sec into teleop, Shift 1 = 35-60 sec, etc.
	postTransitionSec := matchTimeSec - transitionEndSec
	if postTransitionSec < 0 {
		return false
	}
	shift := int(postTransitionSec / ShiftDurationSec)

	if blueWonAuto {
		// Blue won auto, so Blue hub is INACTIVE for first alternating shift
		// Blue is INACTIVE on even shifts (0, 2, 4...), ACTIVE on odd shifts (1, 3, 5...)
		return shift%2 == 1
	} else {
		// Red won auto or tie, so Blue hub is ACTIVE for first alternating shift
		// Blue is ACTIVE on even shifts (0, 2, 4...), INACTIVE on odd shifts (1, 3, 5...)
		return shift%2 == 0
	}
}

// IsRedHubActiveForScoring returns true if the red alliance's hub should accept FUEL as "active".
// This includes the grace period after the hub deactivates to account for FUEL in flight.
// The grace period applies even after the match ends.
func IsRedHubActiveForScoring(matchTimeSec float64, redWonAuto bool) bool {
	teleopStartSec := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)
	teleopEndSec := teleopStartSec + float64(MatchTiming.TeleopDurationSec)
	transitionEndSec := teleopStartSec + float64(TransitionDurationSec)

	// Check if hub is currently active
	if IsRedHubActive(matchTimeSec, redWonAuto) {
		return true
	}

	// Check if we're in the grace period after the match ends
	if matchTimeSec >= teleopEndSec && matchTimeSec < teleopEndSec+HubScoringGracePeriodSec {
		// Both hubs are active during END GAME, so grace period applies
		return true
	}

	// Check if we're in the grace period after the transition period ends
	if matchTimeSec >= transitionEndSec && matchTimeSec < transitionEndSec+HubScoringGracePeriodSec {
		// Both hubs are active during transition, so grace period applies
		return true
	}

	// Calculate which alternating shift we're in (after transition period)
	postTransitionSec := matchTimeSec - transitionEndSec
	if postTransitionSec < 0 {
		return false
	}
	shift := int(postTransitionSec / ShiftDurationSec)
	timeInShift := postTransitionSec - float64(shift)*ShiftDurationSec

	// If we're within the grace period after a shift transition, check if the hub was active in the previous moment
	if timeInShift < HubScoringGracePeriodSec {
		// Check if the hub was active at the end of the previous shift
		previousShift := shift - 1
		if previousShift < 0 {
			// Previous shift was the transition period, which had both hubs active
			return true
		}

		if redWonAuto {
			// Red won auto, so Red is INACTIVE on even shifts, ACTIVE on odd shifts
			return previousShift%2 == 1
		} else {
			// Blue won auto or tie, so Red is ACTIVE on even shifts, INACTIVE on odd shifts
			return previousShift%2 == 0
		}
	}

	return false
}

// IsBlueHubActiveForScoring returns true if the blue alliance's hub should accept FUEL as "active".
// This includes the grace period after the hub deactivates to account for FUEL in flight.
// The grace period applies even after the match ends.
func IsBlueHubActiveForScoring(matchTimeSec float64, blueWonAuto bool) bool {
	teleopStartSec := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)
	teleopEndSec := teleopStartSec + float64(MatchTiming.TeleopDurationSec)
	transitionEndSec := teleopStartSec + float64(TransitionDurationSec)

	// Check if hub is currently active
	if IsBlueHubActive(matchTimeSec, blueWonAuto) {
		return true
	}

	// Check if we're in the grace period after the match ends
	if matchTimeSec >= teleopEndSec && matchTimeSec < teleopEndSec+HubScoringGracePeriodSec {
		// Both hubs are active during END GAME, so grace period applies
		return true
	}

	// Check if we're in the grace period after the transition period ends
	if matchTimeSec >= transitionEndSec && matchTimeSec < transitionEndSec+HubScoringGracePeriodSec {
		// Both hubs are active during transition, so grace period applies
		return true
	}

	// Calculate which alternating shift we're in (after transition period)
	postTransitionSec := matchTimeSec - transitionEndSec
	if postTransitionSec < 0 {
		return false
	}
	shift := int(postTransitionSec / ShiftDurationSec)
	timeInShift := postTransitionSec - float64(shift)*ShiftDurationSec

	// If we're within the grace period after a shift transition, check if the hub was active in the previous moment
	if timeInShift < HubScoringGracePeriodSec {
		// Check if the hub was active at the end of the previous shift
		previousShift := shift - 1
		if previousShift < 0 {
			// Previous shift was the transition period, which had both hubs active
			return true
		}

		if blueWonAuto {
			// Blue won auto, so Blue is INACTIVE on even shifts, ACTIVE on odd shifts
			return previousShift%2 == 1
		} else {
			// Red won auto or tie, so Blue is ACTIVE on even shifts, INACTIVE on odd shifts
			return previousShift%2 == 0
		}
	}

	return false
}

// GetShiftIndicator returns the shift indicator letter for the team sign display.
// "A" = Auto (both hubs active), "T" = Transition (both hubs active),
// "R" = Red hub active only, "B" = Blue hub active only, "E" = End Game (both hubs active).
// Returns "" if not in an active match period.
func GetShiftIndicator(matchTimeSec float64, isRedHubActive, isBlueHubActive bool) string {
	teleopStartSec := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)
	teleopEndSec := teleopStartSec + float64(MatchTiming.TeleopDurationSec)
	transitionEndSec := teleopStartSec + float64(TransitionDurationSec)

	// Before teleop starts (auto + pause), both hubs are active
	if matchTimeSec < teleopStartSec {
		return "A"
	}

	// Transition period (first 10 seconds of teleop)
	if matchTimeSec < transitionEndSec {
		return "T"
	}

	// End Game (last 30 seconds of teleop)
	if matchTimeSec >= teleopEndSec-float64(EndGameDurationSec) && matchTimeSec < teleopEndSec {
		return "E"
	}

	// After match end
	if matchTimeSec >= teleopEndSec {
		return ""
	}

	// Alternating shifts - determine which hub is active
	if isRedHubActive && isBlueHubActive {
		return "E" // Shouldn't happen during alternating, but safety fallback
	}
	if isRedHubActive {
		return "R"
	}
	if isBlueHubActive {
		return "B"
	}

	return ""
}

// GetShiftTimeRemaining returns the number of seconds remaining in the current shift/period.
// During Auto: auto time remaining. During Transition: transition time remaining.
// During alternating shifts: time remaining in the current 25-second shift.
// During End Game: end game time remaining. Returns 0 if not in an active match period.
func GetShiftTimeRemaining(matchTimeSec float64) int {
	teleopStartSec := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec + MatchTiming.PauseDurationSec)
	teleopEndSec := teleopStartSec + float64(MatchTiming.TeleopDurationSec)
	transitionEndSec := teleopStartSec + float64(TransitionDurationSec)

	// Auto period (including warmup)
	autoEndSec := float64(MatchTiming.WarmupDurationSec + MatchTiming.AutoDurationSec)
	if matchTimeSec < autoEndSec {
		return int(autoEndSec - matchTimeSec)
	}

	// Pause period - show time remaining until teleop
	if matchTimeSec < teleopStartSec {
		return int(teleopStartSec - matchTimeSec)
	}

	// Transition period
	if matchTimeSec < transitionEndSec {
		return int(transitionEndSec - matchTimeSec)
	}

	// End Game (last 30 seconds of teleop)
	if matchTimeSec >= teleopEndSec-float64(EndGameDurationSec) && matchTimeSec < teleopEndSec {
		return int(teleopEndSec - matchTimeSec)
	}

	// After match end
	if matchTimeSec >= teleopEndSec {
		return 0
	}

	// Alternating shifts - calculate time remaining in current 25-second shift
	postTransitionSec := matchTimeSec - transitionEndSec
	// But cap at the start of end game
	endGameStartSec := teleopEndSec - float64(EndGameDurationSec)
	timeInShift := postTransitionSec - float64(int(postTransitionSec/ShiftDurationSec))*float64(ShiftDurationSec)
	shiftEnd := matchTimeSec + (float64(ShiftDurationSec) - timeInShift)
	if shiftEnd > endGameStartSec {
		// Current shift is cut short by end game
		return int(endGameStartSec - matchTimeSec)
	}
	return int(float64(ShiftDurationSec) - timeInShift)
}
