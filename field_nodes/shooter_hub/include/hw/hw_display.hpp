#ifndef HW_DISPLAY_H
#define HW_DISPLAY_H

#include <ESP32-HUB75-MatrixPanel-I2S-DMA.h>

/**
 * @brief Initialize the HUB75 LED matrix display hardware.
 */
void hw_display_init(void);

/**
 * @brief Get the display panel pointer for drawing operations.
 * @return MatrixPanel_I2S_DMA* Pointer to the initialized display panel.
 */
MatrixPanel_I2S_DMA *hw_display_getPanel(void);

#endif // HW_DISPLAY_H
