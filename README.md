[![Build Status](https://drone.io/github.com/hdhog/findilo/status.png)](https://drone.io/github.com/hdhog/findilo/latest)

#Поиск ILO в сети.
По мотивам скрипта findilos http://blog.nachotech.com/?p=63

Работает как на Windows так и на *nix

пример вызова:
```bash
findilo 10.0.0.0/24
```
```bash
usage: findilo [<flags>] <network>

Flags:
  --help  Show context-sensitive help (also try --help-long and --help-man).

Args:
  <network>  Scan network, format 10.0.0.0/24
```