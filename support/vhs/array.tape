Source "support/vhs/_config.tape"
Source "support/vhs/_seed.tape"

Type "config use app.toml"
Enter
Sleep 1s
Type "config get plugins.enabled"
Enter
Sleep 1s
Type "config array add plugins.enabled cache"
Enter
Sleep 1s
Type "config array delete plugins.enabled metrics"
Enter
Sleep 1s
Type "cat app.toml"
Enter
Sleep 5s
