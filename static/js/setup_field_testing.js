// Copyright 2018 Team 254. All Rights Reserved.
// Author: pat@patfairbank.com (Patrick Fairbank)
//
// Client-side logic for the Field Testing page.

var websocket;

// Sends a websocket message to play a given game sound on the audience display.
var playSound = function (sound) {
  websocket.send("playSound", sound);
};

// Handles a websocket message to update the PLC IO status.
var handlePlcIoChange = function (data) {
  $.each(data.Inputs, function (index, input) {
    $("#input" + index).text(input)
    $("#input" + index).attr("data-plc-value", input);
  });

  $.each(data.Registers, function (index, register) {
    $("#register" + index).text(register)
  });

  $.each(data.Coils, function (index, coil) {
    $("#coil" + index).text(coil)
    $("#coil" + index).attr("data-plc-value", coil);
  });
};

// Handles a websocket message to update the hub LED status.
var handleHubLedChange = function (data) {
  $("#hubLedRed").text("rgb(" + data.Red.R + ", " + data.Red.G + ", " + data.Red.B + ")");
  $("#hubLedBlue").text("rgb(" + data.Blue.R + ", " + data.Blue.G + ", " + data.Blue.B + ")");
};

// Helper to set a boolean cell's text and data-plc-value attribute.
var setBoolCell = function (id, value) {
  $(id).text(value);
  $(id).attr("data-plc-value", value);
};

// Helper to set a configured/connected status pair.
var setNodeStatus = function (prefix, configured, connected) {
  $(prefix + "Configured").text(configured ? "Yes" : "No");
  $(prefix + "Connected").text(connected ? "Connected" : (configured ? "Disconnected" : "Not Configured"));
  $(prefix + "Connected").attr("data-plc-value", connected);
};

// Helper to populate stack light output cells for a node.
var setStackLightOutputs = function (prefix, node) {
  setBoolCell(prefix + "StackRed", node.StackRed);
  setBoolCell(prefix + "StackBlue", node.StackBlue);
  setBoolCell(prefix + "StackOrange", node.StackOrange);
  setBoolCell(prefix + "StackGreen", node.StackGreen);
  setBoolCell(prefix + "StackBuzzer", node.StackBuzzer);
  $(prefix + "MatchState").text(node.MatchState);
};

// Handles a websocket message to update all ESP32 node I/O status.
var handleArenaStatus = function (data) {
  // Score Table
  var st = data.ScoreTableIO;
  setNodeStatus("#espScoreTable", st.Configured, st.Connected);
  setBoolCell("#espScoreTableFieldEStop", st.FieldEStop);
  setStackLightOutputs("#espScoreTable", st);

  // Red Estops
  var re = data.RedEstopsIO;
  setNodeStatus("#espRedEstops", re.Configured, re.Connected);
  for (var i = 0; i < 3; i++) {
    setBoolCell("#espRedEstopsEStop" + i, re.EStops[i]);
    setBoolCell("#espRedEstopsAStop" + i, re.AStops[i]);
  }
  setStackLightOutputs("#espRedEstops", re);

  // Blue Estops
  var be = data.BlueEstopsIO;
  setNodeStatus("#espBlueEstops", be.Configured, be.Connected);
  for (var i = 0; i < 3; i++) {
    setBoolCell("#espBlueEstopsEStop" + i, be.EStops[i]);
    setBoolCell("#espBlueEstopsAStop" + i, be.AStops[i]);
  }
  setStackLightOutputs("#espBlueEstops", be);

  // Red Hub
  var rh = data.RedHubIO;
  setNodeStatus("#espRedHub", rh.Configured, rh.Connected);
  $("#espRedHubScore").text(rh.Score);
  $("#espRedHubHubState").text(rh.HubState);
  $("#espRedHubMotorDuty").text(rh.MotorDuty.toFixed(2));
  setBoolCell("#espRedHubRedLight", rh.RedLight);
  setBoolCell("#espRedHubBlueLight", rh.BlueLight);
  $("#espRedHubMatchState").text(rh.MatchState);

  // Blue Hub
  var bh = data.BlueHubIO;
  setNodeStatus("#espBlueHub", bh.Configured, bh.Connected);
  $("#espBlueHubScore").text(bh.Score);
  $("#espBlueHubHubState").text(bh.HubState);
  $("#espBlueHubMotorDuty").text(bh.MotorDuty.toFixed(2));
  setBoolCell("#espBlueHubRedLight", bh.RedLight);
  setBoolCell("#espBlueHubBlueLight", bh.BlueLight);
  $("#espBlueHubMatchState").text(bh.MatchState);
};

$(function () {
  // Set up the websocket back to the server.
  websocket = new CheesyWebsocket("/setup/field_testing/websocket", {
    plcIoChange: function (event) {
      handlePlcIoChange(event.data);
    },
    hubLed: function (event) {
      handleHubLedChange(event.data);
    },
    arenaStatus: function (event) {
      handleArenaStatus(event.data);
    }
  });
});
