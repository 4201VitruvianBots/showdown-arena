#ifndef IO_COMMS_H
#define IO_COMMS_H
#include "hw/hw_motor.hpp"

#include <stdint.h>

typedef enum {
  REQUESTED_HUB_STATE_DISABLED,
  REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_ACTIVE,
  REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_INACTIVE,
  REQUESTED_HUB_STATE_DEBUG_MOTOR_SPINUP,
  REQUESTED_HUB_STATE_DEBUG_SCORING_TEST,
} RequestedHubState_E;

/**
 * @brief Get the requested hub state.
 * @return RequestedHubState_E The requested hub state.
 */
RequestedHubState_E io_comms_getRequestedHubState(void);

/**
 * @brief Get the motor duty command.
 * @param motor The motor identifier.
 * @return float The motor duty command (-1.0 -> 1.0).
 */
float io_comms_getConfigMotorDutyCmd(hw_motor_configInstalledMotors_E motor);

#endif // IO_COMMS_H