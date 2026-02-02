#include <unity.h>
#include "app/app_motor.hpp"
#include "hw/hw_motor.hpp"
#include "hw/hw_scoring.hpp"
#include "io/io_comms.h"

// Mock implementations not used by this test, because no way to tell PIO not to link them.
static bool hw_scoring_break_beam_sensor_states[HW_SCORING_BREAK_BEAM_SENSOR_COUNT] = {false};
bool hw_scoring_getBreakBeamSensorState(hw_scoring_breakBeamSensors_E sensor) {
    if (sensor < HW_SCORING_BREAK_BEAM_SENSOR_COUNT) {
        return hw_scoring_break_beam_sensor_states[sensor];
    }
    return false; // Default to false for invalid sensor
}

// Mock implementations used by this test
static RequestedHubState_E mock_requested_hub_state = REQUESTED_HUB_STATE_DISABLED;
RequestedHubState_E io_comms_getRequestedHubState(void)
{
    return mock_requested_hub_state;
}

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


void setUp(void) {
    app_motor_reset();
}

void tearDown(void) {
    // clean stuff up here
}

void test_app_motor_init(void) {
    for (uint8_t motor = HW_MOTOR_INSTALLED_MOTOR_LEFT; motor < HW_MOTOR_INSTALLED_MOTOR_COUNT; motor++) {
        const float duty = app_motor_getCommandedDuty((hw_motor_configInstalledMotors_E)motor);
        TEST_ASSERT_EQUAL_FLOAT(0.0f, duty);
    }
}

void test_app_motor_duty_cycle_commanding(void)
{
    mock_requested_hub_state = REQUESTED_HUB_STATE_DEBUG_MOTOR_SPINUP;
    mock_motor_duty_cmds[HW_MOTOR_INSTALLED_MOTOR_LEFT] = 0.5f;
    mock_motor_duty_cmds[HW_MOTOR_INSTALLED_MOTOR_RIGHT] = -0.5f;
    app_motor_run();
    const float left_duty = app_motor_getCommandedDuty(HW_MOTOR_INSTALLED_MOTOR_LEFT);
    const float right_duty = app_motor_getCommandedDuty(HW_MOTOR_INSTALLED_MOTOR_RIGHT);
    TEST_ASSERT_EQUAL_FLOAT(0.5f, left_duty);
    TEST_ASSERT_EQUAL_FLOAT(-0.5f, right_duty);
    TEST_ASSERT_EQUAL_FLOAT(0.5f, hw_commanded_duties[HW_MOTOR_INSTALLED_MOTOR_LEFT]);
    TEST_ASSERT_EQUAL_FLOAT(-0.5f, hw_commanded_duties[HW_MOTOR_INSTALLED_MOTOR_RIGHT]);

    mock_requested_hub_state = REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_ACTIVE;
    mock_motor_duty_cmds[HW_MOTOR_INSTALLED_MOTOR_LEFT] = 0.8f;
    mock_motor_duty_cmds[HW_MOTOR_INSTALLED_MOTOR_RIGHT] = -0.8f;
    app_motor_run();
    float left_duty_active = app_motor_getCommandedDuty(HW_MOTOR_INSTALLED_MOTOR_LEFT);
    float right_duty_active = app_motor_getCommandedDuty(HW_MOTOR_INSTALLED_MOTOR_RIGHT);
    TEST_ASSERT_EQUAL_FLOAT(0.8f, left_duty_active);
    TEST_ASSERT_EQUAL_FLOAT(-0.8f, right_duty_active);
    TEST_ASSERT_EQUAL_FLOAT(0.8f, hw_commanded_duties[HW_MOTOR_INSTALLED_MOTOR_LEFT]);
    TEST_ASSERT_EQUAL_FLOAT(-0.8f, hw_commanded_duties[HW_MOTOR_INSTALLED_MOTOR_RIGHT]);

    mock_requested_hub_state = REQUESTED_HUB_STATE_MATCH_PLAY_SCORING_INACTIVE;
    app_motor_run();
    left_duty_active = app_motor_getCommandedDuty(HW_MOTOR_INSTALLED_MOTOR_LEFT);
    right_duty_active = app_motor_getCommandedDuty(HW_MOTOR_INSTALLED_MOTOR_RIGHT);
    TEST_ASSERT_EQUAL_FLOAT(0.8f, left_duty_active);
    TEST_ASSERT_EQUAL_FLOAT(-0.8f, right_duty_active);
    TEST_ASSERT_EQUAL_FLOAT(0.8f, hw_commanded_duties[HW_MOTOR_INSTALLED_MOTOR_LEFT]);
    TEST_ASSERT_EQUAL_FLOAT(-0.8f, hw_commanded_duties[HW_MOTOR_INSTALLED_MOTOR_RIGHT]);

    mock_requested_hub_state = REQUESTED_HUB_STATE_DISABLED;
    app_motor_run();      
    const float left_duty_disabled = app_motor_getCommandedDuty(HW_MOTOR_INSTALLED_MOTOR_LEFT);
    const float right_duty_disabled = app_motor_getCommandedDuty(HW_MOTOR_INSTALLED_MOTOR_RIGHT);
    TEST_ASSERT_EQUAL_FLOAT(0.0f, left_duty_disabled);
    TEST_ASSERT_EQUAL_FLOAT(0.0f, right_duty_disabled);
    TEST_ASSERT_EQUAL_FLOAT(0.0f, hw_commanded_duties[HW_MOTOR_INSTALLED_MOTOR_LEFT]);
    TEST_ASSERT_EQUAL_FLOAT(0.0f, hw_commanded_duties[HW_MOTOR_INSTALLED_MOTOR_RIGHT]);
}

int main() {
    UNITY_BEGIN();
    RUN_TEST(test_app_motor_init);
    RUN_TEST(test_app_motor_duty_cycle_commanding);
    UNITY_END();
}