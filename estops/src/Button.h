#pragma once
#include "common.h"
#include <Arduino.h>

class Button
{
public:
    Button(ButtonType _button_type, int _pin) : button_type(_button_type), pin(_pin) {}

    void init() {
        /*
         * Configure pin as inputs.
         */
        pinMode(pin, INPUT);
    }

    bool get_state()
    {
        bool pressed = !digitalRead(pin);
        int current_time = millis();

        bool state = false;

        bool latched = current_time < latch_timer_start_time_millis + latch_time_millis;
        if (pressed || latched)
        {
            state = true;

            if (!last_state)
            {
                latch_timer_start_time_millis = current_time;
            }
        }

        return state;
    }

private:
    ButtonType button_type;

    /*
     * Amount of time to latch E/A stops on when pressed.
     */
    int latch_time_millis = 1000;

    int pin;

    bool last_state = false;

    int latch_timer_start_time_millis = 0;
};