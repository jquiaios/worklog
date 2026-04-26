Output docs/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 600
Set Theme "Dracula"
Set Padding 20
Set TypingSpeed 60ms
Set Framerate 30

Sleep 1s

# Step 1: log a highlight
Type "worklog hl 'Shipped the multi-tenant feature'"
Sleep 500ms
Enter
Sleep 1.5s

# Step 2: log a lowlight
Type "worklog ll 'Prod outage - 2 hours lost'"
Sleep 500ms
Enter
Sleep 1.5s

# Step 3: list to confirm both are stored
Type "worklog list"
Sleep 800ms
Enter
Sleep 3s

# Step 4: open the TUI and let the layout breathe
Type "worklog"
Sleep 800ms
Enter
Sleep 5s
Type "q"
Sleep 500ms
