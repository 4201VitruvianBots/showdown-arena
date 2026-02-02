#ifndef APP_MOTOR_H
#define APP_MOTOR_H
#include "hw/hw_motor.hpp"

/**
 * @file app_motor.hpp
 * @brief Application layer for motor control.
 */
void app_motor_reset(void);

/**
 * @brief Run the motor control application logic.
 */
void app_motor_run(void);

/**
 * @brief Get the commanded duty cycle for a specified motor.
 * @param motor The motor identifier.
 * @return float The commanded duty cycle (-1.0 to 1.0).
 */
float app_motor_getCommandedDuty(hw_motor_configInstalledMotors_E motor);

#endif // APP_MOTOR_H