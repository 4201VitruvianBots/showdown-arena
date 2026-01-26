#ifndef HW_SCORING_H
#define HW_SCORING_H

/**
 * @brief Initialize the hardware scoring module.
 */
void hw_scoring_init(void);

/**
 * @brief Get the HW state of the break beam sensor.
 * @return true if the break beam sensor is triggered, false otherwise.
 */
bool hw_scoring_getBreakBeamSensorState(void);

#endif // HW_SCORING_H