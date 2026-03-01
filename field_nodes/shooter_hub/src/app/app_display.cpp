#include "app/app_display.hpp"
#include "hw/hw_display.hpp"

static uint32_t lastDisplayedScore = UINT32_MAX; // force initial draw

void app_display_init(void)
{
    MatrixPanel_I2S_DMA *panel = hw_display_getPanel();
    if (!panel) return;

    panel->fillScreen(panel->color565(0, 0, 0));
}

void draw_centered_string(const String &buf, int x, int y)
{

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

    // Use text size 3 (12x16 px per character) for readability on 64x64
    panel->setTextSize(3);
    panel->setTextWrap(false);
    panel->setTextColor(panel->color565(255, 255, 255)); // white text

    // Centre the text on the 64x64 panel

    int16_t x1, y1;
    uint16_t w, h;
    panel->getTextBounds(scoreStr, 32, 32, &x1, &y1, &w, &h); //calc width of new string
    panel->setCursor(32 - w / 2, 32 - h / 2);
    panel->print(scoreStr);
}
