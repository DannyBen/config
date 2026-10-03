Source "support/vhs/_config.tape"
Source "support/vhs/_seed.tape"

Type "config use app.toml"
Enter
Sleep 1s
Type "config get server.port"
Enter
Sleep 1s
Type "config list server"
Enter
Sleep 2s
Type "exit"
Enter
Sleep 3s
