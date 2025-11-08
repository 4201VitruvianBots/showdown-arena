// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Web routes for generating practice and qualification schedules.

package web

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Team254/cheesy-arena/game"
	"github.com/Team254/cheesy-arena/model"
	"github.com/Team254/cheesy-arena/tournament"
)

// Global vars to hold schedules that are in the process of being generated.
var cachedMatches = make(map[model.MatchType][]model.Match)
var cachedTeamFirstMatches = make(map[model.MatchType]map[int]string)

// Shows the schedule editing page.
func (web *Web) scheduleGetHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	matchTypeString := getMatchType(r)
	matchType, _ := model.MatchTypeFromString(matchTypeString)
	if matchType != model.Practice && matchType != model.Qualification {
		http.Redirect(w, r, "/setup/schedule?matchType=practice", 302)
		return
	}

	web.renderSchedule(w, r, "")
}

// Generates the schedule, presents it for review without saving it, and saves the schedule blocks to the database.
func (web *Web) scheduleGeneratePostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	matchTypeString := getMatchType(r)
	matchType, err := model.MatchTypeFromString(matchTypeString)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	scheduleBlocks, err := getScheduleBlocks(r)
	// Save blocks even if there is an error, so that any good ones are not discarded.
	deleteBlocksErr := web.arena.Database.DeleteScheduleBlocksByMatchType(matchType)
	if deleteBlocksErr != nil {
		handleWebErr(w, err)
		return
	}
	for _, block := range scheduleBlocks {
		block.MatchType = matchType
		createBlockErr := web.arena.Database.CreateScheduleBlock(&block)
		if createBlockErr != nil {
			handleWebErr(w, err)
			return
		}
	}
	if err != nil {
		web.renderSchedule(w, r, "Incomplete or invalid schedule block parameters specified.")
		return
	}

	// Build the schedule.
	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if len(teams) == 0 {
		web.renderSchedule(
			w,
			r,
			"No team list is configured. Set up the list of teams at the event before generating the schedule.",
		)
		return
	}
	if len(teams) < 6 {
		web.renderSchedule(
			w,
			r,
			fmt.Sprintf("There are only %d teams. There must be at least 6 teams to generate a schedule.", len(teams)),
		)
		return
	}

	matches, err := tournament.BuildRandomSchedule(teams, scheduleBlocks, matchType)
	if err != nil {
		web.renderSchedule(w, r, fmt.Sprintf("Error generating schedule: %s.", err.Error()))
		return
	}
	cachedMatches[matchType] = matches

	// Determine each team's first match.
	teamFirstMatches := make(map[int]string)
	for _, match := range matches {
		checkTeam := func(team int) {
			_, ok := teamFirstMatches[team]
			if !ok {
				teamFirstMatches[team] = match.ShortName
			}
		}
		checkTeam(match.Red1)
		checkTeam(match.Red2)
		checkTeam(match.Red3)
		checkTeam(match.Blue1)
		checkTeam(match.Blue2)
		checkTeam(match.Blue3)
	}
	cachedTeamFirstMatches[matchType] = teamFirstMatches

	http.Redirect(w, r, "/setup/schedule?matchType="+matchTypeString, 303)
}

// Saves the generated schedule to the database.
func (web *Web) scheduleSavePostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	matchTypeString := getMatchType(r)
	matchType, err := model.MatchTypeFromString(matchTypeString)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	existingMatches, err := web.arena.Database.GetMatchesByType(matchType, true)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	if len(existingMatches) > 0 {
		web.renderSchedule(
			w,
			r,
			fmt.Sprintf(
				"Can't save schedule because a schedule of %d %s matches already exists. Clear it first on the "+
					"Settings page.",
				len(existingMatches),
				matchType,
			),
		)
		return
	}

	for _, match := range cachedMatches[matchType] {
		err = web.arena.Database.CreateMatch(&match)
		if err != nil {
			handleWebErr(w, err)
			return
		}
	}

	// Back up the database.
	err = web.arena.Database.Backup(web.arena.EventSettings.Name, "post_scheduling")
	if err != nil {
		handleWebErr(w, err)
		return
	}

	http.Redirect(w, r, "/setup/schedule?matchType="+matchTypeString, 303)
}

// Handles CSV upload to replace practice and qualification schedules.
func (web *Web) scheduleUploadPostHandler(w http.ResponseWriter, r *http.Request) {
	if !web.userIsAdmin(w, r) {
		return
	}

	// Parse the multipart form.
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		handleWebErr(w, err)
		return
	}
	file, _, err := r.FormFile("csvFile")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1

	// Delete existing practice and qualification schedules.
	if err = web.deleteMatchDataForType(model.Practice); err != nil {
		handleWebErr(w, err)
		return
	}
	if err = web.deleteMatchDataForType(model.Qualification); err != nil {
		handleWebErr(w, err)
		return
	}

	// Counters for type orders and naming.
	counters := map[model.MatchType]int{model.Practice: 0, model.Qualification: 0}

	location, _ := time.LoadLocation("Local")
	now := time.Now()

	log.Printf("Received schedule upload from %s", r.RemoteAddr)

	// Read CSV rows.
	rowNum := 0
	totalCreated := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			handleWebErr(w, err)
			return
		}
		rowNum++
		// Skip header row if present.
		if rowNum == 1 {
			if len(record) > 0 && strings.Contains(strings.ToLower(record[0]), "time") {
				continue
			}
		}
		if len(record) == 0 {
			continue
		}

		timeStr := strings.TrimSpace(record[0])
		// Normalize whitespace (replace non-breaking and unusual spaces, remove zero-width/BOM) so parsing succeeds.
		timeStr = strings.TrimPrefix(timeStr, "\uFEFF")
		// Replace common non-standard space characters with a normal space.
		timeStr = strings.NewReplacer(
			"\u00A0", " ", // no-break space
			"\u202F", " ", // narrow no-break space
			"\u200B", " ", // zero-width space
		).Replace(timeStr)
		timeStr = strings.Join(strings.Fields(timeStr), " ")
		matchType := model.Qualification
		if strings.Contains(strings.ToLower(timeStr), "practice") {
			matchType = model.Practice
		}
		// Remove any parenthetical suffix like "(practice)".
		if idx := strings.Index(timeStr, "("); idx != -1 {
			timeStr = strings.TrimSpace(timeStr[:idx])
		}

		// Parse time like "3:04 PM".
		parsedTime, err := time.ParseInLocation("3:04 PM", timeStr, location)
		if err != nil {
			// Try alternative layouts as fallback.
			parsedTime, err = time.ParseInLocation("3:04PM", timeStr, location)
			if err != nil {
				parsedTime, err = time.ParseInLocation("15:04", timeStr, location)
				if err != nil {
					// If parsing fails, log and skip this row instead of aborting the entire upload.
					log.Printf("Skipping row %d: invalid time %q: %v", rowNum, timeStr, err)
					continue
				}
			}
		}
		matchTime := time.Date(now.Year(), now.Month(), now.Day(), parsedTime.Hour(), parsedTime.Minute(), 0, 0, location)

		// Parse team ids. Expect columns: Time, Red1, Red2, Blue1, Blue2 (some may be empty).
		getTeam := func(idx int) int {
			if idx < len(record) {
				s := strings.TrimSpace(record[idx])
				if s == "" {
					return 0
				}
				v, err := strconv.Atoi(s)
				if err != nil {
					return 0
				}
				return v
			}
			return 0
		}

		red1 := getTeam(1)
		red2 := getTeam(2)
		blue1 := getTeam(3)
		blue2 := getTeam(4)

		counters[matchType]++
		typeOrder := counters[matchType]

		shortNamePrefix := "Q"
		longNamePrefix := "Qualification"
		if matchType == model.Practice {
			shortNamePrefix = "P"
			longNamePrefix = "Practice"
		}

		match := model.Match{
			Type:      matchType,
			TypeOrder: typeOrder,
			Time:      matchTime,
			LongName:  fmt.Sprintf("%s %d", longNamePrefix, typeOrder),
			ShortName: fmt.Sprintf("%s%d", shortNamePrefix, typeOrder),
			Red1:      red1,
			Red2:      red2,
			Red3:      0,
			Blue1:     blue1,
			Blue2:     blue2,
			Blue3:     0,
			Status:    game.MatchScheduled,
		}

		if err = web.arena.Database.CreateMatch(&match); err != nil {
			handleWebErr(w, err)
			return
		}
		totalCreated++
		log.Printf("Created match: %s (%s)", match.ShortName, match.Time)
	}

	// Back up the DB after upload.
	if err = web.arena.Database.Backup(web.arena.EventSettings.Name, "post_schedule_upload"); err != nil {
		handleWebErr(w, err)
		return
	}
	log.Printf("Schedule upload complete: %d matches created", totalCreated)

	// Refresh in-memory cached matches and team-first map so the schedule page shows the new data.
	for _, mt := range []model.MatchType{model.Practice, model.Qualification} {
		matches, err := web.arena.Database.GetMatchesByType(mt, true)
		if err != nil {
			log.Printf("Failed to load matches for cache after upload: %v", err)
			continue
		}
		cachedMatches[mt] = matches

		teamFirstMatches := make(map[int]string)
		for _, match := range matches {
			checkTeam := func(team int) {
				_, ok := teamFirstMatches[team]
				if !ok && team > 0 {
					teamFirstMatches[team] = match.ShortName
				}
			}
			checkTeam(match.Red1)
			checkTeam(match.Red2)
			checkTeam(match.Red3)
			checkTeam(match.Blue1)
			checkTeam(match.Blue2)
			checkTeam(match.Blue3)
		}
		cachedTeamFirstMatches[mt] = teamFirstMatches
	}

	// Redirect back to schedule page for qualification by default.
	http.Redirect(w, r, "/setup/schedule?matchType=qualification", 303)
}

func (web *Web) renderSchedule(w http.ResponseWriter, r *http.Request, errorMessage string) {
	matchTypeString := getMatchType(r)
	matchType, err := model.MatchTypeFromString(matchTypeString)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	scheduleBlocks, err := web.arena.Database.GetScheduleBlocksByMatchType(matchType)
	if err != nil {
		handleWebErr(w, err)
		return
	}

	teams, err := web.arena.Database.GetAllTeams()
	if err != nil {
		handleWebErr(w, err)
		return
	}
	template, err := web.parseFiles("templates/setup_schedule.html", "templates/base.html")
	if err != nil {
		handleWebErr(w, err)
		return
	}
	data := struct {
		*model.EventSettings
		MatchType        model.MatchType
		ScheduleBlocks   []model.ScheduleBlock
		NumTeams         int
		Matches          []model.Match
		TeamFirstMatches map[int]string
		ErrorMessage     string
	}{
		web.arena.EventSettings,
		matchType,
		scheduleBlocks,
		len(teams),
		cachedMatches[matchType],
		cachedTeamFirstMatches[matchType],
		errorMessage,
	}
	err = template.ExecuteTemplate(w, "base", data)
	if err != nil {
		handleWebErr(w, err)
		return
	}
}

// Converts the post form variables into a slice of schedule blocks.
func getScheduleBlocks(r *http.Request) ([]model.ScheduleBlock, error) {
	numScheduleBlocks, err := strconv.Atoi(r.PostFormValue("numScheduleBlocks"))
	if err != nil {
		return []model.ScheduleBlock{}, err
	}
	var returnErr error
	scheduleBlocks := make([]model.ScheduleBlock, numScheduleBlocks)
	location, _ := time.LoadLocation("Local")
	for i := 0; i < numScheduleBlocks; i++ {
		scheduleBlocks[i].StartTime, err = time.ParseInLocation(
			"2006-01-02 03:04:05 PM", r.PostFormValue(fmt.Sprintf("startTime%d", i)), location,
		)
		if err != nil {
			returnErr = err
		}
		scheduleBlocks[i].NumMatches, err = strconv.Atoi(r.PostFormValue(fmt.Sprintf("numMatches%d", i)))
		if err != nil {
			returnErr = err
		}
		scheduleBlocks[i].MatchSpacingSec, err = strconv.Atoi(r.PostFormValue(fmt.Sprintf("matchSpacingSec%d", i)))
		if err != nil {
			returnErr = err
		}
	}
	return scheduleBlocks, returnErr
}

func getMatchType(r *http.Request) string {
	if matchType, ok := r.URL.Query()["matchType"]; ok {
		return matchType[0]
	}
	return r.PostFormValue("matchType")
}
