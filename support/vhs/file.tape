Source "support/vhs/_config.tape"
Source "support/vhs/_seed.tape"

Type "config get -f app.toml server.port"
Enter
Sleep 1s
Type "config set -f app.toml logging.level debug"
Enter
Sleep 1s
Type "cat app.toml"
Enter
Sleep 5s
