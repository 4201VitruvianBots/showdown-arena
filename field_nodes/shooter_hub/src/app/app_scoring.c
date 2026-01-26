#include "app_scoring.h"
#include <string.h>
#include "io_comms.h"
#include "hw_scoring.h"

typedef struct {
    uint32_t score;
    bool last_break_beam_sensor_state;
} app_scoring_data_t;

static app_scoring_data_t app_scoring_data;

void app_scoring_reset(void) {
    memset(&app_scoring_data, 0, sizeof(app_scoring_data_t));
}

void app_scoring_run(void)
{
    const RequestedHubState_E hub_state = io_comms_getRequestedHubState();
    const bool current_break_beam_sensor_state = hw_scoring_getBreakBeamSensorState();

    switch (hub_state) {
        case REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_ACTIVE:
        {
            // Edge detection: rising edge
            if ((app_scoring_data.last_break_beam_sensor_state == false) && 
                (current_break_beam_sensor_state == true)) {
                app_scoring_data.score++;
            }
        }
            break;

        case REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_INACTIVE:
        default:
            // Scoring is inactive, but we don't want to reset.
            break;
        
        case REQUESTED_HUB_STATE_DEBUG_MOTOR_SPINUP:
        case REQUESTED_HUB_STATE_DISABLED:
        {
            // Reset score when disabled
            app_scoring_data.score = 0;
        }
            break;
        
    }

    app_scoring_data.last_break_beam_sensor_state = current_break_beam_sensor_state;
}

uint32_t app_scoring_getScore(void) {
    return app_scoring_data.score;
}