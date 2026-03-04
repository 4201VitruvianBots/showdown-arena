#include "app/app_scoring.hpp"
#include "app/app_motor.hpp"
#include "app/app_display.hpp"
#include "app/app_led_strip.hpp"
#include "hw/hw_scoring.hpp"
#include "hw/hw_motor.hpp"
#include "hw/hw_display.hpp"
#include "hw/hw_led_strip.hpp"
#include "io/io_comms.h"

#include <Arduino.h>
#include "comms/ethernet.hpp"
#include "comms/udp_comms.hpp"
#include <ArduinoJson.h>
#include "WebServerSetup.h"
#include "GlobalSettings.h"
#include <ArduinoOTA.h>

// Define the base URL for the API (kept for web server setup compatibility)
const char *baseUrl = "http://10.0.100.5:8080";

// Define the IP address and DHCP/Static configuration
extern String deviceRole;
extern String deviceIP;
extern String deviceGWIP;
extern bool useDHCP;
extern String arenaIP;
extern String arenaPort;

// Estop button pin definitions (adjust to your hardware)
// These are example pins — update to match your actual wiring
#define ESTOP_FIELD_PIN 33
#define ESTOP_STATION1E_PIN 1
#define ASTOP_STATION1_PIN 2
#define ESTOP_STATION2E_PIN 3
#define ASTOP_STATION2_PIN 15
#define ESTOP_STATION3E_PIN 18
#define ASTOP_STATION3_PIN 16

// Setup function
void setup()
{
  Serial.begin(115200);

  hw_scoring_init();
  hw_motor_init();
  hw_display_init();
  hw_led_strip_init();

  app_scoring_reset();
  app_motor_reset();
  app_display_init();

  // Initialize preferences
  preferences.begin("settings", false);

  // Load IP address and DHCP/Static configuration from preferences
  deviceIP = preferences.getString("deviceIP", "");
  deviceGWIP = preferences.getString("deviceGWIP", "");
  useDHCP = preferences.getBool("useDHCP", true);

  // Setup Ethernet
  setup_ethernet(useDHCP, deviceIP, deviceGWIP);

  // Print the IP address
  Serial.print("init - IP Address: ");
  Serial.println(getEthernetIPAddress());

  // Set up the web server (for configuration)
  setupWebServer();

  // Load arena IP from preferences and initialize UDP comms
  arenaIP = preferences.getString("arenaIP", "10.0.100.5");
  deviceRole = preferences.getString("deviceRole", "RED_ALLIANCE");
  Serial.printf("Device Role: %s\n", deviceRole.c_str());
  Serial.printf("Arena IP: %s\n", arenaIP.c_str());

  // Initialize UDP communication
  udp_comms_init(arenaIP);

  // Initialize ArduinoOTA
  ArduinoOTA.setHostname(deviceRole.c_str());
  ArduinoOTA.begin();

  // Configure estop pins as inputs with pulldown
  // (Buttons are active-low: pressed = LOW)
  if (deviceRole == "FMS_TABLE")
  {
    pinMode(ESTOP_FIELD_PIN, INPUT_PULLDOWN);
  }
  if (deviceRole == "RED_ALLIANCE" || deviceRole == "BLUE_ALLIANCE")
  {
    // Only configure station estops for alliance hubs
    pinMode(ESTOP_STATION1E_PIN, INPUT_PULLDOWN);
    pinMode(ASTOP_STATION1_PIN, INPUT_PULLDOWN);
    pinMode(ESTOP_STATION2E_PIN, INPUT_PULLDOWN);
    pinMode(ASTOP_STATION2_PIN, INPUT_PULLDOWN);
    pinMode(ESTOP_STATION3E_PIN, INPUT_PULLDOWN);
    pinMode(ASTOP_STATION3_PIN, INPUT_PULLDOWN);
  }
}

// Read estop button states
static ButtonStates readButtonStates()
{
  ButtonStates states;
  // Buttons are active-low (pressed = LOW = true for estop)
  states.fieldEStop = !digitalRead(ESTOP_FIELD_PIN);
  states.station1EStop = !digitalRead(ESTOP_STATION1E_PIN);
  states.station1AStop = !digitalRead(ASTOP_STATION1_PIN);
  states.station2EStop = !digitalRead(ESTOP_STATION2E_PIN);
  states.station2AStop = !digitalRead(ASTOP_STATION2_PIN);
  states.station3EStop = !digitalRead(ESTOP_STATION3E_PIN);
  states.station3AStop = !digitalRead(ASTOP_STATION3_PIN);
  return states;
}

// Latch AStop states for at least one second (10 cycles of 100ms)
static void latchAStops(ButtonStates &states)
{
  static int latch1 = 0;
  static int latch2 = 0;
  static int latch3 = 0;

  if (states.station1AStop)
    latch1 = 10;
  if (states.station2AStop)
    latch2 = 10;
  if (states.station3AStop)
    latch3 = 10;

  if (latch1 > 0)
  {
    states.station1AStop = true;
    latch1--;
  }
  if (latch2 > 0)
  {
    states.station2AStop = true;
    latch2--;
  }
  if (latch3 > 0)
  {
    states.station3AStop = true;
    latch3--;
  }
}

// Main loop
void loop()
{
  app_scoring_run();
  app_motor_run();
  app_led_strip_run();

  // Process incoming UDP commands from the Go server
  udp_comms_run();

  // Process ArduinoOTA
  ArduinoOTA.handle();

  static unsigned long lastSendTime = 0;
  unsigned long currentMillis = millis();

  // Send data and update display every 100ms
  if (currentMillis - lastSendTime >= 100)
  {
    lastSendTime = currentMillis;

    uint32_t currentScore = app_scoring_getScore();
    app_display_updateScore(currentScore);

    // Send unified node_status message (includes estop states + score)
    // The Go server routes by "role" and reads only the relevant fields.
    ButtonStates states = readButtonStates();
    latchAStops(states);
    udp_comms_sendNodeStatus(states, currentScore, deviceRole);
  }

  // Print debug info every 1 second
  static unsigned long lastPrint = 0;
  if (currentMillis - lastPrint >= 1000)
  {
    lastPrint = currentMillis;
    uint32_t currentScore = app_scoring_getScore();
    Serial.printf("Score: %d | Hub: %s | Server: %s | Pattern: <%s>\n",
                  currentScore,
                  udp_comms_getHubState().c_str(),
                  udp_comms_isServerConnected() ? "Connected" : "Disconnected",
                  udp_comms_getLedPattern());
  }

  delay(10); // Small delay to prevent watchdog issues
}
