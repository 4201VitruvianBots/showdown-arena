#ifndef HW_SCORING_H
#define HW_SCORING_H
#include <stdbool.h>

typedef enum {
  HW_SCORING_BREAK_BEAM_SENSOR_LEFTMOST,
  HW_SCORING_BREAK_BEAM_SENSOR_MIDDLE_LEFT,
  HW_SCORING_BREAK_BEAM_SENSOR_MIDDLE_RIGHT,
  HW_SCORING_BREAK_BEAM_SENSOR_RIGHTMOST,
  HW_SCORING_BREAK_BEAM_SENSOR_COUNT
} hw_scoring_breakBeamSensors_E;

/**
 * @brief Initialize the hardware scoring module.
 */
void hw_scoring_init(void);

/**
 * @brief Get the HW state of the break beam sensor.
 * @param sensor The break beam sensor identifier.
 * @return true if the break beam sensor is triggered, false otherwise.
 */
bool hw_scoring_getBreakBeamSensorState(hw_scoring_breakBeamSensors_E sensor);

#endif // HW_SCORING_H