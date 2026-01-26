#include <unity.h>
#include "app_scoring.h"
#include "io_comms.h"
#include "hw_scoring.h"

// Mock implementations
static bool hw_scoring_break_beam_sensor_state = false;
void hw_scoring_init(void) {
    hw_scoring_break_beam_sensor_state = false;
}
bool hw_scoring_getBreakBeamSensorState(void) {
    return hw_scoring_break_beam_sensor_state;
}

static RequestedHubState_E mock_requested_hub_state = REQUESTED_HUB_STATE_DISABLED;
RequestedHubState_E io_comms_getRequestedHubState(void)
{
    return mock_requested_hub_state;
}

void setUp(void) {
    hw_scoring_break_beam_sensor_state = false;
    mock_requested_hub_state = REQUESTED_HUB_STATE_DISABLED;
    app_scoring_reset();
}

void tearDown(void) {
    // clean stuff up here
}

void test_app_scoring_init(void) {
    const uint32_t score = app_scoring_getScore();
    TEST_ASSERT_EQUAL_UINT32(0, score);
}

void test_app_scoring_edge_detection(void) {
    hw_scoring_break_beam_sensor_state = true;
    app_scoring_run();
    uint32_t score = app_scoring_getScore();
    TEST_ASSERT_EQUAL_UINT32(0, score); // still disabled
    
    mock_requested_hub_state = REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_ACTIVE;
    app_scoring_run();
    score = app_scoring_getScore();
    TEST_ASSERT_EQUAL_UINT32(0, score); // score should not increment because we haven't simulated a rising edge
    
    hw_scoring_break_beam_sensor_state = false;
    app_scoring_run();
    TEST_ASSERT_EQUAL_UINT32(0, app_scoring_getScore());

    hw_scoring_break_beam_sensor_state = true;
    app_scoring_run(); 
    score = app_scoring_getScore();
    TEST_ASSERT_EQUAL_UINT32(1, score); // score should increment on rising edge

    hw_scoring_break_beam_sensor_state = false;
    app_scoring_run();
    TEST_ASSERT_EQUAL_UINT32(1, app_scoring_getScore());
    
    hw_scoring_break_beam_sensor_state = true;
    app_scoring_run();
    score = app_scoring_getScore();
    TEST_ASSERT_EQUAL_UINT32(2, score); // score should increment on rising edge

    mock_requested_hub_state = REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_INACTIVE;
    app_scoring_run();
    score = app_scoring_getScore();
    TEST_ASSERT_EQUAL_UINT32(2, score); // score should not increment when inactive

    hw_scoring_break_beam_sensor_state = false;
    app_scoring_run();
    TEST_ASSERT_EQUAL_UINT32(2, app_scoring_getScore());

    hw_scoring_break_beam_sensor_state = true;
    app_scoring_run();
    score = app_scoring_getScore();
    TEST_ASSERT_EQUAL_UINT32(2, score); // score should not increment when inactive

    mock_requested_hub_state = REQUESTED_HUB_STATE_DISABLED;
    app_scoring_run();
    score = app_scoring_getScore();
    TEST_ASSERT_EQUAL_UINT32(0, score); // score should reset when disabled
}

int main() {
    UNITY_BEGIN();
    RUN_TEST(test_app_scoring_init);
    RUN_TEST(test_app_scoring_edge_detection);
    UNITY_END();
}