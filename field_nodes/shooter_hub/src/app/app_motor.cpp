#include "app/app_motor.hpp"
#include "io/io_comms.h"
#include <string.h>

typedef struct {
  float commanded_duty[HW_MOTOR_INSTALLED_MOTOR_COUNT];
} app_motor_data_t;

static app_motor_data_t app_motor_data;

void app_motor_reset(void) {
  for (int i = 0; i < HW_MOTOR_INSTALLED_MOTOR_COUNT; i++) {
    app_motor_data.commanded_duty[i] = 0.0f;
  }
}

void app_motor_run(void) {
  const RequestedHubState_E requested_state = io_comms_getRequestedHubState();
  switch (requested_state) {
  case REQUESTED_HUB_STATE_DEBUG_MOTOR_SPINUP:
  case REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_INACTIVE:
  case REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_ACTIVE:
    for (uint8_t motor = HW_MOTOR_INSTALLED_MOTOR_LEFT;
         motor < HW_MOTOR_INSTALLED_MOTOR_COUNT; motor++) {
      app_motor_data.commanded_duty[motor] = io_comms_getConfigMotorDutyCmd(
          (hw_motor_configInstalledMotors_E)motor);
    }
    break;

  case REQUESTED_HUB_STATE_DEBUG_SCORING_TEST:
  case REQUESTED_HUB_STATE_DISABLED:
  default:
    for (uint8_t motor = HW_MOTOR_INSTALLED_MOTOR_LEFT;
         motor < HW_MOTOR_INSTALLED_MOTOR_COUNT; motor++) {
      app_motor_data.commanded_duty[motor] = 0.0f;
    }
    break;
  }

  for (uint8_t motor = HW_MOTOR_INSTALLED_MOTOR_LEFT;
       motor < HW_MOTOR_INSTALLED_MOTOR_COUNT; motor++) {
    hw_motor_setDutyCycle((hw_motor_configInstalledMotors_E)motor,
                          app_motor_data.commanded_duty[motor]);
  }
}

float app_motor_getCommandedDuty(hw_motor_configInstalledMotors_E motor) {
  if (motor < HW_MOTOR_INSTALLED_MOTOR_COUNT) {
    return app_motor_data.commanded_duty[motor];
  }
  return 0.0f; // Invalid motor, return 0.0f
}