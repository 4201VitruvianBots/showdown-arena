#include <unity.h>
#include "app/app_scoring.hpp"
#include "io/io_comms.h"
#include "hw/hw_scoring.hpp"

// Mock implementations not used by this test, because no way to tell PIO not to link them.
static float mock_motor_duty_cmds[HW_MOTOR_INSTALLED_MOTOR_COUNT] = {0.0f};
float io_comms_getConfigMotorDutyCmd(hw_motor_configInstalledMotors_E motor)
{
    return mock_motor_duty_cmds[motor];
}

static float hw_commanded_duties[HW_MOTOR_INSTALLED_MOTOR_COUNT] = {0.0f};
void hw_motor_setDutyCycle(hw_motor_configInstalledMotors_E motor, float duty_cycle)
{
    hw_commanded_duties[motor] = duty_cycle;
}

// Mock implementations
static bool hw_scoring_break_beam_sensor_states[HW_SCORING_BREAK_BEAM_SENSOR_COUNT] = {false};
bool hw_scoring_getBreakBeamSensorState(hw_scoring_breakBeamSensors_E sensor) {
    if (sensor < HW_SCORING_BREAK_BEAM_SENSOR_COUNT) {
        return hw_scoring_break_beam_sensor_states[sensor];
    }
    return false; // Default to false for invalid sensor
}

static RequestedHubState_E mock_requested_hub_state = REQUESTED_HUB_STATE_DISABLED;
RequestedHubState_E io_comms_getRequestedHubState(void)
{
    return mock_requested_hub_state;
}


void setUp(void) {
    for (uint8_t i = 0; i < HW_SCORING_BREAK_BEAM_SENSOR_COUNT; i++) {
        hw_scoring_break_beam_sensor_states[i] = false;
    }
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

void test_app_scoring_single_sensor_edge_detection(void) {
    // Test rising edge detection for a single sensor
    int sensor = 0;
    hw_scoring_break_beam_sensor_states[sensor] = true;
    app_scoring_run();
    uint32_t score = app_scoring_getScore();
    TEST_ASSERT_EQUAL_UINT32(0, score); // still disabled
    
    mock_requested_hub_state = REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_ACTIVE;
    app_scoring_run();
    score = app_scoring_getScore();
    TEST_ASSERT_EQUAL_UINT32(0, score); // no rising edge yet
    
    hw_scoring_break_beam_sensor_states[sensor] = false;
    app_scoring_run();
    TEST_ASSERT_EQUAL_UINT32(0, app_scoring_getScore());

    hw_scoring_break_beam_sensor_states[sensor] = true;
    app_scoring_run();
    score = app_scoring_getScore();
    TEST_ASSERT_EQUAL_UINT32(1, score); // rising edge

    hw_scoring_break_beam_sensor_states[sensor] = false;
    app_scoring_run();
    TEST_ASSERT_EQUAL_UINT32(1, app_scoring_getScore());
    
    hw_scoring_break_beam_sensor_states[sensor] = true;
    app_scoring_run();
    score = app_scoring_getScore();
    TEST_ASSERT_EQUAL_UINT32(2, score); // another rising edge
}

void test_app_scoring_multi_sensor_edge_detection(void) {
    // Test rising edge detection for all sensors and sum
    mock_requested_hub_state = REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_ACTIVE;
    for (uint8_t i = 0; i < HW_SCORING_BREAK_BEAM_SENSOR_COUNT; i++) {
        hw_scoring_break_beam_sensor_states[i] = false;
    }
    app_scoring_run();
    TEST_ASSERT_EQUAL_UINT32(0, app_scoring_getScore());

    // Simulate rising edge on all sensors one by one
    for (uint8_t i = 0; i < HW_SCORING_BREAK_BEAM_SENSOR_COUNT; i++) {
        hw_scoring_break_beam_sensor_states[i] = true;
        app_scoring_run();
        TEST_ASSERT_EQUAL_UINT32(i+1, app_scoring_getScore());
        hw_scoring_break_beam_sensor_states[i] = false;
        app_scoring_run();
        TEST_ASSERT_EQUAL_UINT32(i+1, app_scoring_getScore());
    }

    // Simulate simultaneous rising edges
    for (uint8_t i = 0; i < HW_SCORING_BREAK_BEAM_SENSOR_COUNT; i++) {
        hw_scoring_break_beam_sensor_states[i] = true;
    }
    app_scoring_run();
    TEST_ASSERT_EQUAL_UINT32(HW_SCORING_BREAK_BEAM_SENSOR_COUNT*2, app_scoring_getScore());
}

void test_app_scoring_inactive_and_reset(void) {
    // Test that scoring does not increment when inactive and resets when disabled
    mock_requested_hub_state = REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_ACTIVE;
    for (uint8_t i = 0; i < HW_SCORING_BREAK_BEAM_SENSOR_COUNT; i++) {
        hw_scoring_break_beam_sensor_states[i] = false;
    }
    app_scoring_run();
    // Simulate a rising edge on sensor 0
    hw_scoring_break_beam_sensor_states[0] = true;
    app_scoring_run();
    TEST_ASSERT_EQUAL_UINT32(1, app_scoring_getScore());
    // Set inactive
    mock_requested_hub_state = REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_INACTIVE;
    hw_scoring_break_beam_sensor_states[0] = false;
    app_scoring_run();
    hw_scoring_break_beam_sensor_states[0] = true;
    app_scoring_run();
    TEST_ASSERT_EQUAL_UINT32(1, app_scoring_getScore()); // no increment
    // Set disabled
    mock_requested_hub_state = REQUESTED_HUB_STATE_DISABLED;
    app_scoring_run();
    TEST_ASSERT_EQUAL_UINT32(0, app_scoring_getScore()); // reset
}

int main() {
    UNITY_BEGIN();
    RUN_TEST(test_app_scoring_init);
    RUN_TEST(test_app_scoring_single_sensor_edge_detection);
    RUN_TEST(test_app_scoring_multi_sensor_edge_detection);
    RUN_TEST(test_app_scoring_inactive_and_reset);
    UNITY_END();
}