#include "app/app_led_strip.hpp"
#include "hw/hw_led_strip.hpp"
#include "comms/udp_comms.hpp"
#include "pinmaps.hpp"
#include <Arduino.h>

// helper for blending
void blendPixel(int index, uint8_t br, uint8_t bg, uint8_t bb, uint8_t or_, uint8_t og, uint8_t ob, float mix) {
    uint8_t r = (uint8_t)(br * (1.0f - mix) + or_ * mix);
    uint8_t g = (uint8_t)(bg * (1.0f - mix) + og * mix);
    uint8_t b = (uint8_t)(bb * (1.0f - mix) + ob * mix);
    hw_led_strip_setPixel(index, r, g, b);
}

void app_led_strip_run(void) {
    String pattern = udp_comms_getLedPattern();

    if (pattern == "red") {
        hw_led_strip_setAll(255, 0, 0);
    } else if (pattern == "blue") {
        hw_led_strip_setAll(0, 0, 255);
    } else if (pattern == "purple") {
        hw_led_strip_setAll(128, 0, 128);
    } else if (pattern == "green") {
        hw_led_strip_setAll(0, 255, 0);
    } else if (pattern == "white") {
        hw_led_strip_setAll(255, 255, 255);
    } else if (pattern == "red_flash" || pattern == "blue_flash") {
        // 2Hz flash (250ms on, 250ms off)
        long timeInCycle = millis() % 1000;
        float multiplier = timeInCycle < 500 ? 0 : 1.0f;
        
        uint8_t r = pattern == "red_flash" ? (uint8_t)(255 * multiplier) : 0;
        uint8_t g = 0;
        uint8_t b = pattern == "blue_flash" ? (uint8_t)(255 * multiplier) : 0;

        hw_led_strip_setAll(r, g, b);
    } else if (pattern == "red_chase" || pattern == "blue_chase") {
        // 1 second continuous chase cycle
        long timeInCycle = millis() % 500;
        int chaseIndex = (int)((timeInCycle / 250.0f) * HW_LED_STRIP_NUM_LEDS);
        
        uint8_t br = pattern == "red_chase" ? 255 : 0;
        uint8_t bg = 0;
        uint8_t bb = pattern == "blue_chase" ? 255 : 0;

        // Render base color
        hw_led_strip_setAll(br, bg, bb);

        // Render a white overlay/gradient that trails 44% of the strip (analogous to the DMX 7-segment width)
        int width = HW_LED_STRIP_NUM_LEDS * 0.44f;
        for(int w = -width/2; w <= width/2; w++) {
            int targetIdx = (chaseIndex + w + HW_LED_STRIP_NUM_LEDS) % HW_LED_STRIP_NUM_LEDS;
            float mix = 1.0f - ((float)abs(w) / (width/2.0f));
            if (mix < 0.0f) mix = 0.0f;
            
            blendPixel(targetIdx, br, bg, bb, 255, 255, 255, mix);
        }
    } else {
        hw_led_strip_off();
    }

    hw_led_strip_show();
}
