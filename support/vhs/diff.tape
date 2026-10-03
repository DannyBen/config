Source "support/vhs/_config.tape"
Source "support/vhs/_seed.tape"

Type "config use app.toml"
Enter
Sleep 1s
Type "config set server.port 3000 --diff --color"
Enter
Sleep 3s
Type "config get server.port"
Enter
Sleep 1s
Type "config set server.port 3000"
Enter
Sleep 1s
Type "config get server.port"
Enter
Sleep 4s
