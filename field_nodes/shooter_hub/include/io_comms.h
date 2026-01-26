#ifndef IO_COMMS_H
#define IO_COMMS_H

#include <stdint.h>

typedef enum {
    REQUESTED_HUB_STATE_DISABLED,
    REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_ACTIVE,
    REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_INACTIVE,
    REQUESTED_HUB_STATE_DEBUG_MOTOR_SPINUP,
} RequestedHubState_E;

/**
 * @brief Get the requested hub state.
 * @return RequestedHubState_E The requested hub state.
 */
RequestedHubState_E io_comms_getRequestedHubState(void);

/**
 * @brief Get the motor duty command.
 * @return float The motor duty command (-1.0 -> 1.0).
 */
float io_comms_getConfigMotorDutyCmd(void);


#endif // IO_COMMS_H