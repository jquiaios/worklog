Output docs/tui.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Dracula"
Set Padding 20
Set TypingSpeed 80ms
Set Framerate 30

# Open the TUI
Type "worklog"
Sleep 800ms
Enter
Sleep 3.5s

# Add a highlight from inside the TUI
Type "n"
Sleep 800ms
Type "Closed the last critical bug before the release"
Sleep 500ms
Enter
Sleep 2s

# Switch to the lowlights column
Tab
Sleep 1.5s

# Add a lowlight
Type "n"
Sleep 800ms
Type "Missed standup three times this sprint"
Sleep 500ms
Enter
Sleep 2s

# Quit
Type "q"
Sleep 500ms
