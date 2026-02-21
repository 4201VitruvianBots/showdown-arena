#include "app/app_display.hpp"
#include "hw/hw_display.hpp"

static uint32_t lastDisplayedScore = UINT32_MAX; // force initial draw

void app_display_init(void)
{
    MatrixPanel_I2S_DMA *panel = hw_display_getPanel();
    if (!panel) return;

    panel->fillScreen(panel->color565(0, 0, 0));
}

void app_display_updateScore(uint32_t score)
{
    // Only redraw when the score actually changes
    if (score == lastDisplayedScore) return;
    lastDisplayedScore = score;

    MatrixPanel_I2S_DMA *panel = hw_display_getPanel();
    if (!panel) return;

    // Clear screen
    panel->fillScreen(panel->color565(0, 0, 0));

    // Convert score to string
    char scoreStr[12];
    snprintf(scoreStr, sizeof(scoreStr), "%lu", (unsigned long)score);

    // Use text size 4 (~24x32 px per character) for readability on 64x64
    panel->setTextSize(4);
    panel->setTextWrap(false);
    panel->setTextColor(panel->color565(255, 255, 255)); // white text

    // Centre the text on the 64x64 panel
    // Each char at size 4 is 6*4=24 px wide, 8*4=32 px tall
    int textWidth = strlen(scoreStr) * 24;
    int x = (64 - textWidth) / 2;
    if (x < 0) x = 0;
    int y = (64 - 32) / 2; // vertically centred

    panel->setCursor(x, y);
    panel->print(scoreStr);
}
