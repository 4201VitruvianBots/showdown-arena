#ifndef APP_LED_STRIP_H
#define APP_LED_STRIP_H

/**
 * @brief Run the LED strip application logic.
 *
 * Reads the redLight/blueLight state from the network commands and sets
 * the LED strip color accordingly:
 *   - redLight only  → red
 *   - blueLight only → blue
 *   - both           → purple (indicates some overlap/error state)
 *   - neither        → off
 */
void app_led_strip_run(void);

#endif // APP_LED_STRIP_H
