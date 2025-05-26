# atcli – Modern AT Command Interface for Modems

## 🧭 Overview

atcli is a modern CLI tool written in Go for interacting with AT command–based cellular and GPS/GNSS modems. It aims to replace traditional tools like minicom and screen with a purpose-built, interruption-free, and modern interface designed for today's development workflows.

## 📸 Screenshot

![atcli screenshot](./screenshot.png)

## 🤔 Why atcli?

Traditional terminal-based modem interfaces like minicom were designed decades ago and lack many features that modern developers expect:

- **Clean separation of commands and responses** - Unlike minicom where commands and unsolicited messages mix together, atcli clearly separates your input from modem output
- **Interruption-free experience** - No more lost commands when a modem sends unsolicited notifications
- **Modern UX** - Intuitive interface with command history, filtering, and visual cues
- **Developer-friendly** - Built for cellular/IoT developers who work with modems daily

atcli addresses these pain points by providing a specialized tool that understands AT command workflows and modem behavior patterns, making development and debugging significantly more efficient.

## ✨ Core Features (Implemented)

- Serial port connection via go.bug.st/serial
- Interactive REPL-style input
- Command buffer separated from unsolicited modem output
- Will show an arrow `<-` or `->` to indicate if the output is from the modem or from the user
- Will show a line number to indicate the command number so the user can determine something is happening and make it easy to see changes
- Default serial settings: /dev/serial0, 115200 baud

---

## 📌 Roadmap: Planned Features

Some of these features might be removed if they are found to be impossible, or maybe I was just wishfully thinking they would work. I will update this list as I progress.

### 1. 💬 Command Interface
- Input history navigation (up/down arrow) ✅
- Command auto-complete for known AT commands
- Multi-line/multi-command blocks (e.g. send several commands in sequence)
- Graceful Ctrl+C handling with optional cleanup/reset

### 2. 🐛 Debugging Tools
- Signal strength polling (AT+CSQ loop with graph) ✅
- Live registration status
- Detect known error patterns (e.g. SIM failure, boot loops)

### 3. 📌 GPS Features
- GPS location polling (AT+CGPSINFO loop with graph) ✅
- GPS Time

### 4. 📱 Network Information
- Network registration status monitor (AT+CREG, AT+CGREG, AT+CEREG)
- Display network type (2G/3G/4G/5G), registration status, and cell ID
- Network scan tool to show available operators (AT+COPS=?)
- Operator selection interface

### 5. 💳 SIM Card Management
- SIM card information display (AT+CIMI, AT+CCID)
- PIN status and management
- SIM toolkit integration

### 6. ℹ️ Modem Information
- Comprehensive modem information panel (AT+CGMI, AT+CGMM, AT+CGMR)
- IMEI and capability reporting (AT+GSN)
- Firmware version and update status

### 7. 📨 SMS Management
- SMS reading and sending capabilities (AT+CMGR, AT+CMGS)
- Message storage and browsing interface
- SMS configuration options

### 8. 🌐 Data Connection Management
- PDP context management (AT+CGDCONT, AT+CGACT)
- APN configuration interface
- Data connection statistics

### 9. 🔋 Power Management
- Power saving mode controls (AT+CPSMS)
- Battery status monitoring (where supported)
- Power consumption optimization tools

### 10. 📋 Command Templates
- Library of common AT command sequences
- User-definable command macros
- Profile-based configuration management

⸻---

## 🔧 Project Structure

```
/cmd      # CLI entry point
/layouts  # UI layout files
/views    # View files
/services # Service files
/types    # Type definitions

```

⸻

## ✅ Example Usage

atcli opens a modern two-panel terminal interface:

- **Left panel:** Command history buffer. Enter AT commands in the text entry at the bottom; your previous commands appear here for easy recall and editing.
- **Right panel:** Numbered list of outputs. Each command’s response and any unsolicited modem output are grouped and displayed clearly, making it easy to track which output belongs to which command.
- Entering `/log` will open a small log panel where certain logging messages might appear if things aren't working as expected.
- Entering `/signal` will open a small signal page where it will show you the signal strength of the modem.
- Entering `/gps` will open a small GPS page where it will show you the GPS coordinates of the modem.
- Entering `/help` will open a small help page where certain help messages might appear if things aren't working as expected.
- Entering `/<cmd> close` will close the page or panel currently open, closing a page navigates back to the home page, closing a panel just closes that panel
- Entering `/quit` will close the app.
- The argument --version will print the version of the app.
- The argument --port will set the serial port to use. E.g. `--port /dev/serial0`
- The argument --baud will set the baud rate to use. E.g. `--baud 115200`

⸻

## 📬 Dependencies

- `go.bug.st/serial` – Serial I/O abstraction

⸻---

## 🧑‍💻 Development

Normally, I run `make build arm64 linux` to build the binary for my device. Then I run `./build/atcli-linux-arm64` to run the binary.

I build the arm64 / darwin binary for my macbook and I can provide `/dev/ttys0` 

## 📱 Use as a modem to get internet access

**SIMCom A7670E PPP Setup Guide (pppd😉 over UART)**

This guide explains how to configure and use `pppd` on a Linux system (e.g., Raspberry Pi) to establish a cellular data connection using a SIMCom A7670E module over UART.

---

#### 2. Files and Configuration

##### /etc/ppp/peers/simcom

```
/dev/serial0 115200
connect "/usr/sbin/chat -v -f /etc/ppp/chat-connect"
noauth
defaultroute
usepeerdns
persist
nodetach
```

##### /etc/ppp/chat-connect

```
ABORT "BUSY"
ABORT "NO CARRIER"
ABORT "ERROR"
ABORT "NO DIALTONE"
TIMEOUT 10
"" AT
OK AT+CGDCONT=1,"IP","internet"
OK ATD*99#
CONNECT ""
```

> Replace `"internet"` with your carrier's APN if different.

---

#### 3. Start the Connection

Run the following command to start the PPP session with verbose output:

```bash
sudo pppd call simcom nodetach debug dump logfd 2
```

This will:

* Keep output in the terminal
* Show detailed log information
* Attempt to dial and bring up the PPP interface (e.g., `ppp0`)

---

#### 4. Troubleshooting

##### Check interface:

```bash
ip a show ppp0
```

##### Test connectivity:

```bash
ping 8.8.8.8
```

##### DNS issues?

If name resolution fails:

* Check `/etc/resolv.conf`
* If it is a symlink to systemd, replace it with:

  ```bash
  sudo rm /etc/resolv.conf
  sudo cp /etc/ppp/resolv.conf /etc/resolv.conf
  ```

---

#### 5. Notes

* No SIM PIN is required for most IoT SIMs, but if your SIM is locked, unlock it using `AT+CPIN="1234"`
* Ensure GNSS and PPP are powered separately; PPP does **not** require GNSS
* The `persist` option causes automatic redialing if disconnected

---

## 📚 Interesting Inforamtion

```
Sending SMS using USB From Computer 
Connect Modem to Computer using either on board USB Port or using External USB to Serial converter board.
Open Serial port software and enter the below command to send SMS.
 
AT<CR><LF> 
Attention Command,  this signifies that our Modem is working properly. 
Answer Expected : OK

ATE0<CR><LF> 
This Command is being sent to stop the echo.
Answer Expected : OK

AT+CREG?<CR><LF> 
It is being used to check whether the SIM got registered with the Network.
Answer Expected : +CREG: 0,1  or +CREG: 0,5

AT+CMGF=1<CR><LF> 
Configuring Text mode for sending SMS
Answer Expected : OK

AT+CMGS="Mobile Number"<CR><LF> 
Set the destination mobile number enclosed in the DOUBLE QUOTES.
Answer Expected : >

"Hi, How are you?"<SUB>
Here we enter our message body followed by CONTROL-Z (<SUB>) 
Answer Expected : SMS confirmation starting with "+CMGS"
```
---

```
Reading SMS using Modem 
AT<CR><LF> 
Attention Command,  this signifies that our Modem is working properly. 
Answer Expected : OK
ATE0<CR><LF> 
This Command is being sent to stop the echo.
Answer Expected : OK
AT+CREG?<CR><LF> 
It is being used to check whether the SIM got registered with the Network.
Answer Expected : +CREG: 0,1  or +CREG: 0,5
AT+CMGF=1<CR><LF> 
Configuring Text mode for SMS
Answer Expected : OK
AT+CNMI=2,1,0,0,0<CR><LF>
Configure Modem to notify on incoming SMS reciept. Modem will inform the SMS reciept by "+CMTI:<Message index Number> "
AT+CMGR=<Message index Number>
Now Read SMS using <Message index Number> received.
```

---

```
How to get GPS Location using simcom A7672S 4G LTE Modem
AT<CR><LF> 
Attention Command,  this signifies that our Modem is working properly. 
Answer Expected : OK
AT_CGNSSPWR=1<CR><LF>
should be executed to let GNSS module power on firstly and have to wait till getting the confirmation . may take 10-30Seconds.
Answer Expected : "OK" and "+CGNSSPWR: READY!" afetr few seconds
AT+CGPSCOLD<CR><LF> 
COLD start GNSS, When first used;
Answer Expected : OK
AT+CGNSSTST=1<CR><LF> 
Answer Expected : OK
AT+AGPS<CR><LF> 
Activating Assisted GPS function for better accuracy.
Answer Expected : OK
AT+CGPSINFO<CR><LF> 
Get GPS fixed position information , After 1 minute , GPS location information will be recived .
Answer Expected : OK
```
--- 

```
This will get you a NMEA stream from the modem

AT+CGDRT=4,1
AT+CGSETV=4,0
AT+CGNSSPWR=1

Waiting to return >  +CGNSSPWR: READY! 
Next send

AT+CGNSSMODE=3
AT+CGNSSNMEA=1,1,1,1,1,1,0,0
AT+CGPSNMEARATE=2
AT+CGNSSTST=1
AT+CGNSSPORTSWITCH=0,1
```

---

```
-> AT+CPSI?

+CPSI: LTE,Online,262-03,0x5607,5506356,266,EUTRAN-BAND1,300,5,5,30,55,52,21
```

---

```
-> AT+SIMCOMATI

Manufacturer: SIMCOM INCORPORATED                                                                   
Model: A7670E-MASA                                                                                   
Revision: A131B02A7670M6C_M                                                                          
A7670M6_B02V03_241108                                                                                
IMEI: 863957072600361
```

---

| Command      | Description  | Return                   |
|--------------|--------------|-------------------------:|
| AT           | AT Test Command | OK |
| ATE          | ATE1 Set Up Echo | OK |
| ATE0         | Turn off Echo | OK |
| AT+SIMCOMATI | Query Module Information | OK |
| AT+IPREX     | Setting the module hardware serial port baud rate | +IPREX: OK |
| AT+CRESET    | Reset Module | OK |
| AT+CSQ       | Network signal quality check, returning signal strength value | +CSQ: 25,99 OK |
| AT+CPIN?     | Check SIM card status, returning 'READY,' indicating the SIM card is recognized and functioning properly | +CPIN: READY |
| AT+COPS?     | Query the current network operator; upon successful connection, it will return information about the network | +COPS: OK |
| AT+CREG?     | Query network registration status | +CREG: OK |
| AT+CPSI?     | Query UE (User Equipment) system information | +CPSI: LTE,Online,262-03,0x5607,5506356,266,EUTRAN-BAND1,300,5,5,30,55,52,21 |
| AT+CNMP      | Network mode selection command: <br>2: Automatic<br>13: GSM only<br>14: WCDMA only<br>38: LTE only | OK |

---

This waveshare page has an interesting page with a lot of information on it

[text](https://www.waveshare.com/wiki/A7670E_Cat-1/GNSS_HAT)

---

The common commands for GNSS satellite positioning functionality are as follows:

- AT+CGNSSPWR=1 // Activate GNSS
- AT+CGNSSTST=1 // Enable information output
- AT+CGNSSPORTSWITCH=0,1 // Switch NMEA data output port
- AT+CGPSINFO // Retrieve satellite latitude and longitude data