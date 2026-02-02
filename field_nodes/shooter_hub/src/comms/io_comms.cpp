#include "io/io_comms.h"

RequestedHubState_E io_comms_getRequestedHubState(void) {
  return REQUESTED_HUB_STATE_DEBUG_SCORING_TEST;
}

float io_comms_getConfigMotorDutyCmd(hw_motor_configInstalledMotors_E motor) {
  return 0.0f;
}