#ifndef HW_LED_STRIP_H
#define HW_LED_STRIP_H

#include <cstdint>

/**
 * @brief Initialize the FastLED strip hardware.
 */
void hw_led_strip_init(void);

/**
 * @brief Set all LEDs on the strip to the specified RGB color.
 * @param r Red component (0-255)
 * @param g Green component (0-255)
 * @param b Blue component (0-255)
 */
void hw_led_strip_setAll(uint8_t r, uint8_t g, uint8_t b);

/**
 * @brief Turn off all LEDs on the strip.
 */
void hw_led_strip_off(void);

/**
 * @brief Call FastLED.show() to push updates to the strip.
 */
void hw_led_strip_show(void);

#endif // HW_LED_STRIP_H
