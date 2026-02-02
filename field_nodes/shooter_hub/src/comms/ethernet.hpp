#ifndef ETHERNET_HPP
#define ETHERNET_HPP
#include <Arduino.h>
#include <ETH.h>
#include <stdbool.h>

void setup_ethernet(bool useDHCP, const String &deviceIP,
                    const String &deviceGWIP);
void onEvent(arduino_event_id_t event, arduino_event_info_t info);

bool isEthernetConnected(void);
IPAddress getEthernetIPAddress(void);

#endif // ETHERNET_HPP