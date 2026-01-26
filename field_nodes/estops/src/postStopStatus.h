/*  ______                              ___
   / ____/_______  ___  ____  __  __   /   |  ________  ____  ____ _
  / /_  / ___/ _ \/ _ \/_  / / / / /  / /| | / ___/ _ \/ __ \/ __ `/
 / __/ / /  /  __/  __/ / /_/ /_/ /  / ___ |/ /  /  __/ / / / /_/ /
/_/ __/_/___\___/\___/ /___/\__, /  /_/  |_/_/   \___/_/ /_/\__,_/
   / ____/ ___// /_____  __/____/___
  / __/  \__ \/ __/ __ \/ __ \/ ___/
 / /___ ___/ / /_/ /_/ / /_/ (__  )
/_____//____/\__/\____/ .___/____/
                     /_/
*/
#ifndef POSTSTOPSTATUS_H
#define POSTSTOPSTATUS_H

#include <HTTPClient.h>
#include <ArduinoJson.h>
#include "GlobalSettings.h"
#include "common.h"

extern const char *baseUrl;
extern bool eth_connected;

/**
 * Sends an HTTP POST request to update the stop status for all 6 channels.
 *
 * @param stopButtonStates An array of the states of the stop buttons (false if pressed, true otherwise).
 */
void postAllStopStatus(const ButtonStates &states, String deviceRole)
{
    // Send the HTTP POST request
    if (eth_connected)
    {
        HTTPClient http;

        // Define payload
        StaticJsonDocument<200> payload;
        JsonArray array = payload.to<JsonArray>();

        int offset = 0;
        if (deviceRole == "RED_ALLIANCE")
        {
            offset = 0;
        }
        else if (deviceRole == "BLUE_ALLIANCE")
        {
            offset = 6;
        }
        else if (deviceRole == "FMS_TABLE")
        {
            // Only post the field Estop state if this is the FMS table.
            JsonObject channel = array.createNestedObject();
            channel["channel"] = 0;
            channel["state"] = states.fieldEStop;
        }

        if (deviceRole == "RED_ALLIANCE" || deviceRole == "BLUE_ALLIANCE")
        {

            {
                JsonObject channel = array.createNestedObject();
                channel["channel"] = 1 + offset;
                channel["state"] = states.station1EStop;
            }
            {
                JsonObject channel = array.createNestedObject();
                channel["channel"] = 2 + offset;
                channel["state"] = states.station1AStop;
            }
            {
                JsonObject channel = array.createNestedObject();
                channel["channel"] = 3 + offset;
                channel["state"] = states.station2EStop;
            }
            {
                JsonObject channel = array.createNestedObject();
                channel["channel"] = 4 + offset;
                channel["state"] = states.station2AStop;
            }
            {
                JsonObject channel = array.createNestedObject();
                channel["channel"] = 5 + offset;
                channel["state"] = states.station3EStop;
            }
            {
                JsonObject channel = array.createNestedObject();
                channel["channel"] = 6 + offset;
                channel["state"] = states.station3AStop;
            }
        }

        // Convert payload to JSON string
        String jsonString;
        serializeJson(payload, jsonString);

        // Configure HTTP POST request
        String url = String(baseUrl) + "/api/freezy/eStopState";
        Serial.println("URL: " + url); // Print the URL
        http.begin(url);
        http.addHeader("Content-Type", "application/json");

        // Send the request
        int httpResponseCode = http.POST(jsonString);

        // Handle the response
        if (httpResponseCode > 0)
        {
            Serial.println("postAllStopStatus");
            Serial.printf("Request successful! HTTP code: %d\n", httpResponseCode);
            String response = http.getString();
            Serial.println("Response:");
            Serial.println(response);
        }
        else
        {
            Serial.println("postAllStopStatus");
            Serial.printf("Request failed! Error code: %d\n", httpResponseCode);
        }

        // Close the connection
        http.end();
    }
    else
    {
        Serial.println("Network not connected![PSS]");
    }
}

#endif // POSTSTOPSTATUS_H