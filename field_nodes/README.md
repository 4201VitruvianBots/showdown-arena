# Field Nodes

This repository contains the firmware source code and deployment tooling for the ESP32-based field hardware nodes (such as the `shooter_hub`, score tables, and driver station E-Stops).

## Over-The-Air (OTA) Deployment

With the integration of `ArduinoOTA` in the field node firmware, you can re-flash the nodes remotely over the network without needing a physical USB connection.

### Prerequisites

1.  **PlatformIO:** Ensure you have PlatformIO installed. The script automatically looks for the `pio` command in your environment.
2.  **Network Access:** Ensure your computer is connected to the same network as the ESP32 field nodes. Use ping or another tool to verify availability if necessary.

### Using `deploy_ota.py`

This deployment script allows you to rapidly deploy to multiple node IPs in one go.

#### Deploying to Specific IPs

You can pass one or more IP addresses using the `--ips` argument:

```bash
python deploy_ota.py --ips 10.0.100.11 10.0.100.12
```

#### Deploying to a list of IPs in a file

If you manage a large number of nodes, you can store their IP addresses in a plain text file (e.g., `nodes.txt`, one IP per line, `#` comments allowed):

```bash
python deploy_ota.py --file nodes.txt
```

#### Specifying the Target Project

By default, the script targets the `shooter_hub` project using the `esp32_s3_devkitm_1` environment. You can override these if necessary:

```bash
python deploy_ota.py --ips 10.0.100.11 --project my_other_node --env esp32_custom
```

### Future Considerations

If a node fails to deploy via ArduinoOTA, you may need to fallback to a USB connection. Verify that the correct subnet and route are configured on your host computer, and the node's firewall allows inbound UDP on the default OTA port (3232 for ESP32).
