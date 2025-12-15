# Gitta Demo Script
# Demonstrate Gitta’s core value and quick usage

Output docs/demo.gif

Set FontSize 16
Set Width 1400
Set Height 800
Set Theme "catppuccin-mocha"
Set Padding 20

# 1) Show help - list all available commands
Type "./gitta --help"
Sleep 500ms
Enter
Sleep 3s

# 2) Init workspace - quick start
Type "./gitta init"
Sleep 500ms
Enter
Sleep 2s

# 3) View current Sprint tasks
Type "./gitta list"
Sleep 500ms
Enter
Sleep 2.5s

# 4) View all tasks (Sprint + Backlog)
Type "./gitta list --all"
Sleep 500ms
Enter
Sleep 2.5s

# 5) Create a new story
Type "./gitta story create --title 'Add user authentication'"
Sleep 500ms
Enter
Sleep 2s

# 6) Start working (auto branch creation)
Type "./gitta start US-001"
Sleep 500ms
Enter
Sleep 2s

# 7) Verify branch created (Git integration)
Type "git branch --show-current"
Sleep 500ms
Enter
Sleep 1.5s

# 8) See status change (auto state derivation)
Type "./gitta list"
Sleep 500ms
Enter
Sleep 2.5s

# 9) Filter tasks by status
Type "./gitta list --status doing"
Sleep 500ms
Enter
Sleep 2s

# 10) Show version info
Type "./gitta version"
Sleep 500ms
Enter
Sleep 2s
