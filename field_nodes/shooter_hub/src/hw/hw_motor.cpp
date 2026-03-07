#include "hw/hw_motor.hpp"
#include "pinmaps.hpp"
#include <Arduino.h>

// PWM configuration for servo-style ESC control
#define HW_MOTOR_PWM_FREQ_HZ    50     // Standard servo/ESC frequency
#define HW_MOTOR_PWM_RESOLUTION 16     // 16-bit resolution (0-65535)

static const uint8_t hw_motor_pwm_pins[HW_MOTOR_INSTALLED_MOTOR_COUNT] = {
    HW_MOTOR_LEFT_PWM_PIN,
    HW_MOTOR_RIGHT_PWM_PIN
};

void hw_motor_init(void) {
    // Initialize motor PWM pins using direct LEDC API
    for (uint8_t i = 0; i < HW_MOTOR_INSTALLED_MOTOR_COUNT; i++) {
        ledcAttach(hw_motor_pwm_pins[i], HW_MOTOR_PWM_FREQ_HZ, HW_MOTOR_PWM_RESOLUTION);
        hw_motor_setDutyCycle((hw_motor_configInstalledMotors_E)i, 0.0f);
    }
}

void hw_motor_setDutyCycle(hw_motor_configInstalledMotors_E motor, float duty_cycle) {
    if (motor < HW_MOTOR_INSTALLED_MOTOR_COUNT) {
        const float clamped_duty = constrain(duty_cycle, -1.0f, 1.0f);
        // ESC: neutral 1500us, range +/- 500us, at 50Hz (20ms period)
        const float pulse_width_us = 1500.0f + (clamped_duty * 500.0f);
        const float period_us = 1000000.0f / HW_MOTOR_PWM_FREQ_HZ;
        const uint32_t max_duty = (1 << HW_MOTOR_PWM_RESOLUTION) - 1;
        const uint32_t duty = (uint32_t)((pulse_width_us / period_us) * max_duty);
        ledcWrite(hw_motor_pwm_pins[motor], duty);
    }
}