#include "app/app_scoring.hpp"
#include "app/app_motor.hpp"
#include "hw/hw_scoring.hpp"
#include "hw/hw_motor.hpp"
#include "io/io_comms.h"

#include <Arduino.h>
#include "comms/ethernet.hpp"
#include <ArduinoJson.h>
#include "WebServerSetup.h"          // Include the WebServerSetup header
#include "GlobalSettings.h"          // Include the GlobalSettings header"

#define USE_SERIAL Serial
//#define USE_ETHERNET

// Define the base URL for the API
const char *baseUrl = "http://10.0.100.5:8080";

// Define the IP address and DHCP/Static configuration
extern String deviceRole;
extern String deviceIP;
extern String deviceGWIP;
extern bool useDHCP;


// Setup function
void setup()
{
  Serial.begin(115200);

  hw_scoring_init();
  hw_motor_init();

  app_scoring_reset();
  app_motor_reset();

  // Initialize preferences
  preferences.begin("settings", false);

#ifdef USE_ETHERNET
  // Load IP address and DHCP/Static configuration from preferences
  deviceIP = preferences.getString("deviceIP", "");
  deviceGWIP = preferences.getString("deviceGWIP", "");
  useDHCP = preferences.getBool("useDHCP", true);

  // Setup Ethernet
  setup_ethernet(useDHCP, deviceIP, deviceGWIP);

  // Print the IP address
  Serial.print("init - IP Address: ");
  Serial.println(getEthernetIPAddress());
  // Set up the web server
  setupWebServer();
#endif
}

// Main loop
void loop()
{

  app_scoring_run();
  app_motor_run();

  static unsigned long lastPrint = 0;
  unsigned long currentMillis = millis();

  // print the IP address every 1 seconds
  if (currentMillis - lastPrint >= 1000)
  {
    Serial.printf("Current Ball Scoring Count: %d\n", app_scoring_getScore());
    lastPrint = currentMillis;
#ifdef USE_ETHERNET
    lastPrint = currentMillis;
    deviceIP = preferences.getString("deviceIP", "");
    Serial.printf("Preferences IP Address: %s\n", deviceIP.c_str());
    deviceGWIP = preferences.getString("deviceGWIP", "");
    Serial.printf("Preferences Gateway IP Address: %s\n", deviceGWIP.c_str());
    useDHCP = preferences.getBool("useDHCP", true);
    Serial.printf("Preferences useDHCP: %s\n", useDHCP ? "true" : "false");
    const IPAddress currentIP = getEthernetIPAddress();
    Serial.printf("Current Wired IP Address: %s\n", currentIP.toString().c_str());
#endif
  }

  delay(250);
}
