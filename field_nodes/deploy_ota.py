import argparse
import subprocess
import sys
import os

PIO_PATHS = [
    "pio",
    "pio.exe",
    os.path.expanduser("~/.platformio/penv/Scripts/pio.exe"),
    os.path.expanduser("~/.platformio/penv/bin/pio"),
]

def find_pio():
    for pio in PIO_PATHS:
        try:
            subprocess.run([pio, "--version"], capture_output=True, check=True)
            return pio
        except (FileNotFoundError, subprocess.CalledProcessError):
            continue
    return None

def main():
    parser = argparse.ArgumentParser(description="Deploy firmware to ESP32 nodes via ArduinoOTA.")
    parser.add_argument("--ips", nargs="+", help="List of IP addresses to deploy to")
    parser.add_argument("--file", help="File containing list of IP addresses (one per line)")
    parser.add_argument("--project", default="shooter_hub", help="PlatformIO project directory")
    parser.add_argument("--env", default="esp32_s3_devkitm_1", help="PlatformIO environment")
    args = parser.parse_args()

    ips = []
    if args.ips:
        ips.extend(args.ips)
    if args.file:
        try:
            with open(args.file, "r") as f:
                for line in f:
                    ip = line.strip()
                    if ip and not ip.startswith("#"):
                        ips.append(ip)
        except Exception as e:
            print(f"Error reading file {args.file}: {e}")
            sys.exit(1)

    if not ips:
        print("No IP addresses provided. Use --ips or --file.")
        sys.exit(1)

    pio_cmd = find_pio()
    if not pio_cmd:
        print("Error: PlatformIO (pio) not found. Please ensure it is installed and in your PATH.")
        sys.exit(1)

    print(f"Found PlatformIO at: {pio_cmd}")
    print(f"Deploying to {len(ips)} nodes: {', '.join(ips)}")

    failed_nodes = []

    for ip in ips:
        print(f"\n{'='*50}")
        print(f"Deploying to {ip}...")
        print(f"{'='*50}")
        
        cmd = [
            pio_cmd, "run", 
            "-d", args.project, 
            "-e", args.env, 
            "-t", "upload", 
            "--upload-port", ip
        ]
        
        try:
            result = subprocess.run(cmd)
            if result.returncode != 0:
                print(f"Failed deploying to {ip} (Return code: {result.returncode})")
                failed_nodes.append(ip)
            else:
                print(f"Successfully deployed to {ip}")
        except Exception as e:
            print(f"Error executing deployment to {ip}: {e}")
            failed_nodes.append(ip)

    print(f"\n{'='*50}")
    print("Deployment Summary:")
    print(f"Total nodes: {len(ips)}")
    print(f"Successful : {len(ips) - len(failed_nodes)}")
    print(f"Failed     : {len(failed_nodes)}")
    if failed_nodes:
        print("Failed IPs:")
        for ip in failed_nodes:
            print(f"  - {ip}")
        sys.exit(1)
    else:
        print("All nodes successfully updated!")

if __name__ == "__main__":
    main()
