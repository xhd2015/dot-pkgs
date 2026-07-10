package space

// AppleScripts for Mission Control. Verified on macOS 26.5.1 (arm64):
// Dock > Mission Control > Spaces Bar.
// Keep System Events terms inside tell blocks.

const scriptCreate = `
try
	do shell script "open -b 'com.apple.exposelauncher'"
	delay 0.6
	tell application "System Events"
		set dockProc to first application process whose bundle identifier is "com.apple.dock"
		tell dockProc
			set mc to first group whose name is "Mission Control"
			set inner to group 1 of mc
			set spacesBar to first group of inner whose name is "Spaces Bar"
			set addBtn to first button of spacesBar whose (value of attribute "AXDescription" is "add desktop")
			click addBtn
		end tell
		delay 0.35
		key code 53
	end tell
	return "OK: created Space via Spaces Bar > add desktop"
on error e number n
	try
		tell application "System Events" to key code 53
	end try
	return "FAIL: " & e & " (" & n & ")"
end try
`

const scriptSwitch = `
on run argv
	if (count of argv) < 1 then
		return "FAIL: usage: switch <desktopNumber>"
	end if
	set targetNum to item 1 of argv as text
	set targetDesc to "exit to Desktop " & targetNum
	set targetName to "Desktop " & targetNum
	try
		do shell script "open -b 'com.apple.exposelauncher'"
		delay 0.6
		tell application "System Events"
			set dockProc to first application process whose bundle identifier is "com.apple.dock"
			set clicked to false
			set how to ""
			try
				tell dockProc
					set mc to first group whose name is "Mission Control"
					set inner to group 1 of mc
					set spacesBar to first group of inner whose name is "Spaces Bar"
					try
						set btn to first button of spacesBar whose name is targetName
						click btn
						set clicked to true
						set how to "Spaces Bar button name=" & targetName
					end try
					if not clicked then
						try
							set lst to first list of spacesBar
							set btn2 to first button of lst whose name is targetName
							click btn2
							set clicked to true
							set how to "Spaces Bar list button name=" & targetName
						end try
					end if
				end tell
			end try
			if not clicked then
				tell dockProc
					set entire to entire contents
					repeat with el in entire
						try
							if (role of el as text) is "AXButton" then
								set d to value of attribute "AXDescription" of el
								if d is not missing value then
									if (d as text) is targetDesc then
										click el
										set clicked to true
										set how to "entire contents desc=" & targetDesc
										exit repeat
									end if
								end if
							end if
						end try
					end repeat
				end tell
			end if
			if not clicked then
				key code 53
				return "FAIL: desktop not found: " & targetName
			end if
			delay 0.25
			try
				key code 53
			end try
		end tell
		return "OK: switched to " & targetName & " via " & how
	on error e number n
		try
			tell application "System Events" to key code 53
		end try
		return "FAIL: " & e & " (" & n & ")"
	end try
end run
`

const scriptHighest = `
try
	do shell script "open -b 'com.apple.exposelauncher'"
	delay 0.55
	tell application "System Events"
		set dockProc to first application process whose bundle identifier is "com.apple.dock"
		set maxN to 0
		tell dockProc
			set entire to entire contents
			repeat with el in entire
				try
					if (role of el as text) is "AXButton" then
						set d to value of attribute "AXDescription" of el
						if d is not missing value then
							set ds to d as text
							if ds starts with "exit to Desktop " then
								set numText to text ((length of "exit to Desktop ") + 1) thru -1 of ds
								try
									set num to numText as integer
									if num > maxN then set maxN to num
								end try
							end if
						end if
					end if
				end try
			end repeat
		end tell
		key code 53
		if maxN is 0 then return "FAIL: no Desktop buttons found"
		return maxN as text
	end tell
on error e number n
	try
		tell application "System Events" to key code 53
	end try
	return "FAIL: " & e & " (" & n & ")"
end try
`

const scriptList = `
try
	do shell script "open -b 'com.apple.exposelauncher'"
	delay 0.55
	tell application "System Events"
		set dockProc to first application process whose bundle identifier is "com.apple.dock"
		set n to 0
		set names to {}
		tell dockProc
			set entire to entire contents
			repeat with el in entire
				try
					if (role of el as text) is "AXButton" then
						set d to value of attribute "AXDescription" of el
						if d is not missing value then
							set ds to d as text
							if ds starts with "exit to Desktop " then
								set n to n + 1
								try
									set end of names to (name of el as text)
								end try
							end if
						end if
					end if
				end try
			end repeat
		end tell
		key code 53
		set nameList to ""
		repeat with nm in names
			if nameList is "" then
				set nameList to nm as text
			else
				set nameList to nameList & ", " & (nm as text)
			end if
		end repeat
		return "count=" & (n as text) & " desktops=[" & nameList & "]"
	end tell
on error e number n
	try
		tell application "System Events" to key code 53
	end try
	return "FAIL: " & e & " (" & n & ")"
end try
`
