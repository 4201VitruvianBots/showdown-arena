#include "comms/ethernet.hpp"
#include "pinmaps.hpp"

static bool eth_connected = false;

void onEvent(arduino_event_id_t event, arduino_event_info_t info) {
  switch (event) {
  case ARDUINO_EVENT_ETH_START:
    Serial.println("ETH Started");
    // set eth hostname here
    ETH.setHostname("Freezy_ScoreTable");
    break;
  case ARDUINO_EVENT_ETH_CONNECTED:
    Serial.println("ETH Connected");
    break;
  case ARDUINO_EVENT_ETH_GOT_IP:
    Serial.printf("ETH Got IP: '%s'\n",
                  esp_netif_get_desc(info.got_ip.esp_netif));
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

bool isEthernetConnected(void) { return eth_connected; }

static IPAddress local_ip(192, 168, 10, 220);
static IPAddress local_gateway(192, 168, 10, 1);
static IPAddress subnet(255, 255, 255, 0);
static IPAddress primaryDNS(8, 8, 8, 8);
static IPAddress secondaryDNS(8, 8, 4, 4);

void setup_ethernet(bool useDHCP, const String &deviceIP,
                    const String &deviceGWIP) {
  Network.onEvent(onEvent);
  // Initialize Ethernet with DHCP or Static IP
  if (useDHCP) {
    ETH.begin(ETH_PHY_TYPE, ETH_PHY_ADDR, ETH_PHY_CS, ETH_PHY_IRQ, ETH_PHY_RST,
              ETH_PHY_SPI_HOST, ETH_PHY_SPI_SCK, ETH_PHY_SPI_MISO,
              ETH_PHY_SPI_MOSI);
  } else {
    IPAddress localIP;
    IPAddress localGW;
    if (localIP.fromString(deviceIP)) {
      Serial.println("Setting static IP address.");
      // THis is not working Need to fix
      localGW.fromString(deviceGWIP);
      ETH.begin(ETH_PHY_TYPE, ETH_PHY_ADDR, ETH_PHY_CS, ETH_PHY_IRQ,
                ETH_PHY_RST, ETH_PHY_SPI_HOST, ETH_PHY_SPI_SCK,
                ETH_PHY_SPI_MISO, ETH_PHY_SPI_MOSI);
      ETH.config(localIP, localGW, subnet, primaryDNS, secondaryDNS);
    } else {
      Serial.println("Invalid static IP address. Falling back to DHCP.");
      ETH.begin(ETH_PHY_TYPE, ETH_PHY_ADDR, ETH_PHY_CS, ETH_PHY_IRQ,
                ETH_PHY_RST, ETH_PHY_SPI_HOST, ETH_PHY_SPI_SCK,
                ETH_PHY_SPI_MISO, ETH_PHY_SPI_MOSI);
    }
  }

  // Print the IP address
  Serial.print("init - IP Address: ");
  Serial.println(ETH.localIP());
}

IPAddress getEthernetIPAddress(void) { return ETH.localIP(); }