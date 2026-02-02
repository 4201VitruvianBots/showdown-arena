#include "app/app_scoring.hpp"
#include "hw/hw_scoring.hpp"
#include "pinmaps.hpp"
#include <Arduino.h>

static uint8_t hw_scoring_break_beam_sensor_pins[HW_SCORING_BREAK_BEAM_SENSOR_COUNT] = {
    HW_SCORING_BREAK_BEAM_SENSOR_PIN_LEFTMOST,
    HW_SCORING_BREAK_BEAM_SENSOR_PIN_MIDDLE_LEFT,
    HW_SCORING_BREAK_BEAM_SENSOR_PIN_MIDDLE_RIGHT,
    HW_SCORING_BREAK_BEAM_SENSOR_PIN_RIGHTMOST
};

void hw_scoring_init(void) {
  // Initialize the break beam sensor pin
  for (uint8_t i = 0; i < HW_SCORING_BREAK_BEAM_SENSOR_COUNT; i++) {
    pinMode(hw_scoring_break_beam_sensor_pins[i], INPUT);
  }
}

bool hw_scoring_getBreakBeamSensorState(hw_scoring_breakBeamSensors_E sensor) {
  if (sensor < HW_SCORING_BREAK_BEAM_SENSOR_COUNT) {
    return digitalRead(hw_scoring_break_beam_sensor_pins[sensor]) == HIGH;
  }
  return false; // Default to false for invalid sensor
}