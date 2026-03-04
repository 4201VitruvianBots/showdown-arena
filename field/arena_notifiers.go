// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Contains configuration of the publish-subscribe notifiers that allow the arena to push updates to websocket clients.

package field

import (
	"fmt"
	"strconv"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/led"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/playoff"
	"github.com/Team254/cheesy-arena/websocket"
)

type ArenaNotifiers struct {
	AllianceSelectionNotifier          *websocket.Notifier
	AllianceStationDisplayModeNotifier *websocket.Notifier
	ArenaStatusNotifier                *websocket.Notifier
	AudienceDisplayModeNotifier        *websocket.Notifier
	DisplayConfigurationNotifier       *websocket.Notifier
	EventStatusNotifier                *websocket.Notifier
	LowerThirdNotifier                 *websocket.Notifier
	MatchLoadNotifier                  *websocket.Notifier
	MatchTimeNotifier                  *websocket.Notifier
	MatchTimingNotifier                *websocket.Notifier
	PlaySoundNotifier                  *websocket.Notifier
	RealtimeScoreNotifier              *websocket.Notifier
	ReloadDisplaysNotifier             *websocket.Notifier
	ScorePostedNotifier                *websocket.Notifier
	ScoringStatusNotifier              *websocket.Notifier
	HubLedNotifier                     *websocket.Notifier
}

type MatchTimeMessage struct {
	MatchState
	MatchTimeSec int
	RedRearText  string
	BlueRearText string
}

type audienceAllianceScoreFields struct {
	Score        *game.Score
	ScoreSummary *game.ScoreSummary
}

// Instantiates notifiers and configures their message producing methods.
func (arena *Arena) configureNotifiers() {
	arena.AllianceSelectionNotifier = websocket.NewNotifier("allianceSelection", arena.generateAllianceSelectionMessage)
	arena.AllianceStationDisplayModeNotifier = websocket.NewNotifier(
		"allianceStationDisplayMode", arena.generateAllianceStationDisplayModeMessage,
	)
	arena.ArenaStatusNotifier = websocket.NewNotifier("arenaStatus", arena.generateArenaStatusMessage)
	arena.AudienceDisplayModeNotifier = websocket.NewNotifier(
		"audienceDisplayMode", arena.generateAudienceDisplayModeMessage,
	)
	arena.DisplayConfigurationNotifier = websocket.NewNotifier(
		"displayConfiguration", arena.generateDisplayConfigurationMessage,
	)
	arena.EventStatusNotifier = websocket.NewNotifier("eventStatus", arena.generateEventStatusMessage)
	arena.LowerThirdNotifier = websocket.NewNotifier("lowerThird", arena.generateLowerThirdMessage)
	arena.MatchLoadNotifier = websocket.NewNotifier("matchLoad", arena.GenerateMatchLoadMessage)
	arena.MatchTimeNotifier = websocket.NewNotifier("matchTime", arena.generateMatchTimeMessage)
	arena.MatchTimingNotifier = websocket.NewNotifier("matchTiming", arena.generateMatchTimingMessage)
	arena.PlaySoundNotifier = websocket.NewNotifier("playSound", nil)
	arena.RealtimeScoreNotifier = websocket.NewNotifier("realtimeScore", arena.generateRealtimeScoreMessage)
	arena.ReloadDisplaysNotifier = websocket.NewNotifier("reload", nil)
	arena.ScorePostedNotifier = websocket.NewNotifier("scorePosted", arena.GenerateScorePostedMessage)
	arena.ScoringStatusNotifier = websocket.NewNotifier("scoringStatus", arena.generateScoringStatusMessage)
	arena.HubLedNotifier = websocket.NewNotifier("hubLed", arena.generateHubLedMessage)
}

func (arena *Arena) generateAllianceSelectionMessage() any {
	return &struct {
		Alliances        []model.Alliance
		ShowTimer        bool
		TimeRemainingSec int
		RankedTeams      []model.AllianceSelectionRankedTeam
	}{
		arena.AllianceSelectionAlliances,
		arena.AllianceSelectionShowTimer,
		arena.AllianceSelectionTimeRemainingSec,
		arena.AllianceSelectionRankedTeams,
	}
}

func (arena *Arena) generateAllianceStationDisplayModeMessage() any {
	return arena.AllianceStationDisplayMode
}

func (arena *Arena) generateArenaStatusMessage() any {
	redEStops, blueEStops := arena.Esp32.GetTeamEStops()
	redAStops, blueAStops := arena.Esp32.GetTeamAStops()
	stackRed, stackBlue, stackOrange, stackGreen, stackBuzzer := arena.Esp32.GetStackLights()
	matchState := arena.Esp32.GetMatchState()
	redHubState, redMotorDuty, redLedPattern := arena.Esp32.GetRedHubCommand()
	blueHubState, blueMotorDuty, blueLedPattern := arena.Esp32.GetBlueHubCommand()

	return &struct {
		MatchId          int
		AllianceStations map[string]*AllianceStation
		MatchState
		CanStartMatch         bool
		AccessPointStatus     string
		SwitchStatus          string
		RedSCCStatus          string
		BlueSCCStatus         string
		PlcIsHealthy          bool
		FieldEStop            bool
		PlcArmorBlockStatuses map[string]bool
		// Per-node ESP32 I/O data
		ScoreTableIO struct {
			Configured  bool
			Connected   bool
			FieldEStop  bool
			StackRed    bool
			StackBlue   bool
			StackOrange bool
			StackGreen  bool
			StackBuzzer bool
			MatchState  int
		}
		RedEstopsIO struct {
			Configured  bool
			Connected   bool
			EStops      [3]bool
			AStops      [3]bool
			StackRed    bool
			StackBlue   bool
			StackOrange bool
			StackGreen  bool
			StackBuzzer bool
			MatchState  int
		}
		BlueEstopsIO struct {
			Configured  bool
			Connected   bool
			EStops      [3]bool
			AStops      [3]bool
			StackRed    bool
			StackBlue   bool
			StackOrange bool
			StackGreen  bool
			StackBuzzer bool
			MatchState  int
		}
		RedHubIO struct {
			Configured bool
			Connected  bool
			Score      int
			HubState   string
			MotorDuty  float64
			LedPattern string
			MatchState int
		}
		BlueHubIO struct {
			Configured bool
			Connected  bool
			Score      int
			HubState   string
			MotorDuty  float64
			LedPattern string
			MatchState int
		}
	}{
		MatchId:               arena.CurrentMatch.Id,
		AllianceStations:      arena.AllianceStations,
		MatchState:            arena.MatchState,
		CanStartMatch:         arena.checkCanStartMatch() == nil,
		AccessPointStatus:     arena.accessPoint.Status,
		SwitchStatus:          arena.networkSwitch.Status,
		RedSCCStatus:          arena.redSCC.Status,
		BlueSCCStatus:         arena.blueSCC.Status,
		PlcIsHealthy:          arena.Plc.IsHealthy(),
		FieldEStop:            arena.Plc.GetFieldEStop(),
		PlcArmorBlockStatuses: arena.Plc.GetArmorBlockStatuses(),
		ScoreTableIO: struct {
			Configured  bool
			Connected   bool
			FieldEStop  bool
			StackRed    bool
			StackBlue   bool
			StackOrange bool
			StackGreen  bool
			StackBuzzer bool
			MatchState  int
		}{
			Configured:  arena.Esp32.IsScoreTableIOEnabled(),
			Connected:   arena.Esp32.IsScoreTableHealthy(),
			FieldEStop:  arena.Esp32.GetFieldEStop(),
			StackRed:    stackRed,
			StackBlue:   stackBlue,
			StackOrange: stackOrange,
			StackGreen:  stackGreen,
			StackBuzzer: stackBuzzer,
			MatchState:  matchState,
		},
		RedEstopsIO: struct {
			Configured  bool
			Connected   bool
			EStops      [3]bool
			AStops      [3]bool
			StackRed    bool
			StackBlue   bool
			StackOrange bool
			StackGreen  bool
			StackBuzzer bool
			MatchState  int
		}{
			Configured:  arena.Esp32.IsRedEstopsEnabled(),
			Connected:   arena.Esp32.IsRedEstopsHealthy(),
			EStops:      redEStops,
			AStops:      redAStops,
			StackRed:    stackRed,
			StackBlue:   stackBlue,
			StackOrange: stackOrange,
			StackGreen:  stackGreen,
			StackBuzzer: stackBuzzer,
			MatchState:  matchState,
		},
		BlueEstopsIO: struct {
			Configured  bool
			Connected   bool
			EStops      [3]bool
			AStops      [3]bool
			StackRed    bool
			StackBlue   bool
			StackOrange bool
			StackGreen  bool
			StackBuzzer bool
			MatchState  int
		}{
			Configured:  arena.Esp32.IsBlueEstopsEnabled(),
			Connected:   arena.Esp32.IsBlueEstopsHealthy(),
			EStops:      blueEStops,
			AStops:      blueAStops,
			StackRed:    stackRed,
			StackBlue:   stackBlue,
			StackOrange: stackOrange,
			StackGreen:  stackGreen,
			StackBuzzer: stackBuzzer,
			MatchState:  matchState,
		},
		RedHubIO: struct {
			Configured bool
			Connected  bool
			Score      int
			HubState   string
			MotorDuty  float64
			LedPattern string
			MatchState int
		}{
			Configured: arena.Esp32.IsRedHubEnabled(),
			Connected:  arena.Esp32.IsRedHubHealthy(),
			Score:      arena.Esp32.GetRedHubScore(),
			HubState:   redHubState,
			MotorDuty:  redMotorDuty,
			LedPattern: redLedPattern,
			MatchState: matchState,
		},
		BlueHubIO: struct {
			Configured bool
			Connected  bool
			Score      int
			HubState   string
			MotorDuty  float64
			LedPattern string
			MatchState int
		}{
			Configured: arena.Esp32.IsBlueHubEnabled(),
			Connected:  arena.Esp32.IsBlueHubHealthy(),
			Score:      arena.Esp32.GetBlueHubScore(),
			HubState:   blueHubState,
			MotorDuty:  blueMotorDuty,
			LedPattern: blueLedPattern,
			MatchState: matchState,
		},
	}
}

func (arena *Arena) generateAudienceDisplayModeMessage() any {
	return arena.AudienceDisplayMode
}

func (arena *Arena) generateDisplayConfigurationMessage() any {
	// Notify() for this notifier must always called from a method that has a lock on the display mutex.
	// Make a copy of the map to avoid potential data races; otherwise the same map would get iterated through as it is
	// serialized to JSON, outside the mutex lock.
	displaysCopy := make(map[string]Display)
	for displayId, display := range arena.Displays {
		displaysCopy[displayId] = *display
	}
	return displaysCopy
}

func (arena *Arena) generateEventStatusMessage() any {
	return arena.EventStatus
}

func (arena *Arena) generateLowerThirdMessage() any {
	return &struct {
		LowerThird     *model.LowerThird
		ShowLowerThird bool
	}{arena.LowerThird, arena.ShowLowerThird}
}

func (arena *Arena) GenerateMatchLoadMessage() any {
	teams := make(map[string]*model.Team)
	var allTeamIds []int
	for station, allianceStation := range arena.AllianceStations {
		teams[station] = allianceStation.Team
		if allianceStation.Team != nil {
			allTeamIds = append(allTeamIds, allianceStation.Team.Id)
		}
	}

	matchResult, _ := arena.Database.GetMatchResultForMatch(arena.CurrentMatch.Id)
	isReplay := matchResult != nil

	var matchup *playoff.Matchup
	redOffFieldTeams := []*model.Team{}
	blueOffFieldTeams := []*model.Team{}
	if arena.CurrentMatch.Type == model.Playoff {
		matchGroup := arena.PlayoffTournament.MatchGroups()[arena.CurrentMatch.PlayoffMatchGroupId]
		matchup, _ = matchGroup.(*playoff.Matchup)
		redOffFieldTeamIds, blueOffFieldTeamIds, _ := arena.Database.GetOffFieldTeamIds(arena.CurrentMatch)
		for _, teamId := range redOffFieldTeamIds {
			team, _ := arena.Database.GetTeamById(teamId)
			redOffFieldTeams = append(redOffFieldTeams, team)
			allTeamIds = append(allTeamIds, teamId)
		}
		for _, teamId := range blueOffFieldTeamIds {
			team, _ := arena.Database.GetTeamById(teamId)
			blueOffFieldTeams = append(blueOffFieldTeams, team)
			allTeamIds = append(allTeamIds, teamId)
		}
	}

	rankings := make(map[string]int)
	for _, teamId := range allTeamIds {
		ranking, _ := arena.Database.GetRankingForTeam(teamId)
		if ranking != nil {
			rankings[strconv.Itoa(teamId)] = ranking.Rank
		}
	}

	return &struct {
		Match             *model.Match
		AllowSubstitution bool
		IsReplay          bool
		Teams             map[string]*model.Team
		Rankings          map[string]int
		Matchup           *playoff.Matchup
		RedOffFieldTeams  []*model.Team
		BlueOffFieldTeams []*model.Team
		BreakDescription  string
	}{
		arena.CurrentMatch,
		arena.CurrentMatch.ShouldAllowSubstitution(),
		isReplay,
		teams,
		rankings,
		matchup,
		redOffFieldTeams,
		blueOffFieldTeams,
		arena.breakDescription,
	}
}

func (arena *Arena) generateMatchTimeMessage() any {
	matchTimeSec := arena.MatchTimeSec()
	matchTimeSecInt := int(matchTimeSec)

	// Generate the countdown string.
	var countdownSec int
	switch arena.MatchState {
	case PreMatch:
		if arena.AudienceDisplayMode == "allianceSelection" {
			countdownSec = arena.AllianceSelectionTimeRemainingSec
		} else {
			countdownSec = game.MatchTiming.AutoDurationSec
		}
	case StartMatch, WarmupPeriod:
		countdownSec = game.MatchTiming.AutoDurationSec
	case AutoPeriod:
		countdownSec = game.MatchTiming.WarmupDurationSec + game.MatchTiming.AutoDurationSec - matchTimeSecInt
	case TeleopPeriod:
		countdownSec = game.MatchTiming.WarmupDurationSec + game.MatchTiming.AutoDurationSec +
			game.MatchTiming.TeleopDurationSec + game.MatchTiming.PauseDurationSec - matchTimeSecInt
	case TimeoutActive:
		countdownSec = game.MatchTiming.TimeoutDurationSec - matchTimeSecInt
	default:
		countdownSec = 0
	}
	if countdownSec < 0 {
		countdownSec = 0
	}
	countdown := fmt.Sprintf("%d:%02d", countdownSec/60, countdownSec%60)

	return MatchTimeMessage{
		MatchState:   arena.MatchState,
		MatchTimeSec: matchTimeSecInt,
		RedRearText:  generateInMatchTeamRearText(arena, true, countdown, matchTimeSec),
		BlueRearText: generateInMatchTeamRearText(arena, false, countdown, matchTimeSec),
	}
}

func (arena *Arena) generateMatchTimingMessage() any {
	return &game.MatchTiming
}

func (arena *Arena) generateRealtimeScoreMessage() any {
	fields := struct {
		Red           *audienceAllianceScoreFields
		Blue          *audienceAllianceScoreFields
		RedCards      map[string]string
		BlueCards     map[string]string
		AutoTieWinner string
		MatchState
	}{
		getAudienceAllianceScoreFields(arena.RedRealtimeScore, arena.RedScoreSummary()),
		getAudienceAllianceScoreFields(arena.BlueRealtimeScore, arena.BlueScoreSummary()),
		arena.RedRealtimeScore.Cards,
		arena.BlueRealtimeScore.Cards,
		arena.autoTieWinner,
		arena.MatchState,
	}
	return &fields
}

func (arena *Arena) GenerateScorePostedMessage() any {
	redScoreSummary := arena.SavedMatchResult.RedScoreSummary()
	blueScoreSummary := arena.SavedMatchResult.BlueScoreSummary()
	redRankingPoints := redScoreSummary.BonusRankingPoints
	blueRankingPoints := blueScoreSummary.BonusRankingPoints
	switch arena.SavedMatch.Status {
	case game.RedWonMatch:
		redRankingPoints += 3
	case game.BlueWonMatch:
		blueRankingPoints += 3
	case game.TieMatch:
		redRankingPoints++
		blueRankingPoints++
	}

	// For playoff matches, summarize the state of the series.
	var redWins, blueWins int
	var redDestination, blueDestination string
	redOffFieldTeamIds := []int{}
	blueOffFieldTeamIds := []int{}
	if arena.SavedMatch.Type == model.Playoff {
		matchGroup := arena.PlayoffTournament.MatchGroups()[arena.SavedMatch.PlayoffMatchGroupId]
		if matchup, ok := matchGroup.(*playoff.Matchup); ok {
			redWins = matchup.RedAllianceWins
			blueWins = matchup.BlueAllianceWins
			redDestination = matchup.RedAllianceDestination()
			blueDestination = matchup.BlueAllianceDestination()
		}
		redOffFieldTeamIds, blueOffFieldTeamIds, _ = arena.Database.GetOffFieldTeamIds(arena.SavedMatch)
	}

	redRankings := map[int]*game.Ranking{
		arena.SavedMatch.Red1: nil, arena.SavedMatch.Red2: nil, arena.SavedMatch.Red3: nil,
	}
	blueRankings := map[int]*game.Ranking{
		arena.SavedMatch.Blue1: nil, arena.SavedMatch.Blue2: nil, arena.SavedMatch.Blue3: nil,
	}
	for index, ranking := range arena.SavedRankings {
		if _, ok := redRankings[ranking.TeamId]; ok {
			redRankings[ranking.TeamId] = &arena.SavedRankings[index]
		}
		if _, ok := blueRankings[ranking.TeamId]; ok {
			blueRankings[ranking.TeamId] = &arena.SavedRankings[index]
		}
	}

	return &struct {
		Match               *model.Match
		RedScoreSummary     *game.ScoreSummary
		BlueScoreSummary    *game.ScoreSummary
		RedRankingPoints    int
		BlueRankingPoints   int
		RedFouls            []game.Foul
		BlueFouls           []game.Foul
		RulesViolated       map[int]*game.Rule
		RedCards            map[string]string
		BlueCards           map[string]string
		RedRankings         map[int]*game.Ranking
		BlueRankings        map[int]*game.Ranking
		RedOffFieldTeamIds  []int
		BlueOffFieldTeamIds []int
		RedWon              bool
		BlueWon             bool
		RedWins             int
		BlueWins            int
		RedDestination      string
		BlueDestination     string
		CoopertitionEnabled bool
	}{
		arena.SavedMatch,
		redScoreSummary,
		blueScoreSummary,
		redRankingPoints,
		blueRankingPoints,
		arena.SavedMatchResult.RedScore.Fouls,
		arena.SavedMatchResult.BlueScore.Fouls,
		getRulesViolated(arena.SavedMatchResult.RedScore.Fouls, arena.SavedMatchResult.BlueScore.Fouls),
		arena.SavedMatchResult.RedCards,
		arena.SavedMatchResult.BlueCards,
		redRankings,
		blueRankings,
		redOffFieldTeamIds,
		blueOffFieldTeamIds,
		arena.SavedMatch.Status == game.RedWonMatch,
		arena.SavedMatch.Status == game.BlueWonMatch,
		redWins,
		blueWins,
		redDestination,
		blueDestination,
		false, // Coopertition not used in REBUILT
	}
}

func (arena *Arena) generateScoringStatusMessage() any {
	type positionStatus struct {
		Ready          bool
		NumPanels      int
		NumPanelsReady int
	}
	getStatusForPosition := func(position string) positionStatus {
		return positionStatus{
			Ready:          arena.positionPostMatchScoreReady(position),
			NumPanels:      arena.ScoringPanelRegistry.GetNumPanels(position),
			NumPanelsReady: arena.GetNumScoreCommitted(position),
		}
	}

	return &struct {
		RefereeScoreReady bool
		PositionStatuses  map[string]positionStatus
	}{
		arena.RedRealtimeScore.FoulsCommitted && arena.BlueRealtimeScore.FoulsCommitted,
		map[string]positionStatus{
			"red_near":  getStatusForPosition("red_near"),
			"red_far":   getStatusForPosition("red_far"),
			"blue_near": getStatusForPosition("blue_near"),
			"blue_far":  getStatusForPosition("blue_far"),
		},
	}
}

func (arena *Arena) generateHubLedMessage() any {
	return &struct {
		Red  led.Color
		Blue led.Color
	}{
		arena.RedHubLeds.GetColor(),
		arena.BlueHubLeds.GetColor(),
	}
}

func getAudienceAllianceScoreFields(
	allianceScore *RealtimeScore,
	allianceScoreSummary *game.ScoreSummary,
) *audienceAllianceScoreFields {
	fields := new(audienceAllianceScoreFields)
	fields.Score = &allianceScore.CurrentScore
	fields.ScoreSummary = allianceScoreSummary
	return fields
}

// Produce a map of rules that were violated by either alliance so that they are available to the announcer.
func getRulesViolated(redFouls, blueFouls []game.Foul) map[int]*game.Rule {
	rules := make(map[int]*game.Rule)
	for _, foul := range redFouls {
		rules[foul.RuleId] = game.GetRuleById(foul.RuleId)
	}
	for _, foul := range blueFouls {
		rules[foul.RuleId] = game.GetRuleById(foul.RuleId)
	}
	return rules
}
