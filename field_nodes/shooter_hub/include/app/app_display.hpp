#ifndef APP_DISPLAY_H
#define APP_DISPLAY_H

#include <cstdint>

/**
 * @brief Initialize the display application layer (clears screen).
 */
void app_display_init(void);

/**
 * @brief Update the score shown on the LED matrix display.
 * @param score The current score to render.
 */
void app_display_updateScore(uint32_t score);

#endif // APP_DISPLAY_H
