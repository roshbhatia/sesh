Require sy

Set Shell "zsh"
Set Theme "Catppuccin Macchiato"
Set FontSize 16
Set Width 1400
Set Height 800
Set Padding 24
Set Framerate 20
Set TypingSpeed 70ms

Sleep 2s
Type "sy status provider-release"
Enter
Sleep 5s
Type 'cd "$(sy open provider-release)"'
Enter
Sleep 2s
Type "git -C changes log -3 --oneline --no-decorate"
Enter
Sleep 6s
Type "git -C ask log -3 --oneline --no-decorate"
Enter
Sleep 6s
Type "cd; sy archive provider-release"
Enter
Sleep 3s
Type "sy list --archived"
Enter
Sleep 4s
Type "sy unarchive provider-release"
Enter
Sleep 3s
Type "sy status provider-release"
Enter
Sleep 6s
