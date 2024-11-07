package main

const xmlPreamble = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
`

var defaultLaunchdConfigXML = xmlPreamble +
	`<plist version="1.0">` +
	`<dict>` +
	`<key>Label</key>` +
	`<string>otelcol-sumo</string>` +
	`<key>ProgramArguments</key>` +
	`<array>` +
	`<string>/usr/share/otelcol-sumo.sh</string>` +
	`</array>` +
	`<key>EnvironmentVariables</key>` +
	`<dict>` +
	`<key>SUMOLOGIC_INSTALLATION_TOKEN</key>` +
	`<string></string>` +
	`</dict>` +
	`<!-- Service user -->` +
	`<key>UserName</key>` +
	`<string>_otelcol-sumo</string>` +
	`<!-- Service group -->` +
	`<key>GroupName</key>` +
	`<string>_otelcol-sumo</string>` +
	`<!-- Run the service immediately after it is loaded -->` +
	`<key>RunAtLoad</key>` +
	`<true/>` +
	`<!-- Restart the process if it exits -->` +
	`<key>KeepAlive</key>` +
	`<true/>` +
	`<!-- Redirect stdout to a log file -->` +
	`<key>StandardOutPath</key>` +
	`<string>/var/log/otelcol-sumo/otelcol-sumo.log</string>` +
	`<!-- Redirect stderr to a log file -->` +
	`<key>StandardErrorPath</key>` +
	`<string>/var/log/otelcol-sumo/otelcol-sumo.log</string>` +
	`</dict>` +
	`</plist>`
