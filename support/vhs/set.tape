Source "support/vhs/_config.tape"
Source "support/vhs/_seed.tape"

Type "config use app.toml"
Enter
Sleep 1s
Type "cat app.toml"
Enter
Sleep 2s
Type "config set server.port 3000"
Enter
Sleep 1s
Type "cat app.toml"
Enter
Sleep 5s
