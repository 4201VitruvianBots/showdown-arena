// Copyright 2014 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side methods for the alliance station display.

var station = "";
var blinkInterval;
var currentScreen = "blank";
var websocket;

// Handles a websocket message to change which screen is displayed.
var handleAllianceStationDisplayMode = function (targetScreen) {
  // Keep the match mode visible for alliance station displays but force the
  // layout to pre-match presentation by ensuring the match container is used
  // and the position remains correct. Stop switching to other modes like
  // logo/fieldReset/signalCount which would hide the team number.
  currentScreen = targetScreen;
  if (station === "") {
    // Don't do anything if this screen hasn't been assigned a position yet.
  } else {
    var body = $("body");
    // Ensure the match container is visible so preMatch team number is shown.
    body.attr("data-mode", "match");
    // Keep the position mapping so the layout remains correct for the station.
    switch (station[1]) {
      case "1":
        body.attr("data-position", "right");
        break;
      case "2":
        body.attr("data-position", "middle");
        break;
      case "3":
        body.attr("data-position", "left");
        break;
    }
    // Force the match state to PRE_MATCH so the preMatch (team number) section
    // remains visible even when the match is running.
    $("#match").attr("data-state", "PRE_MATCH");
  }
};

// Handles a websocket message to update the team to display.
var handleMatchLoad = function (data) {
  if (station !== "") {
    var team = data.Teams[station];
    if (team) {
      $("#teamNumber").text(team.Id);
      $("#teamNameText").attr("data-alliance-bg", station[0]).text(team.Nickname);
      
      // Set table number based on station
      var tableNumber;
      if (station === "R1") tableNumber = "Table 1";
      else if (station === "R2") tableNumber = "Table 2";
      else if (station === "B1") tableNumber = "Table 3";
      else if (station === "B2") tableNumber = "Table 4";
      $("#tableNumber").text(tableNumber);
      // Adjust the rotated label horizontal translation: Tables 1 and 3
      // should use a positive X translation (65px) so the text sits
      // inside the viewport; other tables use the default negative
      // translation which centers the rotated text.
      if (tableNumber === "Table 1" || tableNumber === "Table 3") {
        // Use the jQuery .css(property, value) form (not a single "prop: val" string).
        // Move the rotated label inward for Tables 1 and 3 using a positive X
        // offset (65px) and center it vertically with translateY(-50%).
        $("#tableNumber").css("transform", "rotate(90deg) translateX(-65%) translateY(-90%) translateY(-3950px)");
      } else {
        // For the other tables, center the rotated label horizontally and
        // vertically using -50% translations.
        $("#tableNumber").css("transform", "rotate(90deg) translateX(-65%) translateY(-90%)");
      }

      var ranking = data.Rankings[team.Id];
      if (ranking && data.Match.Type === matchTypeQualification) {
        $("#teamRank").attr("data-alliance-bg", station[0]).text(ranking);
      } else {
        $("#teamRank").attr("data-alliance-bg", station[0]).text("");
      }
    } else {
      $("#teamNumber").text("");
      $("#teamNameText").attr("data-alliance-bg", station[0]).text("");
      $("#teamRank").attr("data-alliance-bg", station[0]).text("");
    }

    // Populate extra alliance info if this is a playoff match.
    let playoffAlliance = data.Match.PlayoffRedAlliance;
    let offFieldTeams = data.RedOffFieldTeams;
    if (station[0] === "B") {
      playoffAlliance = data.Match.PlayoffBlueAlliance;
      offFieldTeams = data.BlueOffFieldTeams;
    }
    if (playoffAlliance > 0) {
      let playoffAllianceInfo = `Alliance ${playoffAlliance}`;
      if (offFieldTeams.length) {
        playoffAllianceInfo += `&emsp; Not on field: ${offFieldTeams.map(team => team.Id).join(", ")}`;
      }
      $("#playoffAllianceInfo").html(playoffAllianceInfo);
    } else {
      $("#playoffAllianceInfo").text("");
    }
  }
};

// Handles a websocket message to update the team connection status.
var handleArenaStatus = function (data) {
  stationStatus = data.AllianceStations[station];
  var blink = false;
  if (stationStatus && stationStatus.Bypass) {
    $("#match").attr("data-status", "bypass");
  } else if (stationStatus) {
    if (!stationStatus.DsConn || !stationStatus.DsConn.DsLinked) {
      $("#match").attr("data-status", station[0]);
    } else if (!stationStatus.DsConn.RobotLinked) {
      blink = true;
      if (!blinkInterval) {
        blinkInterval = setInterval(function () {
          var status = $("#match").attr("data-status");
          $("#match").attr("data-status", (status === "") ? station[0] : "");
        }, 250);
      }
    } else {
      $("#match").attr("data-status", "");
    }
  }

  if (!blink && blinkInterval) {
    clearInterval(blinkInterval);
    blinkInterval = null;
  }
};

// Handles a websocket message to update the match time countdown.
var handleMatchTime = function (data) {
  translateMatchTime(data, function (matchState, matchStateText, countdownSec) {
    // For non-alliance displays (station starting with 'N') we want to force
    // an in-match presentation so they show time/score. For alliance station
    // displays (e.g., R1/B2) we should not change the #match data-state here
    // because that would switch the visible mode to inMatch. Instead, only
    // update the time text so preMatch team number remains visible.
    var isNonAlliance = station[0] === "N";
    if (isNonAlliance) {
      // Pin the state for a non-alliance display to an in-match state, so as to always show time or score.
      matchState = "TELEOP_PERIOD";
    }
    var countdownString = String(countdownSec % 60);
    if (countdownString.length === 1) {
      countdownString = "0" + countdownString;
    }
    countdownString = Math.floor(countdownSec / 60) + ":" + countdownString;
    $("#timeRemaining").text(countdownString);
    if (isNonAlliance) {
      $("#match").attr("data-state", matchState);
    }
  });
};

// Handles a websocket message to update the match score.
var handleRealtimeScore = function (data) {
  $("#redScore").text(
    data.Red.ScoreSummary.Score - data.Red.ScoreSummary.BargePoints
  );
  $("#blueScore").text(
    data.Blue.ScoreSummary.Score - data.Blue.ScoreSummary.BargePoints
  );
};

$(function () {
  // Read the configuration for this display from the URL query string.
  var urlParams = new URLSearchParams(window.location.search);
  station = urlParams.get("station");

  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/displays/alliance_station/websocket", {
    allianceStationDisplayMode: function (event) {
      handleAllianceStationDisplayMode(event.data);
    },
    arenaStatus: function (event) {
      handleArenaStatus(event.data);
    },
    matchLoad: function (event) {
      handleMatchLoad(event.data);
    },
    matchTime: function (event) {
      handleMatchTime(event.data);
    },
    matchTiming: function (event) {
      handleMatchTiming(event.data);
    },
    realtimeScore: function (event) {
      handleRealtimeScore(event.data);
    }
  });
});
