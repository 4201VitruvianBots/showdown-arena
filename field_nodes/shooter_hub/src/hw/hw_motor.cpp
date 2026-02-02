#include "hw/hw_motor.hpp"
#include "pinmaps.hpp"
#include <Arduino.h>
#include <ESP32Servo.h>

static uint8_t hw_motor_pwm_pins[HW_MOTOR_INSTALLED_MOTOR_COUNT] = {
    HW_MOTOR_LEFT_PWM_PIN,
    HW_MOTOR_RIGHT_PWM_PIN
};

static Servo motorServos[HW_MOTOR_INSTALLED_MOTOR_COUNT];

void hw_motor_init(void) {
    // Initialize the motor PWM pins
    for (uint8_t i = 0; i < HW_MOTOR_INSTALLED_MOTOR_COUNT; i++) {
        motorServos[i].attach(hw_motor_pwm_pins[i]);
        hw_motor_setDutyCycle((hw_motor_configInstalledMotors_E)i, 0.0f);
    }
}

void hw_motor_setDutyCycle(hw_motor_configInstalledMotors_E motor, float duty_cycle) {
    if (motor < HW_MOTOR_INSTALLED_MOTOR_COUNT) {
        const float clamped_duty = constrain(duty_cycle, -1.0f, 1.0f);
        const float pulse_width_us = 1500.0f + (clamped_duty * 500.0f); // Neutral at 1500us, range +/- 500us
        motorServos[motor].writeMicroseconds(static_cast<int>(pulse_width_us));
    }
}