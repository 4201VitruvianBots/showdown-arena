#include "hw/hw_led_strip.hpp"
#include "pinmaps.hpp"
#include <FastLED.h>

static CRGB leds[HW_LED_STRIP_NUM_LEDS];

void hw_led_strip_init(void) {
    FastLED.addLeds<WS2812B, HW_LED_STRIP_PIN, GRB>(leds, HW_LED_STRIP_NUM_LEDS);
    FastLED.setBrightness(255);
    hw_led_strip_off();
    hw_led_strip_show();
}

void hw_led_strip_setAll(uint8_t r, uint8_t g, uint8_t b) {
    fill_solid(leds, HW_LED_STRIP_NUM_LEDS, CRGB(r, g, b));
}

void hw_led_strip_off(void) {
    fill_solid(leds, HW_LED_STRIP_NUM_LEDS, CRGB::Black);
}

void hw_led_strip_show(void) {
    FastLED.show();
}
