#include "app/app_scoring.hpp"
#include "hw/hw_scoring.hpp"
#include "io/io_comms.h"
#include <stdbool.h>
#include <string.h>

typedef struct {
  bool current_break_beam_sensor_state;
  bool last_break_beam_sensor_state;
  bool rising_edge_detected;
} app_scoring_channelData_t;
typedef struct {
  uint32_t score;
  RequestedHubState_E hub_state;
  app_scoring_channelData_t channelData[HW_SCORING_BREAK_BEAM_SENSOR_COUNT];
} app_scoring_data_t;

static app_scoring_data_t app_scoring_data;

static void app_scoring_captureAndProcessInputs(void) {
  app_scoring_data.hub_state = io_comms_getRequestedHubState();
  for (uint8_t i = 0; i < HW_SCORING_BREAK_BEAM_SENSOR_COUNT; i++) {
    app_scoring_channelData_t *channel = &app_scoring_data.channelData[i];
    channel->last_break_beam_sensor_state =
        channel->current_break_beam_sensor_state;
    channel->current_break_beam_sensor_state =
        hw_scoring_getBreakBeamSensorState((hw_scoring_breakBeamSensors_E)i);

    channel->rising_edge_detected =
        (channel->last_break_beam_sensor_state == false) &&
        (channel->current_break_beam_sensor_state == true);
  }
}

void app_scoring_reset(void) {
  memset(&app_scoring_data, 0, sizeof(app_scoring_data_t));
}

void app_scoring_run(void) {
  app_scoring_captureAndProcessInputs();

  switch (app_scoring_data.hub_state) {
  case REQUESTED_HUB_STATE_DEBUG_SCORING_TEST:
  case REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_ACTIVE: {
    for (uint8_t i = 0; i < HW_SCORING_BREAK_BEAM_SENSOR_COUNT; i++) {
      if (app_scoring_data.channelData[i].rising_edge_detected) {
        app_scoring_data.score++;
      }
    }
  } break;

  case REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_INACTIVE:
  default:
    // Scoring is inactive, but we don't want to reset.
    break;

  case REQUESTED_HUB_STATE_DEBUG_MOTOR_SPINUP:
  case REQUESTED_HUB_STATE_DISABLED: {
    // Reset score when disabled
    app_scoring_data.score = 0;
  } break;
  }
}

uint32_t app_scoring_getScore(void) { return app_scoring_data.score; }