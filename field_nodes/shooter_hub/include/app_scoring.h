#ifndef APP_SCORING_H
#define APP_SCORING_H

#include "io_comms.h"

/**
 * @brief Initialize the scoring application.
 */
void app_scoring_reset(void);

/**
 * @brief Run the scoring application logic, to be called periodically.
 */
void app_scoring_run(void);

/**
 * @brief Get the current score.
 * @return uint32_t The current score.
 */
uint32_t app_scoring_getScore(void);

#endif // APP_SCORING_H