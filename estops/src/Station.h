#pragma once

#include "common.h"
#include <Arduino.h>
#include "Button.h"

class Station
{
public:
    Station(AllianceColor _color, StationId _station_id, int astop_pin, int estop_pin) : color(_color),
                                                                                         station_id(_station_id),
                                                                                         estop(ButtonType::E_STOP, estop_pin),
                                                                                         astop(ButtonType::A_STOP, astop_pin)
    {
    }

    void init()
    {
        estop.init();
        astop.init();
    }

    /*
     * Get state of this station.
     */
    const StationState get_state()
    {
        return StationState{
            .estopped = estop.get_state(),
            .astopped = astop.get_state(),
        };
    }

private:
    Button estop, astop;

    /*
     * Color of this station.
     */
    AllianceColor color;

    /*
     * ID of this station.
     */
    StationId station_id;

    /*
     * Autonomus stop pin.
     */
    int astop_pin;

    /*
     * Emergency stop pin.
     */
    int estop_pin;

    /*
     * Autonomus stop timer to keep track of how long to assert button press.
     */
    int astop_timer_start_time_millis = 0;

    /*
     * Emergency stop timer to keep track of how long to assert button press.
     */
    int estop_timer_start_time_millis = 0;

    StationState last_state{};
};