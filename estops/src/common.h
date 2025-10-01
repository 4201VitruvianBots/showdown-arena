#pragma once

struct ButtonStates
{
    bool fieldEStop;
    bool station1EStop;
    bool station1AStop;
    bool station2EStop;
    bool station2AStop;
    bool station3EStop;
    bool station3AStop;
};

struct StationState {
    bool estopped = false;
    bool astopped = false;
};

enum class AllianceColor {
    RED,
    BLUE,
};

enum class ButtonType {
    E_STOP,
    A_STOP,
};

enum class StationId {
    STATION_1,
    STATION_2,
    STATION_3,
};