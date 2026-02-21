#include "hw/hw_display.hpp"
#include "pinmaps.hpp"

#define PANEL_RES_X 64
#define PANEL_RES_Y 64
#define PANEL_CHAIN 1

static MatrixPanel_I2S_DMA *dma_display = nullptr;

void hw_display_init(void)
{
    HUB75_I2S_CFG mxconfig(
        PANEL_RES_X,
        PANEL_RES_Y,
        PANEL_CHAIN
    );

    // Custom pin mapping from pinmaps.hpp
    mxconfig.gpio.r1  = HW_DISPLAY_R1_PIN;
    mxconfig.gpio.g1  = HW_DISPLAY_G1_PIN;
    mxconfig.gpio.b1  = HW_DISPLAY_B1_PIN;
    mxconfig.gpio.r2  = HW_DISPLAY_R2_PIN;
    mxconfig.gpio.g2  = HW_DISPLAY_G2_PIN;
    mxconfig.gpio.b2  = HW_DISPLAY_B2_PIN;
    mxconfig.gpio.a   = HW_DISPLAY_A_PIN;
    mxconfig.gpio.b   = HW_DISPLAY_B_PIN;
    mxconfig.gpio.c   = HW_DISPLAY_C_PIN;
    mxconfig.gpio.d   = HW_DISPLAY_D_PIN;
    mxconfig.gpio.e   = HW_DISPLAY_E_PIN;  // Required for 64-row panels
    mxconfig.gpio.lat = HW_DISPLAY_LAT_PIN;
    mxconfig.gpio.oe  = HW_DISPLAY_OE_PIN;
    mxconfig.gpio.clk = HW_DISPLAY_CLK_PIN;

    dma_display = new MatrixPanel_I2S_DMA(mxconfig);
    dma_display->begin();
    dma_display->setBrightness8(90);
    dma_display->clearScreen();
}

MatrixPanel_I2S_DMA *hw_display_getPanel(void)
{
    return dma_display;
}
