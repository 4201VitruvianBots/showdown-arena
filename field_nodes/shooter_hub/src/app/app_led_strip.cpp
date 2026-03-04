#include "app/app_led_strip.hpp"
#include "hw/hw_led_strip.hpp"
#include "comms/udp_comms.hpp"

void app_led_strip_run(void) {
    bool redLight = udp_comms_getRedLight();
    bool blueLight = udp_comms_getBlueLight();

    if (redLight && blueLight) {
        // Both active — show purple
        hw_led_strip_setAll(128, 0, 128);
    } else if (redLight) {
        hw_led_strip_setAll(255, 0, 0);
    } else if (blueLight) {
        hw_led_strip_setAll(0, 0, 255);
    } else {
        hw_led_strip_off();
    }

    hw_led_strip_show();
}
