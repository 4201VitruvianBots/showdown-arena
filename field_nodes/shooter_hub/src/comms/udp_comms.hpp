#ifndef UDP_COMMS_HPP
#define UDP_COMMS_HPP

#include <Arduino.h>
#include "common.h"

// Default ports matching the Go server
#define UDP_LISTEN_PORT 5300  // ESP32 listens for commands from Go
#define UDP_SEND_PORT   5301  // ESP32 sends data to Go

/**
 * @brief Initialize UDP communication.
 * @param arenaIP The IP address of the Go arena server to send data to.
 */
void udp_comms_init(const String &arenaIP);

/**
 * @brief Process incoming UDP messages. Call from loop().
 */
void udp_comms_run(void);

/**
 * @brief Send estop/astop button states to the Go server.
 * @param states The current button states.
 * @param role Device role: "RED_ALLIANCE", "BLUE_ALLIANCE", or "FMS_TABLE"
 */
void udp_comms_sendEstopState(const ButtonStates &states, const String &role);

/**
 * @brief Send a score report to the Go server.
 * @param score The current score count.
 * @param role Score role: "RED_HUB" or "BLUE_HUB"
 */
void udp_comms_sendScoreReport(uint32_t score, const String &role);

/**
 * @brief Check if the Go server is sending us commands (heartbeat).
 * @return true if a command was received within the timeout window.
 */
bool udp_comms_isServerConnected(void);

// --- Received command accessors ---

/**
 * @brief Get the hub state string received from the Go server.
 * @return The hub state string (e.g. "DISABLED", "SCORING_ACTIVE", etc.)
 */
String udp_comms_getHubState(void);

/**
 * @brief Get the motor duty command received from the Go server.
 * @return Motor duty value (-1.0 to 1.0)
 */
float udp_comms_getMotorDuty(void);

/**
 * @brief Get the red hub light state from the Go server.
 */
bool udp_comms_getRedLight(void);

/**
 * @brief Get the blue hub light state from the Go server.
 */
bool udp_comms_getBlueLight(void);

/**
 * @brief Get the match state received from the Go server.
 */
int udp_comms_getMatchState(void);

#endif // UDP_COMMS_HPP
