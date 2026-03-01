#include "io/io_comms.h"
#include "comms/udp_comms.hpp"

// Map the hub state string received from the Go server to the enum
static RequestedHubState_E mapHubState(const String &hubState) {
  if (hubState == "SCORING_ACTIVE") {
    return REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_ACTIVE;
  } else if (hubState == "SCORING_INACTIVE") {
    return REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_INACTIVE;
  } else if (hubState == "DEBUG_MOTOR_SPINUP") {
    return REQUESTED_HUB_STATE_DEBUG_MOTOR_SPINUP;
  } else if (hubState == "DEBUG_SCORING_TEST") {
    return REQUESTED_HUB_STATE_DEBUG_SCORING_TEST;
  } else {
    return REQUESTED_HUB_STATE_DISABLED;
  }
}

RequestedHubState_E io_comms_getRequestedHubState(void) {
  // If the server is connected, use the hub state from UDP
  if (udp_comms_isServerConnected()) {
    return mapHubState(udp_comms_getHubState());
  }
  // Default to debug scoring test when no server connection
  return REQUESTED_HUB_STATE_DEBUG_SCORING_TEST;
}

float io_comms_getConfigMotorDutyCmd(hw_motor_configInstalledMotors_E motor) {
  if (udp_comms_isServerConnected()) {
    return udp_comms_getMotorDuty();
  }
  return 0.0f;
}