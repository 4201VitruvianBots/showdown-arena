#ifndef HW_MOTOR_H
#define HW_MOTOR_H

typedef enum {
  HW_MOTOR_INSTALLED_MOTOR_LEFT = 0,
  HW_MOTOR_INSTALLED_MOTOR_RIGHT,
  HW_MOTOR_INSTALLED_MOTOR_COUNT
} hw_motor_configInstalledMotors_E;

/** @brief Initialize the motor hardware */
void hw_motor_init(void);

/**
 * @brief Set the motor duty cycle.
 * @param motor The motor identifier.
 * @param duty_cycle The duty cycle (-1.0 to 1.0).
 */
void hw_motor_setDutyCycle(hw_motor_configInstalledMotors_E motor,
                           float duty_cycle);

#endif // HW_MOTOR_H