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
#include <Arduino.h>
#include <ETH.h>
#include <ArduinoJson.h>
#include "postStopStatus.h"          // Include the postStopStatus header
#include "WebServerSetup.h"          // Include the WebServerSetup header
#include "GlobalSettings.h"          // Include the GlobalSettings header
#include "common.h"

#ifndef ETH_PHY_CS
#define ETH_PHY_TYPE ETH_PHY_W5500
#define ETH_PHY_ADDR 1
#define ETH_PHY_CS 14
#define ETH_PHY_IRQ 10
#define ETH_PHY_RST 9
#define ETH_PHY_SPI_HOST SPI2_HOST
#define ETH_PHY_SPI_SCK 13
#define ETH_PHY_SPI_MISO 12
#define ETH_PHY_SPI_MOSI 11
#endif

#define USE_SERIAL Serial

// Define the base URL for the API
const char *baseUrl = "http://10.0.100.5:8080";

// Define the IP address and DHCP/Static configuration
extern String deviceRole;
extern String deviceIP;
extern String deviceGWIP;
extern bool useDHCP;

// Pins connected to the stop button
#define NUM_BUTTONS 7

const uint8_t field_estop_pin = 33;
const uint8_t station_1_estop_pin = 1;
const uint8_t station_1_astop_pin = 2;
const uint8_t station_2_estop_pin = 3;
const uint8_t station_2_astop_pin = 15;
const uint8_t station_3_estop_pin = 18;
const uint8_t station_3_astop_pin = 16;

const int stopButtonPins[NUM_BUTTONS] = {33,  // Field stop
                                         1,   // 1E stop
                                         2,   // 1A stop
                                         3,   // 2E stop
                                         15,  // 2A stop
                                         18,  // 3E stop
                                         16}; // 3A stop



bool eth_connected = false;

void onEvent(arduino_event_id_t event, arduino_event_info_t info)
{
  switch (event)
  {
  case ARDUINO_EVENT_ETH_START:
    Serial.println("ETH Started");
    // set eth hostname here
    ETH.setHostname("Freezy_ScoreTable");
    break;
  case ARDUINO_EVENT_ETH_CONNECTED:
    Serial.println("ETH Connected");
    break;
  case ARDUINO_EVENT_ETH_GOT_IP:
    Serial.printf("ETH Got IP: '%s'\n", esp_netif_get_desc(info.got_ip.esp_netif));
    Serial.println(ETH);
    eth_connected = true;
    break;
  case ARDUINO_EVENT_ETH_LOST_IP:
    Serial.println("ETH Lost IP");
    eth_connected = false;
    break;
  case ARDUINO_EVENT_ETH_DISCONNECTED:
    Serial.println("ETH Disconnected");
    eth_connected = false;
    break;
  case ARDUINO_EVENT_ETH_STOP:
    Serial.println("ETH Stopped");
    eth_connected = false;
    break;
  default:
    break;
  }
}

IPAddress local_ip(192, 168, 10, 220);
IPAddress local_gateway(192, 168, 10, 1);
IPAddress subnet(255, 255, 255, 0);
IPAddress primaryDNS(8, 8, 8, 8);
IPAddress secondaryDNS(8, 8, 4, 4);

// Setup function
void setup()
{
  Serial.begin(115200);

  // Initialize the stop buttons
  for (int i = 0; i < NUM_BUTTONS; i++)
  {
    pinMode(stopButtonPins[i], INPUT);
  }

  // Initialize preferences
  preferences.begin("settings", false);

  // Load IP address and DHCP/Static configuration from preferences
  deviceIP = preferences.getString("deviceIP", "");
  deviceGWIP = preferences.getString("deviceGWIP", "");
  useDHCP = preferences.getBool("useDHCP", true);

  Network.onEvent(onEvent);
  // Initialize Ethernet with DHCP or Static IP
  if (useDHCP)
  {
    ETH.begin(ETH_PHY_TYPE, ETH_PHY_ADDR, ETH_PHY_CS, ETH_PHY_IRQ, ETH_PHY_RST, ETH_PHY_SPI_HOST, ETH_PHY_SPI_SCK, ETH_PHY_SPI_MISO, ETH_PHY_SPI_MOSI);
  }
  else
  {
    IPAddress localIP;
    IPAddress localGW;
    if (localIP.fromString(deviceIP))
    {
      Serial.println("Setting static IP address.");
      // THis is not working Need to fix
      localGW.fromString(deviceGWIP);
      ETH.begin(ETH_PHY_TYPE, ETH_PHY_ADDR, ETH_PHY_CS, ETH_PHY_IRQ, ETH_PHY_RST, ETH_PHY_SPI_HOST, ETH_PHY_SPI_SCK, ETH_PHY_SPI_MISO, ETH_PHY_SPI_MOSI);
      ETH.config(localIP, localGW, subnet, primaryDNS, secondaryDNS);
    }
    else
    {
      Serial.println("Invalid static IP address. Falling back to DHCP.");
      ETH.begin(ETH_PHY_TYPE, ETH_PHY_ADDR, ETH_PHY_CS, ETH_PHY_IRQ, ETH_PHY_RST, ETH_PHY_SPI_HOST, ETH_PHY_SPI_SCK, ETH_PHY_SPI_MISO, ETH_PHY_SPI_MOSI);
    }
  }

  // Print the IP address
  Serial.print("init - IP Address: ");
  Serial.println(ETH.localIP());

  // Set up the web server
  setupWebServer();
}

// Main loop
void loop()
{
  static unsigned long lastPrint = 0;
  unsigned long currentMillis = millis();
  deviceRole = preferences.getString("deviceRole", "RED_ALLIANCE");

  ButtonStates buttonStates;
  buttonStates.fieldEStop = digitalRead(field_estop_pin);
  buttonStates.station1EStop = digitalRead(station_1_estop_pin);
  buttonStates.station1AStop = !digitalRead(station_1_astop_pin);
  buttonStates.station2EStop = digitalRead(station_2_estop_pin);
  buttonStates.station2AStop = !digitalRead(station_2_astop_pin);
  buttonStates.station3EStop = digitalRead(station_3_estop_pin);
  buttonStates.station3AStop = !digitalRead(station_3_astop_pin);

  // Call the postAllStopStatus method with the array
  postAllStopStatus(buttonStates, deviceRole);

  // print the IP address every 5 seconds
  if (currentMillis - lastPrint >= 5000)
  {
    lastPrint = currentMillis;
    deviceIP = preferences.getString("deviceIP", "");
    Serial.printf("Preferences IP Address: %s\n", deviceIP.c_str());
    deviceGWIP = preferences.getString("deviceGWIP", "");
    Serial.printf("Preferences Gateway IP Address: %s\n", deviceGWIP.c_str());
    useDHCP = preferences.getBool("useDHCP", true);
    Serial.printf("Preferences useDHCP: %s\n", useDHCP ? "true" : "false");

    Serial.printf("Current Wired IP Address: %s\n", ETH.localIP().toString().c_str());
  }

  delay(500);
}
